#!/bin/bash

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

echo "=== Verifying release artifact and installer expectations ==="

assert_contains() {
    local file="$1"
    local pattern="$2"
    local message="$3"

    if grep -Fq "$pattern" "$file"; then
        echo "✓ $message"
    else
        echo "✗ $message"
        echo "  Expected to find: $pattern"
        exit 1
    fi
}

assert_not_contains() {
    local file="$1"
    local pattern="$2"
    local message="$3"

    if grep -Fq "$pattern" "$file"; then
        echo "✗ $message"
        echo "  Unexpectedly found: $pattern"
        exit 1
    else
        echo "✓ $message"
    fi
}

LINUX_GUI_BUILD="$REPO_ROOT/build/build_gui_linux.sh"
MACOS_GUI_BUILD="$REPO_ROOT/build/build_gui_macos.sh"
INSTALL_SCRIPT="$REPO_ROOT/install.sh"
GUI_LOGO_SOURCE="$REPO_ROOT/docs/logo.png"
GUI_APP_ICON="$REPO_ROOT/src/gui/build/appicon.png"

assert_not_contains "$LINUX_GUI_BUILD" 'tar -C "$payload_dir" -czf "$DIST_DIR/env-sync-gui-linux-${arch}.tar.gz" .' "Linux GUI release build no longer emits tar.gz payloads"
assert_contains "$LINUX_GUI_BUILD" 'dpkg-deb --build "$deb_root" "$DIST_DIR/env-sync-gui-linux-${arch}.deb"' "Linux GUI release build emits .deb packages"
assert_contains "$MACOS_GUI_BUILD" '"$DIST_DIR/env-sync-gui-macos-${arch}"' "macOS GUI release build emits per-arch raw binaries"
assert_contains "$MACOS_GUI_BUILD" '"$DIST_DIR/env-sync-gui-macos-${arch}.dmg"' "macOS GUI release build emits per-arch DMGs"
assert_contains "$INSTALL_SCRIPT" 'download_artifact "env-sync-gui-linux-${PLATFORM_ARCH}.deb"' "Remote installer downloads Linux GUI .deb assets"
assert_not_contains "$INSTALL_SCRIPT" 'download_artifact "env-sync-gui-linux-${PLATFORM_ARCH}.tar.gz"' "Remote installer no longer downloads Linux GUI tarballs"
assert_contains "$INSTALL_SCRIPT" 'GUI_INSTALL_DIR="$HOME/Applications"' "macOS user GUI installs go to ~/Applications"
assert_contains "$INSTALL_SCRIPT" 'GUI_INSTALL_DIR="/Applications"' "macOS system GUI installs go to /Applications"
assert_contains "$INSTALL_SCRIPT" 'xattr -dr com.apple.quarantine "$GUI_INSTALL_TARGET" >/dev/null 2>&1 || true' "macOS GUI installer clears quarantine metadata from installed app bundles"
assert_contains "$INSTALL_SCRIPT" 'GUI_INSTALL_DIR="$HOME/.local/lib/env-sync"' "Linux user GUI installs go to ~/.local/lib/env-sync"
assert_contains "$INSTALL_SCRIPT" 'GUI_INSTALL_DIR="/opt/env-sync"' "Linux system GUI installs go to /opt/env-sync"

if cmp -s "$GUI_LOGO_SOURCE" "$GUI_APP_ICON"; then
    echo "✓ GUI app icon asset matches docs/logo.png"
else
    echo "✗ GUI app icon asset does not match docs/logo.png"
    exit 1
fi

echo "=== Release artifact and installer checks passed ==="
