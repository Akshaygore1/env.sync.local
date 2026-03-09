#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"
SRC_DIR="$ROOT_DIR/src"

mkdir -p "$DIST_DIR"

cd "$SRC_DIR"

GOOS=windows GOARCH=amd64 go build -o "$DIST_DIR/env-sync-windows-amd64.exe" -ldflags "-s -w" ./cmd/env-sync
GOOS=windows GOARCH=arm64 go build -o "$DIST_DIR/env-sync-windows-arm64.exe" -ldflags "-s -w" ./cmd/env-sync
GOOS=darwin GOARCH=amd64 go build -o "$DIST_DIR/env-sync-macos-amd64" -ldflags "-s -w" ./cmd/env-sync
GOOS=darwin GOARCH=arm64 go build -o "$DIST_DIR/env-sync-macos-arm64" -ldflags "-s -w" ./cmd/env-sync
GOOS=linux GOARCH=arm64 go build -o "$DIST_DIR/env-sync-linux-arm64" -ldflags "-s -w" ./cmd/env-sync
GOOS=linux GOARCH=amd64 go build -o "$DIST_DIR/env-sync-linux-amd64" -ldflags "-s -w" ./cmd/env-sync
