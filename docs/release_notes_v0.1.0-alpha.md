# Chorus v0.1.0-alpha Release Notes

We are pleased to announce the first public alpha release of **Chorus** (`v0.1.0-alpha`), an open-source anonymous discussion platform designed for calm, authentic communication without accounts, profiles, followers, or engagement algorithms.

---

## ⚠️ Status & Stability Warning

> [!WARNING]
> **Status: Experimental (`v0.1.0-alpha`)**
> - The HTTP REST API is under active development and considered **unstable**.
> - The underlying Git repository storage schema (`boards/general/threads/thd_xxxxx/`) may change before `v1.0.0`.
> - This release is intended for developer evaluation, self-hosting experimentation, and community feedback.

---

## 🌟 Key Features in v0.1.0-alpha

1. **Git-Backed Persistence Engine (`internal/gitstore`)**:
   - Stores threads, message logs, and identities inside a local Git repository.
   - Append-only NDJSON log format (`messages.ndjson`).
   - Every write operation creates a standardized local Git commit (`thread: create`, `message: append`).
   - Thread summary indexing (`boards/general/index.json`) for high-throughput reads.
   - Automatic startup recovery (`RecoverAndVerify`) rebuilding runtime indexes directly from Git data.

2. **Go Backend Core**:
   - Zero framework dependencies built on Go 1.22+ `net/http` router.
   - Separation of Concerns & Single Source of Truth (SSOT).
   - Injected domain dependencies and central validation (`internal/domain/validation.go`).
   - Structured JSON logging (`log/slog`) and panic recovery.
   - Pagination support (`limit`, `offset`) on list endpoints.

3. **Minimalist Web Application (`web/`)**:
   - Built with React 18, TypeScript, Vite, and TanStack Query.
   - Minimalist technical aesthetic (system typography, light-mode, monospace IDs).
   - No profiles, no accounts, no avatars, no algorithms.
   - Backend is the Single Source of Truth for all IDs, timestamps, and country tags.

4. **Specifications & Operations**:
   - Full OpenAPI 3.0 specification (`openapi.yaml`).
   - Immutable Event Log & Auditability Architecture specification (`docs/event_architecture.md`).
   - Performance benchmarks and load test report (`PERFORMANCE.md`).
   - Multi-stage `Dockerfile` and `docker-compose.yml` with persistent volumes.
   - GitHub Actions CI workflow (`.github/workflows/ci.yml`).

---

## 🚧 Known Limitations

1. **Synchronous Git CLI Overhead**:
   - Each write operation invokes external `git add` and `git commit` processes (`exec.Command`), introducing a ~50ms baseline latency floor per write.
2. **Mutex Contention Under Extreme Load**:
   - Storage writes serialize behind a `sync.RWMutex`, which caps write throughput to ~100 req/sec under heavy concurrent load (see `PERFORMANCE.md`).
3. **Local IP Geolocation Fallback**:
   - Country flag resolution relies on client request IP headers (`X-Forwarded-For` / `RemoteAddr`) and defaults to local/loopback resolution during development.

---

## 🗺️ Roadmap

- **v0.2.0 (Performance & Asynchronous Commits)**:
  - Background batching write-queue to decouple HTTP response latency from Git commit execution.
  - In-memory index mutation to eliminate full `index.json` disk rewrites on appends.
- **v0.3.0 (GeoIP & Reader Translation)**:
  - MaxMind / IP2Location GeoIP integration for accurate country resolution.
  - On-demand reader translation integration.
- **v1.0.0 (Stable Public Release)**:
  - Guaranteed backward compatibility for HTTP API endpoints and Git storage schema.

---

## 📋 Breaking Change Policy

Prior to `v1.0.0`, minor releases (`v0.x.0`) may introduce breaking changes to API endpoint payloads or storage folder hierarchies. All breaking changes will be documented in `CHANGELOG.md` with migration guides provided for self-hosted instances.

---

## 📦 Getting Started

### Using Docker Compose
```bash
git clone https://github.com/barissalihbabacan/Chorus.git
cd Chorus
docker-compose up -d
```
Visit `http://localhost:8080` in your browser.

### Running Locally
```bash
# Start backend
go run ./cmd/server

# Start frontend
cd web && npm run dev
```
