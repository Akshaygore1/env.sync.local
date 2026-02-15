#!/usr/bin/env bash

# Test helper for env-sync Docker tests (v3.0 — 3-mode support)
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

# ─────────────────────────────────────────────────────────────
# Docker helpers
# ─────────────────────────────────────────────────────────────

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

# ─────────────────────────────────────────────────────────────
# Container lifecycle helpers
# ─────────────────────────────────────────────────────────────

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

        local all_running=true
        for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
            if ! docker ps --format "{{.Names}}" | grep -q "^${container}$"; then
                all_running=false
                break
            fi
        done

        if [ "$all_running" = true ]; then
            sleep 3
            return 0
        fi

        sleep 2
    done
}

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

# ─────────────────────────────────────────────────────────────
# Container execution helpers
# ─────────────────────────────────────────────────────────────

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

# ─────────────────────────────────────────────────────────────
# Mode management helpers
# ─────────────────────────────────────────────────────────────

# Set mode on a container
set_mode() {
    local container="$1"
    local mode="$2"
    container_exec "$container" env-sync mode set "$mode" --yes
}

# Get current mode from a container
get_mode() {
    local container="$1"
    container_exec "$container" env-sync mode get 2>/dev/null | head -1 | awk '{print $NF}'
}

# Set mode on all three primary containers
set_mode_all() {
    local mode="$1"
    for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
        set_mode "$container" "$mode"
    done
}

# ─────────────────────────────────────────────────────────────
# Secrets management helpers
# ─────────────────────────────────────────────────────────────

init_container() {
    local container="$1"
    local encrypted="${2:-false}"

    # Remove existing secrets file
    container_exec "$container" rm -f /home/envsync/.secrets.env

    if [ "$encrypted" = true ]; then
        container_exec "$container" rm -f /home/envsync/.config/env-sync/keys/age_key
        container_exec "$container" rm -f /home/envsync/.config/env-sync/keys/age_key.pub
        container_exec "$container" env-sync init --encrypted
    else
        container_exec "$container" env-sync init
    fi
}

add_secret() {
    local container="$1"
    local key="$2"
    local value="$3"

    container_exec "$container" env-sync add "$key=$value"
}

get_secret() {
    local container="$1"
    local key="$2"

    container_exec "$container" env-sync show "$key" 2>/dev/null || echo ""
}

trigger_sync() {
    local container="$1"
    local force="${2:--f}"

    container_exec "$container" env-sync sync "$force"
}

verify_secret() {
    local container="$1"
    local key="$2"
    local expected="$3"
    local actual

    actual=$(get_secret "$container" "$key")
    [[ "$actual" = "$expected" ]]
}

# ─────────────────────────────────────────────────────────────
# HTTP server helpers (Mode A)
# ─────────────────────────────────────────────────────────────

# Start the HTTP server on a container (background)
start_server() {
    local container="$1"
    local port="${2:-5739}"

    # Start server in background
    docker exec -d --user envsync "$container" env-sync serve -p "$port" -q
    sleep 2 # Give server time to start
}

# Stop the HTTP server on a container
stop_server() {
    local container="$1"
    container_exec "$container" pkill -f "env-sync serve" 2>/dev/null || true
    sleep 1
}

# Sync using HTTP (mode A)
trigger_sync_http() {
    local container="$1"
    container_exec "$container" env-sync sync --insecure-http
}

# ─────────────────────────────────────────────────────────────
# Key management helpers (Mode B)
# ─────────────────────────────────────────────────────────────

get_pubkey() {
    local container="$1"
    container_exec "$container" env-sync key export 2>/dev/null || echo ""
}

import_pubkey() {
    local container="$1"
    local pubkey="$2"
    local hostname="$3"

    container_exec "$container" env-sync key import "$pubkey" "$hostname"
}

# ─────────────────────────────────────────────────────────────
# Peer management helpers (Mode C)
# ─────────────────────────────────────────────────────────────

# Create a peer invite on a container
peer_invite() {
    local container="$1"
    local expiry="${2:-24h}"

    container_exec "$container" env-sync peer invite --expiry "$expiry" 2>/dev/null
}

