#!/bin/bash
# Generate shared SSH keys for test containers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_DIR="$SCRIPT_DIR/../docker/ssh-keys"

echo "Generating SSH keys for Docker test containers..."

# Create SSH keys directory
mkdir -p "$KEYS_DIR"

# Generate a single key pair for all containers (simplifies testing)
if [ ! -f "$KEYS_DIR/id_ed25519" ]; then
    echo "  Creating Ed25519 key pair..."
    ssh-keygen -t ed25519 -f "$KEYS_DIR/id_ed25519" -N "" -C "env-sync-test@docker"
    echo "  ✓ Created $KEYS_DIR/id_ed25519"
else
    echo "  ✓ Key pair already exists: $KEYS_DIR/id_ed25519"
fi

# Create authorized_keys file (all containers trust the same key)
if [ ! -f "$KEYS_DIR/authorized_keys" ]; then
    echo "  Creating authorized_keys..."
    cp "$KEYS_DIR/id_ed25519.pub" "$KEYS_DIR/authorized_keys"
    echo "  ✓ Created authorized_keys"
else
    echo "  ✓ authorized_keys already exists"
fi

# Set proper permissions (in case they were lost)
chmod 700 "$KEYS_DIR"
chmod 600 "$KEYS_DIR/id_ed25519"
chmod 644 "$KEYS_DIR/id_ed25519.pub"
chmod 644 "$KEYS_DIR/authorized_keys"

echo ""
echo "SSH keys ready in: $KEYS_DIR"
echo "  - id_ed25519 (private key)"
echo "  - id_ed25519.pub (public key)"
echo "  - authorized_keys (for SSH access between containers)"
