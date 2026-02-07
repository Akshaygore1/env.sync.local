# Docker-Based Integration Tests

This directory contains Docker-based integration tests for env-sync using **bats-core** (Bash Automated Testing System).

## Overview

The tests spin up 3 Docker containers (alpha, beta, gamma) that simulate a distributed environment with:
- SSH server for SCP-based sync
- Avahi daemon for mDNS discovery
- AGE encryption support
- Shared SSH keys for seamless authentication

## Quick Start

### Run All Tests

From the project root:

```bash
./test-dockers.sh
```

This will:
1. Build the Go binary (`make build`)
1. Check prerequisites (Docker, docker-compose, git)
2. Install bats-core testing framework
3. Generate SSH keys for containers
4. Build and start Docker containers
5. Run all integration tests
6. Clean up containers after tests

### Start Test Environment for Manual Testing

```bash
./start-test-env.sh
```

This starts the containers and leaves them running for manual exploration.

Access containers:
```bash
docker exec -it env-sync-alpha bash
docker exec -it env-sync-beta bash
docker exec -it env-sync-gamma bash
```

Stop the environment:
```bash
./start-test-env.sh --stop
```

### Run Specific Tests

```bash
# Run only tests matching "basic"
./test-dockers.sh --filter basic

# Run tests without cleanup (for debugging)
./test-dockers.sh --no-cleanup
```

## Test Structure

### Test Files (`tests/bats/`)

- **01_setup.bats** - Docker/build checks and container startup
- **10_basic_sync.bats** - Unencrypted sync tests
- **20_encrypted_sync.bats** - AGE encryption tests
- **30_propagation.bats** - Secret propagation tests
- **40_add_machine.bats** - Adding 4th machine (delta) tests
- **99_teardown.bats** - Cleanup tests

### Helper Files (`tests/bats/`)

- **test_helper.bash** - Common functions and setup for all tests

### Utilities (`tests/utils/`)

- **generate-ssh-keys.sh** - Generate shared SSH keys
- **verify-secrets.sh** - Verify secrets match across containers
- **wait-for-sync.sh** - Wait for sync to complete
- **check-health.sh** - Check container and service health

### Docker Configuration (`tests/docker/`)

- **Dockerfile** - Ubuntu 22.04 base with env-sync, SSH, avahi, age
- **docker-compose.yml** - 3-container setup with networking
- **entrypoint.sh** - Container startup script

## Test Scenarios

### 1. Basic Sync (Unencrypted)

Tests plaintext secrets synchronization:
- Initialize alpha with secrets
- Sync to beta and gamma
- Verify all containers have the same secrets

### 2. Encrypted Sync

Tests AGE encryption across containers:
- Initialize all containers with encryption
- Exchange AGE public keys
- Add encrypted secrets
- Verify all containers can decrypt

### 3. Secret Propagation

Tests that new secrets propagate to all containers:
- Add secret to beta
- Trigger sync on alpha and gamma
- Verify all have the new secret
- Verify old secrets still exist

### 4. Add New Machine

Tests adding a 4th container dynamically:
- Start delta container
- Initialize with encryption
- Exchange public keys
- Verify delta can decrypt existing secrets
- Add secret on delta
- Verify it propagates to others
- Cleanup delta

## How It Works

### Container Setup

Each container has:
- **Hostname**: alpha.local, beta.local, gamma.local
- **IP**: 172.20.0.2, 172.20.0.3, 172.20.0.4
- **Network**: Shared bridge network with mDNS support
- **Services**: SSH (port 22), Avahi daemon
- **Keys**: Shared Ed25519 SSH key for authentication

### mDNS Discovery

Containers discover each other via Avahi (mDNS/Bonjour):
```bash
# Inside any container
avahi-browse -a -t  # Browse all services
getent hosts beta.local  # Resolve hostname
```

### SSH Authentication

All containers share the same SSH key pair:
- **Private key**: `tests/docker/ssh-keys/id_ed25519`
- **Public key**: `tests/docker/ssh-keys/id_ed25519.pub`
- **Authorized keys**: `tests/docker/ssh-keys/authorized_keys`

This allows passwordless SSH between containers.

### bats-core Testing

Tests use the bats-core framework:

```bash
@test "Description of what we're testing" {
    # Setup
    run some_command
    
    # Assertions
    [ "$status" -eq 0 ]  # Check exit code
    [ "$output" = "expected" ]  # Check output
}
```

The `test_helper.bash` provides utility functions:
- `init_container` - Initialize env-sync in a container
- `add_secret` - Add a secret
- `get_secret` - Get a secret value
- `trigger_sync` - Trigger sync
- `verify_secret_all` - Verify secret on all containers

## Debugging Failed Tests

### View Container Logs

```bash
docker logs env-sync-alpha
docker logs env-sync-beta
docker logs env-sync-gamma
```

### Access Running Containers

```bash
docker exec -it env-sync-alpha bash

# Inside container:
env-sync status              # Check sync status
env-sync list                # List secrets
env-sync show KEY            # Show secret value
cat ~/.secrets.env           # View raw secrets file
ls ~/.config/env-sync/logs/  # View logs
```

### Check Network Connectivity

```bash
# Test mDNS
docker exec env-sync-alpha avahi-browse -a -t

# Test SSH
docker exec env-sync-alpha ssh beta.local echo OK

# Test hostname resolution
docker exec env-sync-alpha getent hosts beta.local
```

### Run Tests With Debug Output

```bash
./test-dockers.sh --no-cleanup
# Then inspect containers manually
```

## Continuous Integration

The tests can run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run Docker tests
  run: ./test-dockers.sh
```

Requirements:
- Docker daemon running
- Docker Compose available
- Git available (for bats-core)

## Adding New Tests

1. Create a new `.bats` file in `tests/bats/`
2. Load the helper: `load 'test_helper'`
3. Use `@test` blocks with descriptive names
4. Use helper functions from `test_helper.bash`
5. Run with `./test-dockers.sh --filter your_pattern`

Example:

```bash
#!/usr/bin/env bats

load 'test_helper'

@test "My new test scenario" {
    run init_container "$CONTAINER_ALPHA" true
    [ "$status" -eq 0 ]
    
    run add_secret "$CONTAINER_ALPHA" "MY_KEY" "my_value"
    [ "$status" -eq 0 ]
    
    run get_secret "$CONTAINER_ALPHA" "MY_KEY"
    [ "$output" = "my_value" ]
}
```

## Troubleshooting

### "Docker daemon not running"

Make sure Docker Desktop or Docker Engine is running:
```bash
docker info
```

### "Containers fail to start"

Check if ports are already in use or if you have permission issues:
```bash
docker-compose -f tests/docker/docker-compose.yml logs
```

### "SSH connection fails"

SSH keys might not be generated:
```bash
./tests/utils/generate-ssh-keys.sh
```

### "mDNS not working"

The containers need privileged mode for avahi-daemon. Check if it started:
```bash
docker exec env-sync-alpha pgrep avahi-daemon
```

## License

Same as the main project (MIT).