# Extract invite token from peer invite output
extract_invite_token() {
    local output="$1"
    echo "$output" | grep "Token:" | awk '{print $2}'
}

# Request peer access
peer_request() {
    local container="$1"
    local host="$2"
    local token="$3"

    container_exec "$container" env-sync peer request "$host" "$token"
}

# Approve a peer
peer_approve() {
    local container="$1"
    local peer_id="$2"

    container_exec "$container" env-sync peer approve "$peer_id"
}

# Revoke a peer
peer_revoke() {
    local container="$1"
    local peer_id="$2"

    container_exec "$container" env-sync peer revoke "$peer_id"
}

# List peers
peer_list() {
    local container="$1"
    container_exec "$container" env-sync peer list 2>/dev/null
}

# Show trust info
peer_trust_show() {
    local container="$1"
    container_exec "$container" env-sync peer trust show 2>/dev/null
}

# Exchange TLS certs between containers for bidirectional mTLS
# Each container needs the other containers' transport certs in its trusted dir
exchange_tls_certs() {
    local containers=("$@")
    local cert_dir="/home/envsync/.config/env-sync/tls"
    local trusted_dir="$cert_dir/trusted"

    for src in "${containers[@]}"; do
        local src_name=$(docker exec "$src" hostname | tr -d '\n')
        local src_cert=$(docker exec "$src" cat "$cert_dir/transport.crt" 2>/dev/null)

        for dst in "${containers[@]}"; do
            if [ "$src" != "$dst" ]; then
                container_exec "$dst" mkdir -p "$trusted_dir"
                docker exec "$dst" bash -c "cat > $trusted_dir/${src_name}.crt" <<< "$src_cert"
            fi
        done
    done
}

# Start mTLS server (for mode C)
start_mtls_server() {
    local container="$1"
    local port="${2:-5739}"

    docker exec -d --user envsync "$container" env-sync serve -p "$port" -q
    sleep 2
}

# ─────────────────────────────────────────────────────────────
# Parallel & utility helpers
# ─────────────────────────────────────────────────────────────

parallel_run() {
    local jobs="${ENV_SYNC_PARALLEL_JOBS:-4}"

    export -f container_exec
    export -f trigger_sync
    export -f trigger_sync_http
    export -f get_secret
    export -f get_pubkey
    export -f import_pubkey
    export -f verify_secret
    export -f set_mode
    export -f start_server
    export -f stop_server

    parallel --halt now,fail=1 -k -j "$jobs" -- bash -c '{}' ::: "$@"
}

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

cleanup_all() {
    echo "Cleaning up test environment..."
    $DOCKER_COMPOSE -f "$DOCKER_DIR/docker-compose.yml" down -v 2>/dev/null || true
    docker stop "$CONTAINER_DELTA" 2>/dev/null || true
    docker rm "$CONTAINER_DELTA" 2>/dev/null || true
    echo "Cleanup complete"
}

# Reset containers to clean state (remove all env-sync data)
reset_containers() {
    for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
        container_exec "$container" rm -rf /home/envsync/.secrets.env 2>/dev/null || true
        container_exec "$container" rm -rf /home/envsync/.config/env-sync/keys/age_key 2>/dev/null || true
        container_exec "$container" rm -rf /home/envsync/.config/env-sync/keys/age_key.pub 2>/dev/null || true
        container_exec "$container" rm -rf /home/envsync/.config/env-sync/keys/known_hosts/ 2>/dev/null || true
        container_exec "$container" rm -rf /home/envsync/.config/env-sync/peers/ 2>/dev/null || true
        container_exec "$container" rm -rf /home/envsync/.config/env-sync/identity/ 2>/dev/null || true
        container_exec "$container" rm -rf /home/envsync/.config/env-sync/invites/ 2>/dev/null || true
        container_exec "$container" rm -rf /home/envsync/.config/env-sync/events/ 2>/dev/null || true
        container_exec "$container" rm -f /home/envsync/.config/env-sync/mode 2>/dev/null || true
        container_exec "$container" rm -rf /home/envsync/.config/env-sync/backups/ 2>/dev/null || true
        stop_server "$container" 2>/dev/null || true
    done
}
