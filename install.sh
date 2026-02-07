#!/bin/bash
# Installation script for env-sync

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
USER_INSTALL=false
INSTALL_PREFIX="/usr/local"
BIN_DIR="$INSTALL_PREFIX/bin"
LIB_DIR="$INSTALL_PREFIX/lib/env-sync"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --user)
            USER_INSTALL=true
            INSTALL_PREFIX="$HOME/.local"
            BIN_DIR="$INSTALL_PREFIX/bin"
            LIB_DIR="$INSTALL_PREFIX/lib/env-sync"
            shift
            ;;
        --help)
            echo "Usage: install.sh [options]"
            echo "Options:"
            echo "  --user    Install to ~/.local (no sudo required)"
            echo "  --help    Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${BLUE}Installing env-sync...${NC}"

# Detect OS
OS=$(uname -s)

# Check dependencies
echo "Checking dependencies..."

MISSING_DEPS=()

if ! command -v curl >/dev/null 2>&1; then
    MISSING_DEPS+=("curl")
fi

if ! command -v nc >/dev/null 2>&1 && ! command -v netcat >/dev/null 2>&1; then
    MISSING_DEPS+=("netcat (nc)")
fi

# Check for age (required for encryption support)
if ! command -v age >/dev/null 2>&1; then
    MISSING_DEPS+=("age")
fi

if ! command -v age-keygen >/dev/null 2>&1; then
    MISSING_DEPS+=("age-keygen")
fi

case "$OS" in
    Linux)
        if ! command -v avahi-browse >/dev/null 2>&1; then
            MISSING_DEPS+=("avahi-utils")
        fi
        ;;
    Darwin)
        # macOS has built-in dns-sd
        ;;
esac

if [[ ${#MISSING_DEPS[@]} -gt 0 ]]; then
    echo -e "${YELLOW}Warning: Missing dependencies:${NC}"
    printf '  - %s\n' "${MISSING_DEPS[@]}"
    echo ""
    echo "Please install them:"
    case "$OS" in
        Linux)
            echo "  Ubuntu/Debian: sudo apt-get install avahi-daemon avahi-utils curl netcat-openbsd age"
            echo "  Fedora/RHEL:   sudo dnf install avahi avahi-tools curl nmap-ncat age"
            echo ""
            echo "  To install age manually:"
            echo "    curl -fsSL https://github.com/FiloSottile/age/releases/latest/download/age-v1.2.0-linux-amd64.tar.gz | tar -xz -C /usr/local/bin --strip-components=1"
            ;;
        Darwin)
            echo "  macOS: brew install age"
            echo ""
            echo "  To install age manually:"
            echo "    curl -fsSL https://github.com/FiloSottile/age/releases/latest/download/age-v1.2.0-darwin-amd64.tar.gz | tar -xz -C /usr/local/bin --strip-components=1"
            ;;
    esac
    echo ""
    read -p "Continue anyway? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Create directories
echo "Creating directories..."
mkdir -p "$BIN_DIR"
mkdir -p "$LIB_DIR"

# Install files
echo "Installing files..."

# Install binaries
cp "$SCRIPT_DIR/bin/env-sync" "$BIN_DIR/"
cp "$SCRIPT_DIR/bin/env-sync-discover" "$BIN_DIR/"
cp "$SCRIPT_DIR/bin/env-sync-client" "$BIN_DIR/"
cp "$SCRIPT_DIR/bin/env-sync-serve" "$BIN_DIR/"

# Install library
cp "$SCRIPT_DIR/lib/common.sh" "$LIB_DIR/"

# Make executable
chmod +x "$BIN_DIR"/env-sync*

# Create symlinks for older macOS compatibility
if [[ "$OS" == "Darwin" ]]; then
    # macOS uses BSD sed which has different syntax
    # Update scripts to use gsed if available
    if command -v gsed >/dev/null 2>&1; then
        for script in "$BIN_DIR"/env-sync*; do
            sed -i.bak 's/sed -i /gsed -i /g' "$script" 2>/dev/null || true
            rm -f "$script.bak"
        done
    fi
fi

echo -e "${GREEN}Installation complete!${NC}"
echo ""

# Check if bin directory is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo -e "${YELLOW}Warning: $BIN_DIR is not in your PATH${NC}"
    echo "Add the following to your ~/.bashrc or ~/.zshrc:"
    echo ""
    echo "  export PATH=\"$BIN_DIR:\$PATH\""
    echo ""
fi

# Post-install instructions
echo "Next steps:"
echo ""
echo "1. Initialize your secrets file:"
echo "   env-sync init"
echo ""
echo "2. Edit ~/.secrets.env to add your secrets"
echo ""
echo "3. Start the server:"
echo "   env-sync serve -d"
echo ""
echo "4. Set up periodic sync (optional):"
echo "   env-sync cron --install"
echo ""
echo "5. On other machines, repeat steps 1-4"
echo ""
echo "The machines will automatically discover each other!"
echo ""

# Verify installation
echo "Verifying installation..."
if command -v env-sync >/dev/null 2>&1; then
    echo -e "${GREEN}✓ env-sync installed successfully${NC}"
    env-sync --help | head -20
else
    echo -e "${RED}✗ Installation verification failed${NC}"
    echo "Please ensure $BIN_DIR is in your PATH"
    exit 1
fi
