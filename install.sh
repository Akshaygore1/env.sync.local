#!/bin/bash
# Installation script for env-sync v2.0

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
USER_INSTALL=false
INSTALL_LEGACY=false
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
        --legacy)
            INSTALL_LEGACY=true
            shift
            ;;
        --help)
            echo "Usage: install.sh [options]"
            echo "Options:"
            echo "  --user    Install to ~/.local (no sudo required)"
            echo "  --legacy  Install legacy bash version instead of Go binary"
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

if [[ "$INSTALL_LEGACY" == "true" ]]; then
    echo -e "${BLUE}Installing env-sync (legacy bash version)...${NC}"
else
    echo -e "${BLUE}Installing env-sync v2.0 (Go binary)...${NC}"
fi

# Detect OS
OS=$(uname -s)

# Check dependencies
echo "Checking dependencies..."

MISSING_DEPS=()

if [[ "$INSTALL_LEGACY" == "true" ]]; then
    # Legacy bash version dependencies
    if ! command -v curl >/dev/null 2>&1; then
        MISSING_DEPS+=("curl")
    fi

    if ! command -v nc >/dev/null 2>&1 && ! command -v netcat >/dev/null 2>&1; then
        MISSING_DEPS+=("netcat (nc)")
    fi

    # Check for age (required for encryption support in bash version)
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
else
    # Go version dependencies (minimal - most is built-in)
    if ! command -v go >/dev/null 2>&1; then
        MISSING_DEPS+=("go (v1.24 or later)")
    fi

    case "$OS" in
        Linux)
            if ! command -v avahi-browse >/dev/null 2>&1; then
                MISSING_DEPS+=("avahi-utils (for mDNS discovery)")
            fi
            ;;
        Darwin)
            # macOS has built-in dns-sd
            ;;
    esac
fi

if [[ ${#MISSING_DEPS[@]} -gt 0 ]]; then
    echo -e "${YELLOW}Warning: Missing dependencies:${NC}"
    printf '  - %s\n' "${MISSING_DEPS[@]}"
    echo ""
    if [[ "$INSTALL_LEGACY" == "false" ]]; then
        echo "Please install them:"
        case "$OS" in
            Linux)
                echo "  Ubuntu/Debian: sudo apt-get install golang-go avahi-daemon avahi-utils"
                echo "  Fedora/RHEL:   sudo dnf install golang avahi avahi-tools"
                ;;
            Darwin)
                echo "  macOS: brew install go"
                ;;
        esac
        echo ""
        echo "Note: AGE encryption is built into the Go binary, no separate age package needed."
    else
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
    fi
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

if [[ "$INSTALL_LEGACY" == "true" ]]; then
    mkdir -p "$LIB_DIR"
fi

# Install files
echo "Installing files..."

# Stop running service if it exists (for Go version only)
SERVICE_WAS_STOPPED=false
if [[ "$INSTALL_LEGACY" == "false" ]]; then
    # Check if env-sync is already installed and try to stop the service
    if command -v env-sync >/dev/null 2>&1; then
        echo "Checking for running service..."
        # Try to stop the service gracefully
        if env-sync service stop 2>&1 | grep -q "Service stopped"; then
            SERVICE_WAS_STOPPED=true
        fi
    fi
fi

if [[ "$INSTALL_LEGACY" == "true" ]]; then
    # Install legacy bash version
    cp "$SCRIPT_DIR/legacy/bin/env-sync" "$BIN_DIR/"
    cp "$SCRIPT_DIR/legacy/bin/env-sync-discover" "$BIN_DIR/"
    cp "$SCRIPT_DIR/legacy/bin/env-sync-client" "$BIN_DIR/"
    cp "$SCRIPT_DIR/legacy/bin/env-sync-serve" "$BIN_DIR/"
    cp "$SCRIPT_DIR/legacy/bin/env-sync-key" "$BIN_DIR/"
    cp "$SCRIPT_DIR/legacy/bin/env-sync-load" "$BIN_DIR/"

    # Install library
    cp "$SCRIPT_DIR/legacy/lib/common.sh" "$LIB_DIR/"

    # Make executable
    chmod +x "$BIN_DIR"/env-sync*
else
    # Build and install Go binary
    echo "Building Go binary..."
    cd "$SCRIPT_DIR"
    make build

    if [[ ! -f "$SCRIPT_DIR/target/env-sync" ]]; then
        echo -e "${RED}✗ Build failed - binary not found${NC}"
        exit 1
    fi

    # Install Go binary
    cp "$SCRIPT_DIR/target/env-sync" "$BIN_DIR/env-sync"
    chmod +x "$BIN_DIR/env-sync"

    # Restart service if it was stopped
    if [[ "$SERVICE_WAS_STOPPED" == "true" ]]; then
        echo "Restarting service..."
        "$BIN_DIR/env-sync" service restart >/dev/null 2>&1 || {
            echo -e "${YELLOW}Note: Service was stopped but could not be restarted automatically${NC}"
            echo "Run 'env-sync serve -d' to start the service manually"
        }
    fi
fi

# Create symlinks for older macOS compatibility
if [[ "$OS" == "Darwin" && "$INSTALL_LEGACY" == "true" ]]; then
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
if [[ "$INSTALL_LEGACY" == "false" ]]; then
    echo "env-sync v2.0 (Go binary) has been installed!"
    echo ""
fi
echo "1. Initialize your secrets file:"
echo "   env-sync init --encrypted"
echo ""
echo "2. Add your secrets:"
echo "   env-sync add OPENAI_API_KEY=\"sk-...\""
echo ""
echo "3. Set up periodic sync (optional):"
echo "   env-sync cron --install"
echo ""
echo "4. On other machines, repeat steps 1-3"
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
