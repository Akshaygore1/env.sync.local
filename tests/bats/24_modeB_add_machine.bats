#!/usr/bin/env bats

# 24_modeB_add_machine.bats — Mode B: adding a 4th container dynamically

load 'test_helper'

@test "modeB: Start delta container (4th machine)" {
    docker rm -f "$CONTAINER_DELTA" 2>/dev/null || true

    docker run -d \
        --name "$CONTAINER_DELTA" \
        --hostname delta.local \
        --network env-sync-test \
        --ip 172.20.0.5 \
        -v "$DOCKER_DIR/ssh-keys:/mnt/ssh-keys:ro" \
        -e ENV_SYNC_DEBUG=1 \
        -e CONTAINER_NAME=delta \
        --privileged \
        docker-alpha:latest

    run wait_for_container "$CONTAINER_DELTA" 60
    [ "$status" -eq 0 ]
}

@test "modeB: Set delta to trusted-owner-ssh mode" {
    run set_mode "$CONTAINER_DELTA" "trusted-owner-ssh"
    [ "$status" -eq 0 ]
}

@test "modeB: Initialize delta with encryption" {
    run init_container "$CONTAINER_DELTA" true
    [ "$status" -eq 0 ]
}

@test "modeB: Exchange keys with delta" {
    DELTA_PUBKEY=$(get_pubkey "$CONTAINER_DELTA")
    [ -n "$DELTA_PUBKEY" ]

    # Import existing keys into delta
    run container_exec "$CONTAINER_DELTA" env-sync key import "$(get_pubkey "$CONTAINER_ALPHA")" "alpha.local"
    [ "$status" -eq 0 ]
    run container_exec "$CONTAINER_DELTA" env-sync key import "$(get_pubkey "$CONTAINER_BETA")" "beta.local"
    [ "$status" -eq 0 ]
    run container_exec "$CONTAINER_DELTA" env-sync key import "$(get_pubkey "$CONTAINER_GAMMA")" "gamma.local"
    [ "$status" -eq 0 ]

    # Share delta's key with others
    run import_pubkey "$CONTAINER_ALPHA" "$DELTA_PUBKEY" "delta.local"
    [ "$status" -eq 0 ]
    run import_pubkey "$CONTAINER_BETA" "$DELTA_PUBKEY" "delta.local"
    [ "$status" -eq 0 ]
    run import_pubkey "$CONTAINER_GAMMA" "$DELTA_PUBKEY" "delta.local"
    [ "$status" -eq 0 ]
}

@test "modeB: Re-encrypt and sync to include delta" {
    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]

    run trigger_sync "$CONTAINER_DELTA"
    [ "$status" -eq 0 ]
}

@test "modeB: Verify delta can decrypt existing secrets" {
    run get_secret "$CONTAINER_DELTA" "ENCRYPTED_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "secret-value-789" ]
}

@test "modeB: Verify delta can decrypt propagated secrets" {
    run get_secret "$CONTAINER_DELTA" "PROPAGATION_TEST"
    [ "$status" -eq 0 ]
    [ "$output" = "propagation-value-abc" ]
}

@test "modeB: Add secret on delta and verify sync back" {
    run add_secret "$CONTAINER_DELTA" "DELTA_SECRET" "delta-specific-value"
    [ "$status" -eq 0 ]

    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]

    run get_secret "$CONTAINER_ALPHA" "DELTA_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "delta-specific-value" ]
}

@test "modeB: Stop and remove delta container" {
    docker stop "$CONTAINER_DELTA" 2>/dev/null || true
    docker rm -f "$CONTAINER_DELTA" 2>/dev/null || true

    run wait_for_container_removed "$CONTAINER_DELTA" 60
    [ "$status" -eq 0 ]
}
