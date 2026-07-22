# Chorus Release Process & Policy

This document outlines the release process, versioning conventions, and breaking change policy for **Chorus**.

---

## 1. Versioning Scheme

Chorus adheres to [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) (`MAJOR.MINOR.PATCH`).

- **Alpha Releases (`v0.1.x-alpha`)**: Initial experimental releases. Focus on core architectural validation, storage format refinement, and security/performance baseline establishment.
- **Beta Releases (`v0.2.x-beta`)**: Feature-complete pre-releases. Focused on stability, performance optimizations, and community feedback.
- **Stable Releases (`v1.0.0+`)**: Production-ready releases with guaranteed public API stability and backward compatibility for Git storage schemas.

---

## 2. Pre-v1.0 Breaking Change Policy

> [!WARNING]
> Prior to `v1.0.0`, breaking changes to the HTTP REST API endpoints or underlying Git repository storage schema (`boards/general/threads/...`) may occur between minor versions.

### Policy Guidelines for `v0.x`:
1. **Notice & Documentation**: All breaking changes, parameter deprecations, or schema migrations will be prominently documented in `CHANGELOG.md` and Release Notes.
2. **Migration Paths**: Whenever the on-disk storage format changes, an automated migration utility or recovery upgrade path will be provided.

---

## 3. Pre-Release Verification Checklist

Before tagging any public release (e.g. `git tag -a v0.1.0-alpha -m "Release v0.1.0-alpha"`), maintainers must verify:

- [ ] All Go tests pass cleanly with race detection enabled (`go test -v -race ./...`).
- [ ] Go code passes static analysis (`go vet ./...`).
- [ ] React frontend builds cleanly (`cd web && npm run build`).
- [ ] `openapi.yaml` matches the implemented REST API endpoints and response payloads.
- [ ] `CHANGELOG.md` is updated with all notable additions, changes, and deprecations.
- [ ] Multi-stage Docker image builds without errors (`docker build -t chorus:v0.1.0-alpha .`).
- [ ] `audit_readme.py` passes without broken links or missing media.

---

## 4. Release Execution Workflow

1. Update `CHANGELOG.md` with release highlights and date.
2. Commit release metadata changes (`git commit -m "release: prepare v0.1.0-alpha"`).
3. Create annotated Git tag:
   ```bash
   git tag -a v0.1.0-alpha -m "Chorus v0.1.0-alpha"
   ```
4. Push tag to remote:
   ```bash
   git push origin v0.1.0-alpha
   ```
5. Draft GitHub Release using the contents of `docs/release_notes_v0.1.0-alpha.md`.
