#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"

mkdir -p "$DIST_DIR"

shopt -s nullglob
artifacts=()
for path in "$DIST_DIR"/*; do
  [[ -f "$path" ]] || continue

  name="$(basename "$path")"
  if [[ "$name" == "checksums.txt" ]]; then
    continue
  fi

  artifacts+=("$path")
done

has_artifacts=false
if ((${#artifacts[@]} > 0)); then
  bash "$ROOT_DIR/build/generate_checksums.sh"
  has_artifacts=true
else
  echo "No release artifacts were downloaded; skipping checksum generation and release publishing."
fi

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  echo "has_artifacts=$has_artifacts" >> "$GITHUB_OUTPUT"
fi
