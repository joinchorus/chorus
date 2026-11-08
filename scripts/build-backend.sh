#!/usr/bin/env bash
# Builds the Go backend into ./server (linux/amd64, static) from the checked-out
# source. Firebase App Hosting runs it through `npm run build`; CI runs it too.
#
# The build image has no Go toolchain, so one is downloaded and verified when
# `go` is not on PATH. Locally (or in CI with setup-go) the installed Go is used.
# Any failure stops the build, so a broken backend never reaches a rollout.
set -euo pipefail

# Keep in step with the `go` line in go.mod and the Dockerfile base image.
GO_VERSION="1.25.14"
GO_SHA256="a21ae5633a269bcd7e90cf767e48225633795e99d831742cbf3397064fee7712" # go1.25.14.linux-amd64.tar.gz

cd "$(dirname "$0")/.."

# Toolchain and caches live outside the app directory so they are not deployed.
WORK="${TMPDIR:-/tmp}/chorus-build"
mkdir -p "$WORK"

if ! command -v go >/dev/null 2>&1; then
  if [ "$(uname -s)-$(uname -m)" != "Linux-x86_64" ]; then
    echo "error: no Go on PATH and automatic download only supports linux/amd64" >&2
    exit 1
  fi
  archive="$WORK/go${GO_VERSION}.tar.gz"
  echo "Downloading Go ${GO_VERSION}..."
  curl -fsSL --retry 3 -o "$archive" "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
  echo "${GO_SHA256}  ${archive}" | sha256sum -c -
  rm -rf "$WORK/go"
  tar -C "$WORK" -xzf "$archive"
  export PATH="$WORK/go/bin:$PATH"
fi

# Stamp the commit so /healthz shows which source a running instance was built from.
commit="${COMMIT_SHA:-}"
if [ -z "$commit" ]; then
  commit="$(git rev-parse HEAD 2>/dev/null || echo unknown)"
fi

export GOCACHE="$WORK/gocache" GOMODCACHE="$WORK/gomod" GOFLAGS="-mod=readonly"
echo "Building backend $(go version | cut -d' ' -f3) commit ${commit}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
  -ldflags "-s -w -X main.commit=${commit}" \
  -o server ./cmd/server

test -s server
echo "Built ./server ($(wc -c < server | tr -d ' ') bytes)"
