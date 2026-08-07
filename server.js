const http = require('http');
const { spawn, execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const PORT = parseInt(process.env.PORT || '8080', 10);
const BACKEND_PORT = 8085;
const DIST_DIR = path.join(__dirname, 'web', 'dist');

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
const serverBinary = path.join(__dirname, 'server');

function startGoBackend() {
  const env = { ...process.env, PORT: String(BACKEND_PORT) };

  if (fs.existsSync(serverBinary)) {
    console.log(`Starting compiled Go server binary (${serverBinary}) on port ${BACKEND_PORT}...`);
    try {
      goProcess = spawn(serverBinary, [], { env, stdio: 'inherit' });
    } catch (err) {
      console.error(`Failed spawning ${serverBinary}:`, err.message);
    }
  } else if (commandExists('go')) {
    try {
      console.log('Building Go binary from ./cmd/server...');
      execSync('go build -o server ./cmd/server', { stdio: 'inherit' });
      if (fs.existsSync(serverBinary)) {
        console.log('Starting compiled Go server binary...');
        goProcess = spawn(serverBinary, [], { env, stdio: 'inherit' });
      }
    } catch (err) {
      console.warn('go build failed, attempting go run ./cmd/server...', err.message);
      try {
        goProcess = spawn('go', ['run', './cmd/server'], { env, stdio: 'inherit' });
      } catch (spawnErr) {
        console.error('Failed spawning go run:', spawnErr.message);
      }
    }
  } else {
    console.warn('Go runtime toolchain ("go" binary) is not installed in this environment.');
    console.warn('Node process will serve static SPA assets on port ' + PORT + '. For full Go backend functionality, deploy via Docker container.');
  }

  if (goProcess) {
    goProcess.on('error', (err) => {
      console.error(`Go process error (${err.message}). Operating in static SPA fallback mode.`);
      goProcess = null;
    });

    goProcess.on('exit', (code) => {
      if (code !== null && code !== 0) {
        console.warn(`Go backend process exited with code ${code}.`);
      }
    });
  }
}

startGoBackend();

// 2. HTTP Reverse Proxy & SPA Static File Server
const proxyServer = http.createServer((req, res) => {
  if (!goProcess) {
    // Backend not running, serve static SPA assets directly
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
    serveStaticSPA(req, res);
  });

  req.pipe(proxyReq, { end: true });
});

function serveStaticSPA(req, res) {
  const urlPath = req.url ? req.url.split('?')[0] : '/';
  let safePath = path.normalize(urlPath).replace(/^(\.\.[\/\\])+/, '');
  let filePath = path.join(DIST_DIR, safePath);

  fs.stat(filePath, (statErr, stats) => {
    if (!statErr && stats.isFile()) {
      res.writeHead(200);
      fs.createReadStream(filePath).pipe(res);
    } else {
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
