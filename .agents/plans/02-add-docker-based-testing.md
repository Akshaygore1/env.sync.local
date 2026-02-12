# env-sync V2.0 - Docker-Based Testing Infrastructure

## Overview

Implement a comprehensive Docker-based testing infrastructure to verify the distributed secrets sync functionality across multiple containers. This provides isolated, reproducible testing of peer discovery, encryption, and sync workflows.

## Goals

- **Reproducible Testing**: Spin up fresh test environments on demand
- **Multi-Machine Simulation**: Test 3+ machine scenarios on a single host
- **Network Isolation**: Test mDNS discovery within Docker networks
- **CI/CD Ready**: Suitable for automated testing pipelines
- **Zero Host Dependencies**: Tests run entirely in containers

## Architecture

### Test Network Topology

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Network: env-sync-test              │
│                        (bridge mode)                          │
│                                                               │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐      │
│  │   alpha      │   │    beta      │   │   gamma      │      │
│  │              │   │              │   │              │      │
│  │ ~/.secrets.env│   │ ~/.secrets.env│   │ ~/.secrets.env│     │
│  │ AGE keypair   │   │ AGE keypair   │   │ AGE keypair   │     │
│  │ SSH server    │   │ SSH server    │   │ SSH server    │     │
│  │ avahi-daemon  │   │ avahi-daemon  │   │ avahi-daemon  │     │
│  └──────────────┘   └──────────────┘   └──────────────┘      │
│          │                 │                 │               │
│          └─────────────────┴─────────────────┘               │
│                     mDNS Discovery                            │
│                     SCP/SSH Sync                              │
└─────────────────────────────────────────────────────────────┘
```

### Container Configuration

Each container:
- **Base**: Ubuntu 22.04 LTS
- **Hostname**: alpha.local, beta.local, gamma.local
- **Network**: Shared bridge network with mDNS support
- **Services**: SSH daemon, avahi-daemon for mDNS
- **Tools**: env-sync, age, avahi-utils, openssh-server

## Directory Structure

```
env-sync/
├── tests/
│   ├── docker/
│   │   ├── Dockerfile                    # Test container image
│   │   ├── docker-compose.yml            # Multi-container orchestration
│   │   └── entrypoint.sh                 # Container startup script
│   ├── scripts/
│   │   ├── test-runner.sh                # Main test orchestrator
│   │   ├── test-setup.sh                 # Initialize test environment
│   │   ├── test-cleanup.sh               # Clean up after tests
│   │   ├── test-scenario-1.sh            # Basic sync test
│   │   ├── test-scenario-2.sh            # Encryption test
│   │   ├── test-scenario-3.sh            # Add new machine test
│   │   └── test-scenario-4.sh            # Conflict resolution test
│   └── utils/
│       ├── wait-for-sync.sh              # Wait for sync completion
│       ├── verify-secrets.sh             # Verify secrets match across nodes
│       └── generate-ssh-keys.sh          # Generate shared SSH keys
└── V2.md                                 # This document
```

## Implementation Plan

### Phase 1: Docker Infrastructure (Week 1)

#### 1.1 Test Container Dockerfile

```dockerfile
FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    bash \
    curl \
    openssh-server \
    avahi-daemon \
    avahi-utils \
    libnss-mdns \
    age \
    jq \
    netcat \
    && rm -rf /var/lib/apt/lists/*

# Configure SSH
RUN mkdir -p /var/run/sshd
RUN echo 'root:envsync' | chpasswd
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config

# Configure Avahi for mDNS
RUN sed -i 's/#enable-dbus=yes/enable-dbus=no/' /etc/avahi/avahi-daemon.conf
RUN sed -i 's/hosts:.*files dns/hosts: files mdns_minimal [NOTFOUND=return] dns mdns/' /etc/nsswitch.conf

# Copy env-sync binaries (built from source)
COPY bin/ /usr/local/bin/
COPY lib/ /usr/local/lib/env-sync/

# Make scripts executable
RUN chmod +x /usr/local/bin/env-sync*

# Create env-sync directories
RUN mkdir -p /root/.config/env-sync/keys/known_hosts
RUN mkdir -p /root/.config/env-sync/logs

# Copy SSH keys (injected at runtime)
RUN mkdir -p /root/.ssh

# Entrypoint script
COPY tests/docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["bash"]
```

#### 1.2 Docker Compose Configuration

```yaml
version: '3.8'

networks:
  env-sync-test:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/24

services:
  alpha:
    build:
      context: ../..
      dockerfile: tests/docker/Dockerfile
    hostname: alpha.local
    container_name: env-sync-alpha
    networks:
      env-sync-test:
        ipv4_address: 172.20.0.2
    volumes:
      - ./ssh-keys:/root/.ssh:ro
      - alpha-data:/root/.config/env-sync
    environment:
      - ENV_SYNC_DEBUG=1
      - CONTAINER_NAME=alpha
    cap_add:
      - NET_ADMIN  # Required for mDNS
    ports:
      - "5739:5739"  # env-sync HTTP port (optional)

  beta:
    build:
      context: ../..
      dockerfile: tests/docker/Dockerfile
    hostname: beta.local
    container_name: env-sync-beta
    networks:
      env-sync-test:
        ipv4_address: 172.20.0.3
    volumes:
      - ./ssh-keys:/root/.ssh:ro
      - beta-data:/root/.config/env-sync
    environment:
      - ENV_SYNC_DEBUG=1
      - CONTAINER_NAME=beta
    cap_add:
      - NET_ADMIN

  gamma:
    build:
      context: ../..
      dockerfile: tests/docker/Dockerfile
    hostname: gamma.local
    container_name: env-sync-gamma
    networks:
      env-sync-test:
        ipv4_address: 172.20.0.4
    volumes:
      - ./ssh-keys:/root/.ssh:ro
      - gamma-data:/root/.config/env-sync
    environment:
      - ENV_SYNC_DEBUG=1
      - CONTAINER_NAME=gamma
    cap_add:
      - NET_ADMIN

volumes:
  alpha-data:
  beta-data:
  gamma-data:
```

#### 1.3 Container Entrypoint Script

```bash
#!/bin/bash
set -e

# Start SSH daemon
/usr/sbin/sshd

# Start Avahi for mDNS
dbus-daemon --system --fork 2>/dev/null || true
avahi-daemon --daemonize

# Wait for services to be ready
sleep 2

# Initialize SSH keys if not present
if [ ! -f /root/.ssh/id_ed25519 ]; then
    ssh-keygen -t ed25519 -f /root/.ssh/id_ed25519 -N ""
fi

# Trust all other containers' SSH keys
for host in alpha.local beta.local gamma.local; do
    ssh-keyscan -H $host >> /root/.ssh/known_hosts 2>/dev/null || true
done

# If this is first run, initialize env-sync
if [ ! -f /root/.secrets.env ]; then
    # Will be initialized by test scripts
    echo "Container ready for testing"
fi

# Keep container running
exec "$@"
```

### Phase 2: Test Utilities (Week 1)

#### 2.1 SSH Key Generation Script

```bash
#!/bin/bash
# tests/utils/generate-ssh-keys.sh
# Generate shared SSH keys for test containers

KEYS_DIR="$(dirname "$0")/../docker/ssh-keys"
mkdir -p "$KEYS_DIR"

# Generate single key pair for all containers (simplifies testing)
if [ ! -f "$KEYS_DIR/id_ed25519" ]; then
    ssh-keygen -t ed25519 -f "$KEYS_DIR/id_ed25519" -N "" -C "env-sync-test"
fi

# Copy public key as authorized_keys
cp "$KEYS_DIR/id_ed25519.pub" "$KEYS_DIR/authorized_keys"

echo "SSH keys generated in $KEYS_DIR"
```

#### 2.2 Sync Wait Utility

```bash
#!/bin/bash
# tests/utils/wait-for-sync.sh
# Wait for sync to complete between containers

CONTAINER=$1
TIMEOUT=${2:-30}

echo "Waiting for sync in $CONTAINER..."
docker exec $CONTAINER timeout $TIMEOUT bash -c '
    while true; do
        if [ -f /root/.config/env-sync/logs/env-sync.log ]; then
            if grep -q "sync completed\|sync failed" /root/.config/env-sync/logs/env-sync.log 2>/dev/null; then
                exit 0
            fi
        fi
        sleep 1
    done
'
```

#### 2.3 Secrets Verification Utility

```bash
#!/bin/bash
# tests/utils/verify-secrets.sh
# Verify secrets match across all containers

KEY=$1
EXPECTED_VALUE=$2

for container in env-sync-alpha env-sync-beta env-sync-gamma; do
    echo "Checking $container..."
    
    # Get secret value from container
    ACTUAL=$(docker exec $container env-sync show "$KEY" 2>/dev/null || echo "")
    
    if [ "$ACTUAL" != "$EXPECTED_VALUE" ]; then
        echo "ERROR: $container has wrong value for $KEY"
        echo "  Expected: $EXPECTED_VALUE"
        echo "  Actual:   $ACTUAL"
        exit 1
    fi
    
    echo "  ✓ $KEY = $ACTUAL"
done

echo "All containers have matching secrets!"
```

### Phase 3: Test Scenarios (Week 2)

#### 3.1 Test Scenario 1: Basic Sync (Unencrypted)

**Goal**: Verify secrets sync from alpha to beta and gamma without encryption

```bash
#!/bin/bash
# tests/scripts/test-scenario-1.sh

set -e

echo "=== Test Scenario 1: Basic Sync (Unencrypted) ==="

# Step 1: Initialize alpha with a secret
echo "Initializing alpha with plaintext secrets..."
docker exec env-sync-alpha env-sync init

# Add a test secret
docker exec env-sync-alpha env-sync add TEST_KEY="alpha-value-123"

# Step 2: Initialize beta and gamma
echo "Initializing beta and gamma..."
docker exec env-sync-beta env-sync init
docker exec env-sync-gamma env-sync init

# Step 3: Trigger sync from beta to get secrets from alpha
echo "Triggering sync on beta..."
docker exec env-sync-beta env-sync --force

# Step 4: Trigger sync from gamma
echo "Triggering sync on gamma..."
docker exec env-sync-gamma env-sync --force

# Step 5: Verify all containers have the secret
echo "Verifying secrets synced..."
./tests/utils/verify-secrets.sh "TEST_KEY" "alpha-value-123"

echo "✓ Test Scenario 1 PASSED"
```

#### 3.2 Test Scenario 2: Encrypted Sync

**Goal**: Verify AGE encryption works across all containers

```bash
#!/bin/bash
# tests/scripts/test-scenario-2.sh

set -e

echo "=== Test Scenario 2: Encrypted Sync ==="

# Step 1: Clear all containers and initialize with encryption
echo "Initializing alpha with encrypted secrets..."
docker exec env-sync-alpha rm -f /root/.secrets.env
docker exec env-sync-alpha env-sync init --encrypted
docker exec env-sync-alpha env-sync add ENCRYPTED_SECRET="secret-value-456"

# Step 2: Get alpha's pubkey and add to beta/gamma
ALPHA_PUBKEY=$(docker exec env-sync-alpha env-sync key show)
echo "Alpha pubkey: $ALPHA_PUBKEY"

# Initialize beta and gamma with encryption
docker exec env-sync-beta rm -f /root/.secrets.env
docker exec env-sync-beta env-sync init --encrypted
docker exec env-sync-gamma rm -f /root/.secrets.env
docker exec env-sync-gamma env-sync init --encrypted

# Step 3: Exchange public keys between all containers
echo "Exchanging public keys..."
BETA_PUBKEY=$(docker exec env-sync-beta env-sync key show)
GAMMA_PUBKEY=$(docker exec env-sync-gamma env-sync key show)

# Import keys on each container
docker exec env-sync-alpha env-sync key import "$BETA_PUBKEY" beta.local
docker exec env-sync-alpha env-sync key import "$GAMMA_PUBKEY" gamma.local
docker exec env-sync-beta env-sync key import "$ALPHA_PUBKEY" alpha.local
docker exec env-sync-beta env-sync key import "$GAMMA_PUBKEY" gamma.local
docker exec env-sync-gamma env-sync key import "$ALPHA_PUBKEY" alpha.local
docker exec env-sync-gamma env-sync key import "$BETA_PUBKEY" beta.local

# Step 4: Trigger sync on all containers
echo "Triggering sync..."
docker exec env-sync-alpha env-sync --force
docker exec env-sync-beta env-sync --force
docker exec env-sync-gamma env-sync --force

# Step 5: Verify encrypted secrets synced
echo "Verifying encrypted secrets synced..."
./tests/utils/verify-secrets.sh "ENCRYPTED_SECRET" "secret-value-456"

echo "✓ Test Scenario 2 PASSED"
```

#### 3.3 Test Scenario 3: Adding New Secret Propagates to All

**Goal**: When a new secret is added to one container, it appears on all others

```bash
#!/bin/bash
# tests/scripts/test-scenario-3.sh

set -e

echo "=== Test Scenario 3: New Secret Propagation ==="

# Prerequisites: Run test-scenario-2 first to have encrypted setup

# Step 1: Add new secret to beta
echo "Adding new secret to beta..."
docker exec env-sync-beta env-sync add NEW_SECRET="new-value-789"

# Step 2: Trigger sync on alpha and gamma (they should get the new secret)
echo "Triggering sync on alpha..."
docker exec env-sync-alpha env-sync --force

echo "Triggering sync on gamma..."
docker exec env-sync-gamma env-sync --force

# Step 3: Verify all containers have the new secret
echo "Verifying new secret propagated..."
./tests/utils/verify-secrets.sh "NEW_SECRET" "new-value-789"

# Step 4: Verify old secret still exists
./tests/utils/verify-secrets.sh "ENCRYPTED_SECRET" "secret-value-456"

echo "✓ Test Scenario 3 PASSED"
```

#### 3.4 Test Scenario 4: Adding Fourth Container (Delta)

**Goal**: Adding a new machine requires zero changes to existing machines

```bash
#!/bin/bash
# tests/scripts/test-scenario-4.sh

set -e

echo "=== Test Scenario 4: Adding Fourth Container (Delta) ==="

# Step 1: Start delta container
docker run -d \
    --name env-sync-delta \
    --hostname delta.local \
    --network env-sync-test \
    --ip 172.20.0.5 \
    -v "$(pwd)/tests/docker/ssh-keys:/root/.ssh:ro" \
    -e ENV_SYNC_DEBUG=1 \
    env-sync:test

# Wait for delta to be ready
sleep 5

# Step 2: Initialize delta with encryption
docker exec env-sync-delta env-sync init --encrypted

# Step 3: Collect pubkeys from existing machines
echo "Collecting public keys from alpha, beta, gamma..."
docker exec env-sync-delta env-sync discover --collect-keys

# Step 4: Request access by triggering re-encryption on alpha
echo "Requesting access via alpha..."
docker exec env-sync-delta env-sync key request-access --trigger alpha.local

# Step 5: Sync delta to get encrypted secrets
echo "Syncing delta..."
docker exec env-sync-delta env-sync --force

# Step 6: Verify delta can decrypt and see all secrets
echo "Verifying delta has all secrets..."
DELTA_VALUE=$(docker exec env-sync-delta env-sync show ENCRYPTED_SECRET 2>/dev/null || echo "")
if [ "$DELTA_VALUE" != "secret-value-456" ]; then
    echo "ERROR: Delta cannot access encrypted secrets"
    exit 1
fi

# Step 7: Add secret on delta and verify it syncs back
docker exec env-sync-delta env-sync add DELTA_SECRET="delta-specific"
docker exec env-sync-alpha env-sync --force
./tests/utils/verify-secrets.sh "DELTA_SECRET" "delta-specific"

# Cleanup
docker stop env-sync-delta && docker rm env-sync-delta

echo "✓ Test Scenario 4 PASSED"
```

### Phase 4: Test Orchestration (Week 2)

#### 4.1 Test Setup Script

```bash
#!/bin/bash
# tests/scripts/test-setup.sh

set -e

echo "=== Setting up test environment ==="

# Generate SSH keys
echo "Generating SSH keys..."
./tests/utils/generate-ssh-keys.sh

# Build Docker image
echo "Building Docker image..."
docker-compose -f tests/docker/docker-compose.yml build

# Start containers
echo "Starting containers..."
docker-compose -f tests/docker/docker-compose.yml up -d

# Wait for containers to be ready
echo "Waiting for containers to be ready..."
sleep 10

# Verify all containers are running
for container in env-sync-alpha env-sync-beta env-sync-gamma; do
    if ! docker ps | grep -q $container; then
        echo "ERROR: $container is not running"
        exit 1
    fi
    echo "  ✓ $container is running"
done

# Verify mDNS is working
echo "Verifying mDNS discovery..."
docker exec env-sync-alpha timeout 10 avahi-browse -a -t | grep -E "alpha|beta|gamma" || true

echo "✓ Test environment ready"
```

#### 4.2 Test Cleanup Script

```bash
#!/bin/bash
# tests/scripts/test-cleanup.sh

echo "=== Cleaning up test environment ==="

# Stop and remove containers
docker-compose -f tests/docker/docker-compose.yml down -v

# Remove delta if it exists
docker stop env-sync-delta 2>/dev/null || true
docker rm env-sync-delta 2>/dev/null || true

# Optionally remove test image
docker rmi env-sync:test 2>/dev/null || true

echo "✓ Cleanup complete"
```

#### 4.3 Main Test Runner

```bash
#!/bin/bash
# tests/scripts/test-runner.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# Parse arguments
RUN_CLEANUP=1
SCENARIO=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --no-cleanup)
            RUN_CLEANUP=0
            shift
            ;;
        --scenario)
            SCENARIO="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --no-cleanup    Don't cleanup containers after tests"
            echo "  --scenario N    Run only specific scenario (1-4)"
            echo "  --help          Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Cleanup on exit
cleanup() {
    if [ $RUN_CLEANUP -eq 1 ]; then
        echo ""
        echo "Running cleanup..."
        ./scripts/test-cleanup.sh
    fi
}
trap cleanup EXIT

# Setup
echo "========================================"
echo "env-sync Docker Test Suite"
echo "========================================"
./scripts/test-setup.sh

# Run tests
FAILED=0

if [ -z "$SCENARIO" ] || [ "$SCENARIO" = "1" ]; then
    echo ""
    ./scripts/test-scenario-1.sh || FAILED=1
fi

if [ -z "$SCENARIO" ] || [ "$SCENARIO" = "2" ]; then
    echo ""
    ./scripts/test-scenario-2.sh || FAILED=1
fi

if [ -z "$SCENARIO" ] || [ "$SCENARIO" = "3" ]; then
    echo ""
    ./scripts/test-scenario-3.sh || FAILED=1
fi

if [ -z "$SCENARIO" ] || [ "$SCENARIO" = "4" ]; then
    echo ""
    ./scripts/test-scenario-4.sh || FAILED=1
fi

# Summary
echo ""
echo "========================================"
if [ $FAILED -eq 0 ]; then
    echo "✓ All tests PASSED"
    echo "========================================"
    exit 0
else
    echo "✗ Some tests FAILED"
    echo "========================================"
    exit 1
fi
```

### Phase 5: CI/CD Integration (Week 3)

#### 5.1 GitHub Actions Workflow

```yaml
# .github/workflows/docker-tests.yml
name: Docker Integration Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  docker-tests:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Install dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y age jq
    
    - name: Run Docker tests
      run: |
        cd tests
        ./scripts/test-runner.sh
    
    - name: Upload logs on failure
      if: failure()
      uses: actions/upload-artifact@v3
      with:
        name: test-logs
        path: |
          tests/logs/
```

#### 5.2 Makefile Integration

```makefile
# Add to main Makefile

test-docker:
	cd tests && ./scripts/test-runner.sh

test-docker-scenario:
	cd tests && ./scripts/test-runner.sh --scenario $(SCENARIO)

test-docker-no-cleanup:
	cd tests && ./scripts/test-runner.sh --no-cleanup

test-docker-clean:
	cd tests && ./scripts/test-cleanup.sh

test-docker-setup:
	cd tests && ./scripts/test-setup.sh
```

## Test Execution

### Quick Start

```bash
# Run all tests
make test-docker

# Run specific scenario
make test-docker-scenario SCENARIO=2

# Run tests without cleanup (for debugging)
make test-docker-no-cleanup

# Manual cleanup
make test-docker-clean
```

### Manual Container Access

```bash
# Start test environment
make test-docker-setup

# Access alpha container
docker exec -it env-sync-alpha bash

# Inside container:
env-sync status              # Check status
env-sync list                # List secrets
env-sync show TEST_KEY       # Show specific secret
env-sync discover            # Discover peers
cat ~/.secrets.env           # View raw file

# View logs
docker logs env-sync-alpha
docker exec env-sync-alpha cat ~/.config/env-sync/logs/env-sync.log
```

## Debugging Failed Tests

### Check Container Health

```bash
# List running containers
docker ps | grep env-sync

# Check container logs
docker logs env-sync-alpha
docker logs env-sync-beta
docker logs env-sync-gamma

# Check if services are running
docker exec env-sync-alpha pgrep sshd
docker exec env-sync-alpha pgrep avahi-daemon

# Test mDNS discovery
docker exec env-sync-alpha avahi-browse -a -t
docker exec env-sync-alpha dns-sd -B _envsync._tcp
```

### Verify Network Connectivity

```bash
# Test SSH between containers
docker exec env-sync-alpha ssh -o ConnectTimeout=5 beta.local echo "SSH works"
docker exec env-sync-alpha ssh -o ConnectTimeout=5 gamma.local echo "SSH works"

# Test SCP
docker exec env-sync-alpha scp beta.local:/root/.secrets.env /tmp/test.env

# Check mDNS resolution
docker exec env-sync-alpha getent hosts beta.local
docker exec env-sync-alpha getent hosts gamma.local
```

### Inspect Secrets Files

```bash
# View secrets on each container
echo "=== Alpha ==="
docker exec env-sync-alpha cat /root/.secrets.env
echo ""
echo "=== Beta ==="
docker exec env-sync-beta cat /root/.secrets.env
echo ""
echo "=== Gamma ==="
docker exec env-sync-gamma cat /root/.secrets.env

# Check AGE keys
docker exec env-sync-alpha env-sync key show
docker exec env-sync-alpha ls -la ~/.config/env-sync/keys/
```

## Success Criteria

- [ ] All 4 test scenarios pass consistently
- [ ] Tests complete in under 5 minutes
- [ ] No manual intervention required
- [ ] Clean shutdown and resource cleanup
- [ ] Works in CI/CD pipelines
- [ ] Clear error messages on failure
- [ ] Logs available for debugging

## Future Enhancements

### Phase 6: Advanced Testing (Post V2)

- **Network Partition Simulation**: Test split-brain scenarios
- **Chaos Testing**: Random container restarts during sync
- **Performance Testing**: Measure sync latency with 1000+ secrets
- **Conflict Resolution**: Test simultaneous edits on multiple nodes
- **Encryption Rotation**: Test key rotation scenarios
- **Upgrade Testing**: Test pre-v2 to v2.0 migration

### Phase 7: Extended Scenarios

- **Mixed Mode**: Test encrypted and unencrypted containers together
- **Partial Connectivity**: Test when only some nodes can reach each other
- **Large File Sync**: Test with multi-MB secrets files
- **Concurrent Access**: Test multiple syncs happening simultaneously

## Documentation

### For Developers

- Update AGENTS.md with testing information
- Add test development guide
- Document common failures and solutions

### For Users

- Add "Testing" section to README.md
- Document how to run tests locally
- Provide troubleshooting guide

## Resources

- Docker networking: https://docs.docker.com/network/
- mDNS in Docker: https://github.com/aweber/docker-mdns
- Avahi documentation: https://www.avahi.org/doxygen/html/
- SSH key management: https://www.ssh.com/academy/ssh/keygen
- AGE encryption: https://age-encryption.org/
