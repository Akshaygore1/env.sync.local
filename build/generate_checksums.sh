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

  artifacts+=("$name")
done

if ((${#artifacts[@]} == 0)); then
  echo "No release artifacts found in $DIST_DIR"
  exit 0
fi

cd "$DIST_DIR"
sha256sum -- "${artifacts[@]}" > checksums.txt
cat checksums.txt
