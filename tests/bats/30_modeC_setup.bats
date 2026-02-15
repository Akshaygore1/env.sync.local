#!/usr/bin/env bats

# 30_modeC_setup.bats — Mode C: secure-peer
# Set up containers for mTLS-based secure peer mode

load 'test_helper'

@test "modeC: Reset containers to clean state" {
    run reset_containers
    [ "$status" -eq 0 ]
}

@test "modeC: Set all containers to secure-peer mode" {
    for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
        run set_mode "$container" "secure-peer"
        [ "$status" -eq 0 ]
    done
}

@test "modeC: Verify mode is secure-peer on all containers" {
    for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
        run container_exec "$container" env-sync mode get
        [ "$status" -eq 0 ]
        [[ "$output" =~ "secure-peer" ]]
    done
}

@test "modeC: TLS identity was generated on all containers" {
    for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
        run container_exec "$container" test -f /home/envsync/.config/env-sync/tls/transport.crt
        [ "$status" -eq 0 ]
        run container_exec "$container" test -f /home/envsync/.config/env-sync/tls/transport.key
        [ "$status" -eq 0 ]
    done
}

@test "modeC: Initialize alpha with encrypted secrets" {
    run init_container "$CONTAINER_ALPHA" true
    [ "$status" -eq 0 ]
}

@test "modeC: Initialize beta with encrypted secrets" {
    run init_container "$CONTAINER_BETA" true
    [ "$status" -eq 0 ]
}

@test "modeC: Initialize gamma with encrypted secrets" {
    run init_container "$CONTAINER_GAMMA" true
    [ "$status" -eq 0 ]
}

@test "modeC: Start mTLS server on alpha" {
    start_mtls_server "$CONTAINER_ALPHA"

    run container_exec "$CONTAINER_ALPHA" pgrep -f "env-sync serve"
    [ "$status" -eq 0 ]
}

@test "modeC: Start mTLS server on beta" {
    start_mtls_server "$CONTAINER_BETA"

    run container_exec "$CONTAINER_BETA" pgrep -f "env-sync serve"
    [ "$status" -eq 0 ]
}

@test "modeC: Start mTLS server on gamma" {
    start_mtls_server "$CONTAINER_GAMMA"

    run container_exec "$CONTAINER_GAMMA" pgrep -f "env-sync serve"
    [ "$status" -eq 0 ]
}
