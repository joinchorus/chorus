# Chorus — Anonymous Discussion Engine

[![CI/CD](https://github.com/chorus-project/chorus/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/chorus-project/chorus/actions/workflows/ci-cd.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)
[![React](https://img.shields.io/badge/React-18.3-61DAFB.svg)](https://react.dev)

> **Identity belongs to the conversation. Not to the person.**

Chorus is an open-source, high-performance anonymous discussion platform. It eliminates traditional account registration, profile tracking, and social media mechanics (likes, followers, upvotes, algorithms) in favor of thread-scoped temporary identities, chronological reading, and native Git persistence.

---

## 🔗 Project Ecosystem

- 🌐 **Marketing Website:** [joinchorus.app](https://joinchorus.app)
- 📖 **Technical Documentation:** [docs.joinchorus.app](https://docs.joinchorus.app)
- 💬 **Live Application:** [chat.joinchorus.app](https://chat.joinchorus.app)

---

## 🚀 Architecture Highlights

- **Go Backend (`/cmd/server`, `/internal`):** Clean architecture with domain isolation (`internal/thread`, `internal/identity`, `internal/gitstore`, `internal/geoip`).
- **React SPA (`/web`):** Modern TypeScript, Vite, TanStack Query frontend.
- **Git Persistence Engine (`internal/gitstore`):** Audit-proof, immutable commit log where every thread and message generates native Git commit history.
- **Subdomain Host Router:** Subdomain dispatcher routing `chat.joinchorus.app` traffic seamlessly.

---

## 🛠️ Local Quickstart

### Prerequisites
- **Go 1.22+**
- **Node.js 22+**
- **Git**

### 1. Run Backend Server
```bash
go run ./cmd/server
# Server listens on http://localhost:8085
```

### 2. Run Frontend Development Server
```bash
cd web
npm install
npm run dev
# Vite dev server runs on http://localhost:3000
```

### 3. Run with Docker Compose
```bash
docker-compose up --build
```

---

## 🧪 Testing & Verification

```bash
# Run backend Go test suite
go test -v ./...

# Build React production bundle
cd web && npm run build
```

---

## 📦 Deployment

The application is deployed independently to **Firebase App Hosting** targeting `chat.joinchorus.app` via GitHub Actions (`.github/workflows/ci-cd.yml`).

---

## 📄 License & Community

Distributed under the MIT License. See [LICENSE](LICENSE) for details.
See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md) for contribution & security guidelines.
