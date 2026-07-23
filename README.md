<p align="center">
  <img src="assets/readme/hero.svg" alt="Chorus — Anonymous Discussion Engine" width="100%">
</p>

<p align="center">
  <a href="https://github.com/joinchorus/chorus/actions/workflows/ci-cd.yml"><img src="https://github.com/joinchorus/chorus/actions/workflows/ci-cd.yml/badge.svg" alt="CI/CD Status"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="MIT License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8.svg" alt="Go Version"></a>
  <a href="https://react.dev"><img src="https://img.shields.io/badge/React-18.3-61DAFB.svg" alt="React 18"></a>
  <a href="https://git-scm.com"><img src="https://img.shields.io/badge/Storage-Native_Git-F05032.svg" alt="Native Git Engine"></a>
</p>

<h3 align="center">Identity belongs to the conversation. Not to the person.</h3>

---

## 🌐 Project Ecosystem

- 🌐 **Marketing Website:** [joinchorus.app](https://joinchorus.app)
- 📖 **Technical Documentation:** [docs.joinchorus.app](https://docs.joinchorus.app)
- 💬 **Live Application:** [chat.joinchorus.app](https://chat.joinchorus.app)

---

## 💡 What is Chorus?

Chorus is an open-source, high-performance anonymous discussion engine. It replaces account registration, user profile tracking, and social media mechanics (likes, upvotes, followers, ranking algorithms) with thread-scoped temporary identities, chronological reading, and an immutable Git persistence engine.

---

## ⚡ Why it is Different

| Traditional Social Platforms | Chorus Discussion Engine |
| :--- | :--- |
| Persistent user accounts & profiles | Zero accounts. No signups or OAuth tracking. |
| Reputation, likes, karma & follower counts | No likes, upvotes, downvotes, or leaderboards. |
| Algorithmically ranked feed loops | Pure chronological discourse. |
| Opaque proprietary databases | Immutable, audit-proof native Git commit engine. |
| Global identity tracking | Temporary identity assigned per thread (`River 🇹🇷`). |

---

## 🛠️ Architecture & Mechanism

Every thread creation and message append produces a native, cryptographic Git commit:

```text
./data/repository/
├── .git/
├── identities/
│   └── <identity_id>.json
└── threads/
    └── <thread_id>/
        ├── index.json
        └── messages/
            └── <message_id>.json
```

Commit history remains cryptographic and tamper-proof:
- `c19d788` &mdash; `thread: create thd_059f40`
- `dee2875` &mdash; `message: append msg_111`
- `32f90ec` &mdash; `message: append msg_222`

---

## 🚀 Quickstart

### 1. Run Backend Server (Go 1.22+)
```bash
go run ./cmd/server
# Server running on http://localhost:8085
```

### 2. Run Frontend Dev Server (Node.js 22+)
```bash
cd web
npm install
npm run dev
# Vite dev server running on http://localhost:3000
```

### 3. Run with Docker Compose
```bash
docker-compose up --build
```

---

## 🧪 Testing & Verification

```bash
# Run backend Go unit and integration test suite
go test -v ./...

# Build production React SPA bundle
cd web && npm run build
```

---

## 🤝 Credits & Origin

- 💡 **Concept & Product Idea:** **Mehmet Emin Dereci** ([@MEmin00](https://github.com/MEmin00))
- 🏗️ **Architecture, Lead Development & First Release:** **Barış Salih Babacan** ([@barissalihbabacan](https://github.com/barissalihbabacan))

---

## 📄 License & Community

Distributed under the MIT License. See [LICENSE](LICENSE) for details.  
Read our [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md) guidelines.
