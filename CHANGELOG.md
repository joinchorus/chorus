# Changelog

All notable changes to the Chorus platform will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.1.0-alpha] - 2026-08-07

### Added
- **Runtime Compatibility**: Improved runtime compatibility for legacy Node-based build environments by introducing a lightweight launcher (`server.js`) that starts the Go server when required.
- **Admin Session Login & Cookies**: Added `POST /api/v0.1/moderation/login` and `POST /api/v0.1/moderation/logout` endpoints setting `HttpOnly`, `SameSite=Lax`, `Path=/` admin session cookies (`chorus_admin_session`).
- **IP Token Bucket Rate Limiting**: Added thread-safe `IPRateLimiter` (`internal/http/middleware/ratelimit.go`) guarding write operations (`POST /threads`, `POST /messages`, `POST /reports`) with HTTP 429 status codes and `Retry-After` headers.
- **Request Correlation Observability**: Added `RequestID` middleware (`internal/http/middleware/request_id.go`) generating or preserving `X-Request-ID` headers, attaching request correlation IDs to HTTP contexts, response headers, `log/slog` logs (`request_id=req_<hex>`), panic traces, and error JSON payloads.
- **Graceful Shutdown Lifecycle**: Implemented 10-second graceful shutdown context drain (`cmd/server/main.go`) capturing `SIGINT` and `SIGTERM` signals without terminating active Git write operations.
- **Unit & Integration Test Suite**: Added test coverage for authorization (`auth_test.go`), rate limiting (`ratelimit_test.go`), security headers (`security_test.go`), request correlation IDs (`request_id_test.go`), and graceful shutdown (`shutdown_test.go`).

### Changed
- **Admin Moderation UI**: Updated `AdminModeration.tsx` to render an interactive authentication overlay using credentials `same-origin`.

### Deprecated
- *None in this release.*

### Removed
- **Obsolete Security Header**: Removed `X-XSS-Protection` header (superseded by modern CSP; prone to auditor vulnerabilities in legacy browsers).
- **Client-Side Fallback Mocks**: Removed `MOCK_THREADS` and `MOCK_MODERATION_ITEMS` fallbacks from `web/src/lib/api.ts`.
- **LocalStorage Secret Storage**: Removed admin token persistence in browser `localStorage`.

### Fixed
- *None in this release.*

### Security
- **Moderation Auth Guard**: Protected `/api/v0.1/moderation/*` and `/api/v1/moderation/*` endpoints with `RequireAdminAuth` middleware using `crypto/subtle.ConstantTimeCompare` to mitigate side-channel timing attacks.
- **OWASP Security Headers**: Enforced `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, `Permissions-Policy`, `COOP` (`same-origin`), `CORP` (`same-origin`), and `HSTS` (`max-age=31536000`).
- **Modern Content Security Policy (CSP)**: Updated CSP directives compatible with Markdown rendering, KaTeX math typesetting, Web Workers, and Monaco Editor.
- **Payload Limits**: Enforced 1MB request body limit in `httputil.DecodeJSON` via `http.MaxBytesReader`.
