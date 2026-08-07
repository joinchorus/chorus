const http = require('http');
const { spawn, execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const PORT = parseInt(process.env.PORT || '8080', 10);
const BACKEND_PORT = 8085;
const DIST_DIR = path.join(__dirname, 'web', 'dist');

const MIME_TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.mjs': 'application/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.ttf': 'font/ttf',
  '.otf': 'font/otf',
  '.map': 'application/json; charset=utf-8',
};

// Ensure frontend assets exist
if (!fs.existsSync(DIST_DIR)) {
  try {
    console.log('Building React SPA bundle...');
    execSync('cd web && npm ci && npm run build', { stdio: 'inherit' });
  } catch (err) {
    console.error('Frontend build failed:', err.message);
  }
}

// Check if command exists in PATH
function commandExists(cmd) {
  try {
    execSync(`command -v ${cmd}`, { stdio: 'ignore' });
    return true;
  } catch {
    return false;
  }
}

// 1. Locate or compile Go binary
let goProcess = null;
let lastGoError = 'Go process initialization in progress.';
const serverBinarySrc = path.join(__dirname, 'server');
const serverBinaryTmp = path.join('/tmp', 'chorus_server_backend');

function startGoBackend() {
  const env = { ...process.env, PORT: String(BACKEND_PORT), DATA_DIR: process.env.DATA_DIR || '/tmp/chorus-repository' };
  let binaryToRun = null;

  if (fs.existsSync(serverBinarySrc)) {
    try {
      fs.copyFileSync(serverBinarySrc, serverBinaryTmp);
      fs.chmodSync(serverBinaryTmp, 0o755);
      binaryToRun = serverBinaryTmp;
      console.log(`Successfully prepared executable Go binary at ${serverBinaryTmp}`);
    } catch (copyErr) {
      console.warn('Failed copying binary to /tmp, falling back to source binary path:', copyErr.message);
      try {
        fs.chmodSync(serverBinarySrc, 0o755);
      } catch {}
      binaryToRun = serverBinarySrc;
    }
  }

  if (binaryToRun) {
    console.log(`Starting compiled Go server binary (${binaryToRun}) on port ${BACKEND_PORT}...`);
    try {
      goProcess = spawn(binaryToRun, [], { env, stdio: 'inherit' });
      lastGoError = `Go backend process spawned successfully (PID: ${goProcess.pid}).`;
    } catch (err) {
      lastGoError = `Failed spawning Go binary (${binaryToRun}): ${err.message}`;
      console.error(lastGoError);
    }
  } else if (commandExists('go')) {
    try {
      console.log('Building Go binary from ./cmd/server...');
      execSync('go build -o server ./cmd/server', { stdio: 'inherit' });
      if (fs.existsSync(serverBinarySrc)) {
        console.log('Starting compiled Go server binary...');
        goProcess = spawn(serverBinarySrc, [], { env, stdio: 'inherit' });
        lastGoError = `Go backend compiled and spawned successfully.`;
      }
    } catch (err) {
      console.warn('go build failed, attempting go run ./cmd/server...', err.message);
      try {
        goProcess = spawn('go', ['run', './cmd/server'], { env, stdio: 'inherit' });
        lastGoError = `Go backend running via 'go run'.`;
      } catch (spawnErr) {
        lastGoError = `Failed spawning 'go run': ${spawnErr.message}`;
        console.error(lastGoError);
      }
    }
  } else {
    lastGoError = 'Go binary or runtime toolchain not found in container environment.';
    console.warn(lastGoError);
  }

  if (goProcess) {
    goProcess.on('error', (err) => {
      lastGoError = `Go process fatal error: ${err.message}`;
      console.error(lastGoError);
      goProcess = null;
    });

    goProcess.on('exit', (code, signal) => {
      if (code !== null && code !== 0) {
        lastGoError = `Go backend process exited with code ${code} (signal: ${signal}).`;
        console.warn(lastGoError);
      }
    });
  }
}

startGoBackend();

// 2. HTTP Reverse Proxy & SPA Static File Server
const proxyServer = http.createServer((req, res) => {
  const urlPath = req.url ? req.url.split('?')[0] : '/';

  // API endpoints should never fallback to index.html
  if (!goProcess && (urlPath.startsWith('/api/') || urlPath === '/healthz')) {
    respondAPIUnavailable(res);
    return;
  }

  if (!goProcess) {
    serveStaticSPA(req, res);
    return;
  }

  const options = {
    hostname: '127.0.0.1',
    port: BACKEND_PORT,
    path: req.url,
    method: req.method,
    headers: req.headers,
  };

  const proxyReq = http.request(options, (proxyRes) => {
    res.writeHead(proxyRes.statusCode, proxyRes.headers);
    proxyRes.pipe(res, { end: true });
  });

  proxyReq.on('error', () => {
    if (urlPath.startsWith('/api/') || urlPath === '/healthz') {
      respondAPIUnavailable(res);
    } else {
      serveStaticSPA(req, res);
    }
  });

  req.pipe(proxyReq, { end: true });
});

function respondAPIUnavailable(res) {
  res.writeHead(503, { 'Content-Type': 'application/json' });
  res.end(
    JSON.stringify({
      error: {
        code: 'service_unavailable',
        message: 'Backend API is starting up or unavailable. Please retry in a moment.',
        detail: lastGoError,
      },
    })
  );
}

function serveStaticSPA(req, res) {
  const urlPath = req.url ? req.url.split('?')[0] : '/';
  let safePath = path.normalize(urlPath).replace(/^(\.\.[\/\\])+/, '');
  let filePath = path.join(DIST_DIR, safePath);
  const ext = path.extname(filePath).toLowerCase();

  // Guard against serving HTML for API calls
  if (urlPath.startsWith('/api/') || urlPath === '/healthz') {
    respondAPIUnavailable(res);
    return;
  }

  fs.stat(filePath, (statErr, stats) => {
    if (!statErr && stats.isFile()) {
      const contentType = MIME_TYPES[ext] || 'application/octet-stream';
      res.writeHead(200, { 'Content-Type': contentType });
      fs.createReadStream(filePath).pipe(res);
    } else {
      // Do NOT serve index.html for missing static asset files (.css, .js, .png, etc.)
      if (ext && ext !== '.html') {
        res.writeHead(404, { 'Content-Type': 'text/plain' });
        res.end(`Asset not found: ${safePath}`);
        return;
      }

      // SPA Fallback for client-side HTML page navigation routes
      const indexPath = path.join(DIST_DIR, 'index.html');
      if (fs.existsSync(indexPath)) {
        res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
        fs.createReadStream(indexPath).pipe(res);
      } else {
        res.writeHead(503, { 'Content-Type': 'text/plain' });
        res.end('Chorus App starting up...');
      }
    }
  });
}

proxyServer.listen(PORT, '0.0.0.0', () => {
  console.log(`Chorus App Supervisor listening on 0.0.0.0:${PORT}`);
});
