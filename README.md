<p align="center">
  <img src="assets/readme/hero.svg" alt="Chorus: an example thread where participants named River, Ash and Echo each map to a Git commit such as 'message: append msg_111'" width="100%">
</p>

<p align="center">
  <a href="https://chat.joinchorus.app"><b>Live demo</b></a> ·
  <a href="https://joinchorus.app">Website</a> ·
  <a href="https://docs.joinchorus.app">Docs</a>
</p>

<p align="center">
  <a href="https://github.com/joinchorus/chorus/actions/workflows/ci-cd.yml"><img src="https://github.com/joinchorus/chorus/actions/workflows/ci-cd.yml/badge.svg" alt="CI/CD status"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT license"></a>
  <img src="https://img.shields.io/badge/Go-1.22-00ADD8.svg" alt="Go 1.22">
  <img src="https://img.shields.io/badge/React-19.2-61DAFB.svg" alt="React 19.2">
  <img src="https://img.shields.io/badge/status-v0.1.0--alpha-orange.svg" alt="Early development, v0.1.0-alpha">
</p>

> **Early development (v0.1.0-alpha).** Chorus works end to end, but expect rough edges and breaking changes. See the [changelog](CHANGELOG.md).

## What is Chorus?

Chorus is an open-source anonymous discussion engine. There are no accounts, likes, followers or ranking. Each thread gives every participant a temporary name (like *River* or *Ash*) that exists only in that thread, and the thread is read in chronological order.

Every thread and message is stored as JSON in a local Git repository and committed, so the full history is plain, append-only Git history you can inspect with ordinary Git tools.

## How it differs

| Typical social platform | Chorus |
| :--- | :--- |
| Persistent accounts and profiles | No accounts, no sign-up |
| Likes, karma, follower counts | None of these exist |
| Ranked feed | Chronological threads |
| One identity everywhere | A new temporary name per thread, with an optional country flag |
| Opaque database | A Git repository on disk |

## How it works

Each write is one commit in the data repository (`DATA_DIR`):

```text
data/repository/
├── .git/
├── identities/<identity_id>.json
└── threads/<thread_id>/
    ├── index.json
    └── messages/<message_id>.json
```

```text
thread: create thd_059f40 in <board>
message: append msg_111
```

Moderation removals are hidden when messages are read, but the original content remains in Git history. That is append-only history, not a guarantee that it cannot be altered by someone with access to the data directory.

## Features

- Thread-scoped identities, boards, and chronological reading.
- Opt-in country flag.
- Reporting, plus an admin moderation queue protected by an admin token.
- Per-IP rate limiting on writes.
- Optional message translation through the Google Translate API. It needs `GOOGLE_TRANSLATE_API_KEY`, and the text of the message you translate is sent to Google.
- REST API under `/api/v1` (also served as `/api/v0.1`), plus `GET /healthz`.

## Quickstart

Requirements: Go 1.22, `git` on your `PATH`, Node.js for the web client.

**1. Start the server** (standard library only, no extra Go dependencies):

```bash
go run ./cmd/server
# listens on http://localhost:8085
```

**2. Start the web client** (React 19 + Vite 8, proxies `/api` to the server):

```bash
cd web
npm install
npm run dev
# http://localhost:3000
```

**Or use Docker:**

```bash
docker-compose up --build
# http://localhost:8080, data kept in the chorus-data volume
```

### Configuration

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `PORT` | `8085` | HTTP port (`8080` in `docker-compose.yml`) |
| `DATA_DIR` | `./data/repository` | Location of the Git data repository |
| `CHORUS_ADMIN_TOKEN` | public placeholder | Admin token for moderation. Read in this order: `CHORUS_LICENSE_KEY`, `CHORUS_ADMIN_TOKEN`, `MODERATOR_API_KEY`. If none is set, the server falls back to a placeholder that is visible in this repository, so **always set one** in any deployment others can reach. |
| `GOOGLE_TRANSLATE_API_KEY` | unset | API key for message translation |

## Tests

```bash
go test ./...
cd web && npm run build
```

## Contributing and security

- Contributing: see the [organization contributing guide](https://github.com/joinchorus/.github/blob/main/CONTRIBUTING.md).
- Security: read [SECURITY.md](SECURITY.md) and report vulnerabilities privately through [GitHub security advisories](https://github.com/joinchorus/chorus/security/advisories/new).
- Conduct: [Code of Conduct](CODE_OF_CONDUCT.md).

## Credits

- Concept and product idea: **Mehmet Emin Dereci** ([@MEmin00](https://github.com/MEmin00))
- Architecture, lead development and first release: **Barış Salih Babacan** ([@barissalihbabacan](https://github.com/barissalihbabacan))

## License

[MIT](LICENSE)
