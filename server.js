const http = require('http');
const { spawn, execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const PORT = parseInt(process.env.PORT || '8080', 10);
const BACKEND_PORT = 8085;
const DIST_DIR = path.join(__dirname, 'web', 'dist');

// Ensure frontend assets are built if missing
if (!fs.existsSync(DIST_DIR)) {
  try {
    console.log('Building React SPA bundle...');
    execSync('cd web && npm ci && npm run build', { stdio: 'inherit' });
  } catch (err) {
    console.error('Frontend build failed:', err);
  }
}

// 1. Locate or compile Go binary
let goProcess = null;
const serverBinary = path.join(__dirname, 'server');

function startGoBackend() {
  const env = { ...process.env, PORT: String(BACKEND_PORT) };

  if (fs.existsSync(serverBinary)) {
    console.log(`Starting compiled Go server binary (${serverBinary}) on port ${BACKEND_PORT}...`);
    goProcess = spawn(serverBinary, [], { env, stdio: 'inherit' });
  } else {
    try {
      console.log('Building Go binary from ./cmd/server...');
      execSync('go build -o server ./cmd/server', { stdio: 'inherit' });
      console.log('Starting compiled Go server binary...');
      goProcess = spawn(serverBinary, [], { env, stdio: 'inherit' });
    } catch (err) {
      console.log('go build unavailable, running via go run ./cmd/server...');
      goProcess = spawn('go', ['run', './cmd/server'], { env, stdio: 'inherit' });
    }
  }

  if (goProcess) {
    goProcess.on('exit', (code) => {
      console.warn(`Go backend process exited with code ${code}. Restarting in 2s...`);
      setTimeout(startGoBackend, 2000);
    });
  }
}

startGoBackend();

// 2. HTTP Reverse Proxy to Go backend
const proxyServer = http.createServer((req, res) => {
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

  proxyReq.on('error', (err) => {
    // If backend is still warming up, fallback to static asset serving temporarily
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
          res.end('Go backend starting up...');
        }
      }
    });
  });

  req.pipe(proxyReq, { end: true });
});

proxyServer.listen(PORT, '0.0.0.0', () => {
  console.log(`Chorus App Supervisor listening on 0.0.0.0:${PORT} (proxying to Go backend on port ${BACKEND_PORT})`);
});
