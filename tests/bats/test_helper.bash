#!/usr/bin/env bash

# Test helper for env-sync Docker tests
# This file is sourced by all .bats test files

# Setup function - runs before each test
setup() {
    # Define paths
    export TESTS_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)"
    export UTILS_DIR="$TESTS_DIR/utils"
    export DOCKER_DIR="$TESTS_DIR/docker"
    export ENV_SYNC_DISCOVERY_TIMEOUT="${ENV_SYNC_DISCOVERY_TIMEOUT:-2}"
    export ENV_SYNC_PARALLEL_JOBS="${ENV_SYNC_PARALLEL_JOBS:-4}"
    export ENV_SYNC_EXEC_TIMEOUT="${ENV_SYNC_EXEC_TIMEOUT:-300}"

    # Container names
    export CONTAINER_ALPHA="env-sync-alpha"
    export CONTAINER_BETA="env-sync-beta"
    export CONTAINER_GAMMA="env-sync-gamma"
    export CONTAINER_DELTA="env-sync-delta"

    # Colors for output (if terminal supports it)
    export GREEN='\033[0;32m'
    export RED='\033[0;31m'
    export YELLOW='\033[1;33m'
    export NC='\033[0m' # No Color

    # Detect Docker Compose
    if command -v docker-compose &> /dev/null; then
        export DOCKER_COMPOSE="docker-compose"
    elif docker compose version &> /dev/null; then
        export DOCKER_COMPOSE="docker compose"
    fi
}

# Teardown function - runs after each test
teardown() {
    : # Nothing to clean up per-test, full cleanup in 99_teardown.bats
}

# Helper: Check if Docker is available
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo "ERROR: Docker is not installed or not in PATH"
        return 1
    fi

    if ! docker info &> /dev/null; then
        echo "ERROR: Docker daemon is not running"
        return 1
    fi

    return 0
}

# Helper: Check if docker-compose is available
check_docker_compose() {
    if command -v docker-compose &> /dev/null; then
        export DOCKER_COMPOSE="docker-compose"
    elif docker compose version &> /dev/null; then
        export DOCKER_COMPOSE="docker compose"
    else
        echo "ERROR: docker-compose is not installed"
        return 1
    fi

    return 0
}

# Helper: Wait for containers to be ready
wait_for_containers() {
    local timeout=${1:-60}
    local start_time=$(date +%s)

    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [ $elapsed -ge $timeout ]; then
            echo "TIMEOUT: Containers did not become ready within ${timeout}s"
            return 1
        fi

        # Check if all containers are running
        local all_running=true
        for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
            if ! docker ps --format "{{.Names}}" | grep -q "^${container}$"; then
                all_running=false
                break
            fi
        done

        if [ "$all_running" = true ]; then
            # Additional check: wait a bit for services to start
            sleep 3
            return 0
        fi

        sleep 2
    done
}

# Helper: Wait for a single container to be running
wait_for_container() {
    local container="$1"
    local timeout=${2:-30}
    local start_time=$(date +%s)

    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [ $elapsed -ge $timeout ]; then
            echo "TIMEOUT: Container $container did not start within ${timeout}s"
            return 1
        fi

        if docker ps --format "{{.Names}}" | grep -q "^${container}$"; then
            return 0
        fi

        sleep 1
    done
}

# Helper: Wait for a container to be removed
wait_for_container_removed() {
    local container="$1"
    local timeout=${2:-30}
    local start_time=$(date +%s)

    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [ $elapsed -ge $timeout ]; then
            echo "TIMEOUT: Container $container was not removed within ${timeout}s"
            return 1
        fi

        if ! docker ps -a --format "{{.Names}}" | grep -q "^${container}$"; then
            return 0
        fi

        sleep 1
    done
}

# Helper: Wait for all containers to be removed
wait_for_containers_removed() {
    local timeout=${1:-30}
    shift
    local start_time=$(date +%s)

    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [ $elapsed -ge $timeout ]; then
            echo "TIMEOUT: Containers not removed within ${timeout}s"
            return 1
        fi

        local all_gone=true
        for container in "$@"; do
            if docker ps -a --format "{{.Names}}" | grep -q "^${container}$"; then
                all_gone=false
                break
            fi
        done

        if $all_gone; then
            return 0
        fi

        sleep 1
    done
}

