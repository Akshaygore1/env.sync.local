#!/usr/bin/env bash

set -euo pipefail

TARGET_ARCH="${TARGET_ARCH:-$(dpkg --print-architecture)}"
NATIVE_ARCH="$(dpkg --print-architecture)"

case "$TARGET_ARCH" in
  amd64|arm64)
    ;;
  *)
    echo "Unsupported TARGET_ARCH for Linux GUI build: $TARGET_ARCH" >&2
    exit 1
    ;;
esac

if [[ "$TARGET_ARCH" != "$NATIVE_ARCH" ]]; then
  echo "Linux GUI builds require native system packages for the target architecture." >&2
  echo "Requested TARGET_ARCH=$TARGET_ARCH on a $NATIVE_ARCH runner." >&2
  echo "Use a native $TARGET_ARCH runner instead of installing conflicting multiarch WebKitGTK dev packages." >&2
  exit 1
fi

sudo apt-get update
sudo apt-get install -y \
  libgtk-3-dev \
  libwebkit2gtk-4.1-dev \
  pkg-config
