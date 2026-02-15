#!/usr/bin/env bats

# 10_modeA_setup.bats — Mode A: dev-plaintext-http
# Set all containers to dev-plaintext-http mode and initialize

load 'test_helper'

@test "modeA: Reset containers to clean state" {
    run reset_containers
    [ "$status" -eq 0 ]
}

@test "modeA: Set all containers to dev-plaintext-http mode" {
    for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
        run set_mode "$container" "dev-plaintext-http"
        [ "$status" -eq 0 ]
    done
}

@test "modeA: Verify mode is dev-plaintext-http on all containers" {
    for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
        run container_exec "$container" env-sync mode get
        [ "$status" -eq 0 ]
        [[ "$output" =~ "dev-plaintext-http" ]]
    done
}

@test "modeA: Initialize alpha with plaintext secrets" {
    run init_container "$CONTAINER_ALPHA" false
    [ "$status" -eq 0 ]
}

@test "modeA: Initialize beta with plaintext secrets" {
    run init_container "$CONTAINER_BETA" false
    [ "$status" -eq 0 ]
}

@test "modeA: Initialize gamma with plaintext secrets" {
    run init_container "$CONTAINER_GAMMA" false
    [ "$status" -eq 0 ]
}

@test "modeA: Start HTTP server on alpha" {
    start_server "$CONTAINER_ALPHA"

    # Verify server is running
    run container_exec "$CONTAINER_ALPHA" pgrep -f "env-sync serve"
    [ "$status" -eq 0 ]
}

@test "modeA: Start HTTP server on beta" {
    start_server "$CONTAINER_BETA"

    run container_exec "$CONTAINER_BETA" pgrep -f "env-sync serve"
    [ "$status" -eq 0 ]
}

@test "modeA: Start HTTP server on gamma" {
    start_server "$CONTAINER_GAMMA"

    run container_exec "$CONTAINER_GAMMA" pgrep -f "env-sync serve"
    [ "$status" -eq 0 ]
}
