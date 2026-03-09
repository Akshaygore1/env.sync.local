#!/usr/bin/env bash

set -euo pipefail

sudo dpkg --add-architecture arm64

if [[ -f /etc/apt/sources.list.d/ubuntu.sources ]]; then
  sudo sed -i '/^Types:/a Architectures: amd64' /etc/apt/sources.list.d/ubuntu.sources
fi

cat <<'SOURCES' | sudo tee /etc/apt/sources.list.d/ubuntu-arm64.sources > /dev/null
Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports/
Suites: noble noble-updates
Components: main restricted universe multiverse
Architectures: arm64
SOURCES

sudo apt-get update
sudo apt-get install -y \
  libgtk-3-dev libwebkit2gtk-4.1-dev pkg-config \
  gcc-aarch64-linux-gnu \
  libgtk-3-dev:arm64 libwebkit2gtk-4.1-dev:arm64
