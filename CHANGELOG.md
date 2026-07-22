# Changelog

All notable changes to **Chorus** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.1.0-alpha] - 2026-07-22

### Status
> [!WARNING]
> **Experimental Release**. The HTTP API is unstable and the underlying Git repository storage format may change before `v1.0.0`.

### Added
- **Git-Backed Persistence Engine (`internal/gitstore`)**:
  - Store threads, messages, and identities inside a local Git repository.
  - Append-only NDJSON format for message logs (`messages.ndjson`).
  - Automatic local Git commits per write (`thread: create`, `message: append`, `identity: create`).
  - High-performance thread summary indexing (`boards/general/index.json`).
  - Automatic startup repository recovery (`RecoverAndVerify`) rebuilding indexes directly from Git data.
- **Go Backend Core**:
  - Zero framework dependencies using Go 1.22+ `net/http` router.
  - Dependency Injection with concrete struct returns and consumer-defined interfaces.
  - Centralized domain validation (`internal/domain/validation.go`).
  - Structured JSON logging with `log/slog` and panic recovery middleware.
  - Pagination support (`limit`, `offset`) on list endpoints.
- **React Web Application (`web/`)**:
  - First MVP application UI using React 18, TypeScript, Vite, and TanStack Query.
  - Minimalist design aesthetic (system typography, light-mode, monospace IDs).
  - Clean separation of UI states (loading, error, empty).
  - 100% integration with backend as Single Source of Truth (no client-side ID/timestamp generation).
- **System Documentation & Specification**:
  - Full OpenAPI 3.0 specification (`openapi.yaml`).
  - Immutable Event Log & Auditability Architecture design (`docs/event_architecture.md`).
  - Performance Micro-benchmarks & Stress Testing report (`PERFORMANCE.md`).
- **Containerization & CI**:
  - Multi-stage `Dockerfile` and `docker-compose.yml` with persistent storage volumes.
  - GitHub Actions CI workflow (`.github/workflows/ci.yml`).

### Fixed
- Fixed deadlock during startup index reconstruction by separating internal locked commit execution (`addAndCommitLocked`).
- Solved CORS and host binding configurations for local development in Vite.
