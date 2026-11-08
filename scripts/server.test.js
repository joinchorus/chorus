// Process-level tests for server.js, run with `node --test scripts/`.
// The Go backend is replaced by a small shell script so no build is needed.
'use strict';

const { test } = require('node:test');
const assert = require('node:assert/strict');
const { spawn } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

function sandbox(backendScript) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'serverjs-'));
  fs.copyFileSync(path.join(__dirname, '..', 'server.js'), path.join(dir, 'server.js'));
  fs.mkdirSync(path.join(dir, 'web', 'dist'), { recursive: true });
  fs.writeFileSync(path.join(dir, 'web', 'dist', 'index.html'), '<html></html>');
  fs.writeFileSync(path.join(dir, 'server'), backendScript, { mode: 0o755 });
  return dir;
}

function start(dir, port, extraEnv = {}) {
  return spawn('node', ['server.js'], {
    cwd: dir,
    env: { ...process.env, ENV: 'production', PORT: String(port), DATA_DIR: path.join(dir, 'data'), ...extraEnv },
    stdio: 'ignore',
  });
}

function exited(child, ms) {
  return new Promise((resolve) => {
    const timer = setTimeout(() => {
      child.kill('SIGKILL');
      resolve({ timedOut: true });
    }, ms);
    child.on('exit', (code, signal) => {
      clearTimeout(timer);
      resolve({ code, signal, timedOut: false });
    });
  });
}

test('production: the process exits when the Go backend dies, so Cloud Run replaces the instance', async () => {
  const dir = sandbox('#!/bin/sh\nexit 1\n');
  const result = await exited(start(dir, 18091), 10000);
  assert.equal(result.timedOut, false, 'server.js kept running with a dead backend');
  assert.equal(result.code, 1);
});

test('production: SIGTERM is forwarded to the Go backend, which gets to finish its shutdown work', async () => {
  const marker = path.join(os.tmpdir(), `serverjs-marker-${process.pid}`);
  fs.rmSync(marker, { force: true });
  const dir = sandbox('#!/bin/sh\ntrap \'echo ok > "$MARKER"; exit 0\' TERM\nwhile true; do sleep 0.2; done\n');
  const child = start(dir, 18092, { MARKER: marker });

  await new Promise((resolve) => setTimeout(resolve, 2000));
  child.kill('SIGTERM');
  const result = await exited(child, 12000);

  assert.equal(result.timedOut, false);
  assert.equal(result.code, 0);
  assert.equal(fs.existsSync(marker), true, 'the backend never received SIGTERM');
  fs.rmSync(marker, { force: true });
});

test('production: a missing backend binary stops the process instead of serving 503 forever', async () => {
  const dir = sandbox('#!/bin/sh\nexit 0\n');
  fs.rmSync(path.join(dir, 'server'));
  const result = await exited(start(dir, 18093), 10000);
  assert.equal(result.timedOut, false);
  assert.equal(result.code, 1);
});