# Helper: Wait for a network to be removed
wait_for_network_removed() {
    local network="$1"
    local timeout=${2:-30}
    local start_time=$(date +%s)

    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [ $elapsed -ge $timeout ]; then
            echo "TIMEOUT: Network $network was not removed within ${timeout}s"
            return 1
        fi

        if ! docker network ls --format "{{.Name}}" | grep -q "^${network}$"; then
            return 0
        fi

        sleep 1
    done
}

# Helper: Run command in container as envsync user
container_exec() {
    local container="$1"
    shift
    local env_args=()
    if [[ -n "${ENV_SYNC_DISCOVERY_TIMEOUT:-}" ]]; then
        env_args+=(-e "ENV_SYNC_DISCOVERY_TIMEOUT=$ENV_SYNC_DISCOVERY_TIMEOUT")
    fi
    if [[ ${#env_args[@]} -gt 0 ]]; then
        docker exec "${env_args[@]}" --user envsync "$container" timeout "${ENV_SYNC_EXEC_TIMEOUT}s" "$@"
    else
        docker exec --user envsync "$container" timeout "${ENV_SYNC_EXEC_TIMEOUT}s" "$@"
    fi
}

# Helper: Initialize env-sync in a container
init_container() {
    local container="$1"
    local encrypted="${2:-false}"

    # Remove existing secrets file
    container_exec "$container" rm -f /home/envsync/.secrets.env

    if [ "$encrypted" = true ]; then
        # For encrypted init, also remove any existing keys to ensure fresh generation
        container_exec "$container" rm -f /home/envsync/.config/env-sync/keys/age_key
        container_exec "$container" rm -f /home/envsync/.config/env-sync/keys/age_key.pub
        container_exec "$container" env-sync init --encrypted
    else
        container_exec "$container" env-sync init
    fi
}

# Helper: Add secret to container
add_secret() {
    local container="$1"
    local key="$2"
    local value="$3"

    container_exec "$container" env-sync add "$key=$value"
}

# Helper: Get secret from container
get_secret() {
    local container="$1"
    local key="$2"

    container_exec "$container" env-sync show "$key" 2>/dev/null || echo ""
}

# Helper: Trigger sync on container
trigger_sync() {
    local container="$1"
    local force="${2:--f}"

    container_exec "$container" env-sync sync "$force"
}

# Helper: Verify secret matches expected value
verify_secret() {
    local container="$1"
    local key="$2"
    local expected="$3"
    local actual

    actual=$(get_secret "$container" "$key")
    [[ "$actual" = "$expected" ]]
}

# Helper: Run multiple commands in parallel
parallel_run() {
    local jobs="${ENV_SYNC_PARALLEL_JOBS:-4}"

    export -f container_exec
    export -f trigger_sync
    export -f get_secret
    export -f get_pubkey
    export -f import_pubkey
    export -f verify_secret

    parallel --halt now,fail=1 -k -j "$jobs" -- bash -c '{}' ::: "$@"
}

# Helper: Verify secret exists and matches on all containers
verify_secret_all() {
    local key="$1"
    local expected="$2"
    local failed=0

    for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
        local actual
        actual=$(get_secret "$container" "$key")
        if [ "$actual" != "$expected" ]; then
            echo "  ✗ $container: expected '$expected', got '$actual'"
            failed=1
        fi
    done

    return $failed
}

# Helper: Get AGE public key from container
get_pubkey() {
    local container="$1"
    container_exec "$container" env-sync key export 2>/dev/null || echo ""
}

# Helper: Import public key to container
import_pubkey() {
    local container="$1"
    local pubkey="$2"
    local hostname="$3"

    container_exec "$container" env-sync key import "$pubkey" "$hostname"
}

# Helper: Clean up all test containers and volumes
cleanup_all() {
    echo "Cleaning up test environment..."

    # Stop and remove containers
    $DOCKER_COMPOSE -f "$DOCKER_DIR/docker-compose.yml" down -v 2>/dev/null || true

    # Remove delta if it exists
    docker stop "$CONTAINER_DELTA" 2>/dev/null || true
    docker rm "$CONTAINER_DELTA" 2>/dev/null || true

    echo "Cleanup complete"
}
