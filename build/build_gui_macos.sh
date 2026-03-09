#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"
GUI_DIR="$ROOT_DIR/src/gui"

mkdir -p "$DIST_DIR"

cd "$GUI_DIR"

for arch in amd64 arm64; do
  rm -rf build/bin
  PATH="$HOME/go/bin:$PATH" \
  CGO_ENABLED=1 \
  CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
  wails build -clean -platform "darwin/${arch}"

  stage_dir="$(mktemp -d)"
  cp -R build/bin/env-sync.app "$stage_dir/env-sync.app"
  hdiutil create \
    -volname "env-sync" \
    -srcfolder "$stage_dir" \
    -ov \
    -format UDZO \
    "$DIST_DIR/env-sync-gui-macos-${arch}.dmg"
  rm -rf "$stage_dir"
done
