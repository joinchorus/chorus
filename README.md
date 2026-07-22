<div align="center">

![Chorus Hero Header](assets/readme/hero.svg)

<br/>

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Architecture: Clean](https://img.shields.io/badge/Architecture-SoC%20%7C%20SSOT-emerald?style=flat-square)](#architecture)
[![Build Status](https://img.shields.io/badge/Tests-Passing-brightgreen?style=flat-square)](#development)

---

[Key Features](#features) • [Why Chorus?](#why-chorus) • [Philosophy](#philosophy) • [Architecture](#architecture) • [Roadmap](#roadmap) • [Getting Started](#development)

</div>

<br />

> [!NOTE]  
> **Core Concept**: In Chorus, participants do not create account profiles. Every discussion thread automatically generates a temporary, cryptographically isolated identity that exists *only* within that thread.

> [!WARNING]
> **Project Status**: `v0.1.0-alpha` (Experimental). Chorus is under active development. The HTTP API is unstable and the underlying Git repository storage schema may evolve before `v1.0.0`.

<br />


## Why Chorus?

Most online communication platforms force users to build permanent profiles, accumulate follower counts, and compete for algorithmic engagement metrics. Over time, these mechanics shift incentives away from thoughtful discussion toward reputation management, echo chambers, and self-censorship.

Furthermore, traditional global forums either isolate users into language silos or force aggressive auto-translations that strip away context and cultural nuance.

```text
Traditional Platforms                       Chorus
┌──────────────────────────────┐            ┌──────────────────────────────┐
│ • Permanent Profiles         │            │ • Ephemeral Thread Identifier│
│ • Followers & Karma          │    vs      │ • Zero Account Creation      │
│ • Algorithmic Feeds          │            │ • Strict Chronological Feed  │
│ • Forced Auto-Translation    │            │ • On-Demand Reader Translation│
└──────────────────────────────┘            └──────────────────────────────┘
```

Chorus solves these structural flaws by decoupling discussion from persistent user identity and providing on-demand translation controls.

---

## Philosophy

Chorus adheres strictly to ten architectural and product principles:

| Principle | Description |
| :--- | :--- |
| **Anonymous by Default** | Zero registration, email collection, or credential storage. |
| **No Usernames** | Ephemeral identity tokens generated automatically per thread session. |
| **No Followers** | No social graph, follower lists, or networking mechanisms. |
| **No Profiles** | User activity history is never tracked, aggregated, or displayed. |
| **No Algorithms** | Discussions are rendered strictly in chronological order. |
| **Substance over Identity** | Focus is placed entirely on message content rather than author reputation. |
| **Optional Translation** | Original text is preserved; readers trigger translation on demand. |
| **Privacy First** | Zero telemetry, zero persistent user tracking, minimal metadata. |
| **Open Source** | Fully audit-able codebase, standard library backend, transparent design. |
| **Responsible by Design** | Built-in participant reporting and thread-level moderation controls. |

---

## Features

- **Thread-Isolated Identities**: Submitting a post generates a temporary `usr_<hex>` identifier unique to that specific thread. The identity cannot be linked across multiple threads.
- **On-Demand Translation**: Readers control when and if a post is translated into their local language, keeping original text primary.
- **Optional Country Flags**: Coarse IP-based geographic flag indicators provide global context without revealing precise locations.
- **Append-Only Git Persistence**: Designed to use a Git repository as a transparent, immutable persistence layer.
- **Zero Global State Backend**: Go service designed around Separation of Concerns (SoC), Single Source of Truth (SSOT), and explicit dependency injection.

---

## Architecture

![Chorus Architecture Diagram](assets/readme/architecture.svg)

Chorus is structured into strict architectural layers where every component depends only on the layer below it:

1. **HTTP Transport Layer**: Route declarations (`net/http`), request parsing (`httputil`), `slog` logging, and panic recovery middleware.
2. **Service Layer**: Business validation, ephemeral identity generation (`idgen`), thread management, and message append logic.
3. **Core Domain**: Single Source of Truth (SSOT) entities (`Identity`, `Thread`, `Message`) and domain errors (`ErrNotFound`, `ErrValidation`).
4. **Persistence Layer**: Thread-safe in-memory repository (`sync.RWMutex`) designed to swap seamlessly with an append-only Git storage engine.

---

## Project Structure

```text
chorus/
├── cmd/
│   └── server/
│       └── main.go           # Application entrypoint & dependency injection root
├── internal/
│   ├── config/               # Environment configuration loader
│   ├── domain/               # Core domain errors and sentinel types
│   ├── http/                 # Transport layer
│   │   ├── handler/          # HTTP handlers & consumer service interfaces
│   │   ├── httputil/         # Request decoding & response rendering helpers
│   │   ├── middleware/       # slog logging & panic recovery middleware
│   │   └── router.go         # Endpoint route declarations (Go 1.22+ net/http)
│   ├── idgen/                # Cryptographic ID generator package
│   ├── identity/             # Identity domain logic & service implementation
│   ├── repository/           # Persistence layer
│   │   └── memory/           # Thread-safe in-memory store
│   └── thread/               # Thread & message domain logic & service
├── go.mod                    # Go module definition
└── README.md
```

---

## Roadmap

- [x] **Core HTTP Server**: Standard library router with structured `slog` logging and panic recovery middleware.
- [x] **Identity Service**: Ephemeral cryptographic identity generation (`usr_<hex>`).
- [x] **Thread & Message Service**: Creation, retrieval, and listing operations for threads and messages.
- [ ] **Git-Backed Persistence**: Immutable append-only Git repository storage engine.
- [ ] **On-Demand Translation**: Translation integration with reader-side triggering.
- [ ] **Geographic Flags**: Coarse IP geolocation for optional country flag tags.
- [ ] **Moderation Tools**: Participant abuse reporting and community moderation workflows.
- [ ] **Realtime Updates**: Server-Sent Events (SSE) / WebSocket live message streaming.

---

## Development

### Prerequisites

- **Go**: 1.22 or higher

### Compiling

To build the standalone server binary:

```bash
go build -o server ./cmd/server
```

### Running Locally

To start the server:

```bash
go run ./cmd/server
```

The server listens on port `8080` by default. Configuration can be customized via environment variables:

```bash
PORT=9090 ENV=production go run ./cmd/server
```

### Testing & Verification

Run all unit and integration tests with the Go race detector enabled:

```bash
go test -v -race ./...
```

Run static analysis:

```bash
go vet ./...
```

---

## Contributing

Contributions are welcome. Please ensure all pull requests strictly maintain layer boundaries, pass unit tests with the race detector enabled (`go test -v -race ./...`), and include clear test coverage.

For major architectural proposals, please open an issue first to discuss the design before submitting code.

---

## License

Chorus is released under the [MIT License](LICENSE).
