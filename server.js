const express = require('express');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 8080;

app.use(express.json());

// Healthcheck endpoint for Firebase App Hosting
app.get('/healthz', (req, res) => {
  res.status(200).json({ status: 'ok' });
});

// Serve compiled static SPA assets
const distPath = path.join(__dirname, 'web', 'dist');
app.use(express.static(distPath));

// SPA Client-side routing fallback for all pages
app.get('*', (req, res) => {
  res.sendFile(path.join(distPath, 'index.html'));
});

app.listen(PORT, '0.0.0.0', () => {
  console.log(`Chorus App Hosting server running on port ${PORT}`);
});
