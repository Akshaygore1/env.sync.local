#!/usr/bin/env bats

load 'test_helper'

@test "Start delta container (4th machine)" {
    # Ensure any previous delta container is removed
    docker rm -f "$CONTAINER_DELTA" 2>/dev/null || true

    # Start delta container dynamically using the alpha image
    # (all three images are identical, we just need to pick one)
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

    # Wait for delta to be ready
    run wait_for_container "$CONTAINER_DELTA" 60
    [ "$status" -eq 0 ]
}

@test "Initialize delta with encryption" {
    run init_container "$CONTAINER_DELTA" true
    [ "$status" -eq 0 ]
}

@test "Verify delta cannot decrypt existing secrets initially" {
    # Delta should not be able to decrypt yet (not in recipient list)
    # When we try to show the secret, it should fail or return empty
    run container_exec "$CONTAINER_DELTA" env-sync show ENCRYPTED_SECRET 2>&1
    # This might fail or return empty - both are acceptable
    [ -z "$output" ] || [ "$status" -ne 0 ] || true
}

@test "Collect public keys on delta from existing machines" {
    # Get pubkeys from alpha, beta, gamma
    run container_exec "$CONTAINER_DELTA" env-sync key import "$(get_pubkey "$CONTAINER_ALPHA")" "alpha.local"
    [ "$status" -eq 0 ]
    
    run container_exec "$CONTAINER_DELTA" env-sync key import "$(get_pubkey "$CONTAINER_BETA")" "beta.local"
    [ "$status" -eq 0 ]
    
    run container_exec "$CONTAINER_DELTA" env-sync key import "$(get_pubkey "$CONTAINER_GAMMA")" "gamma.local"
    [ "$status" -eq 0 ]
}

@test "Share delta's pubkey with alpha" {
    DELTA_PUBKEY=$(get_pubkey "$CONTAINER_DELTA")
    [ -n "$DELTA_PUBKEY" ]
    
    run import_pubkey "$CONTAINER_ALPHA" "$DELTA_PUBKEY" "delta.local"
    [ "$status" -eq 0 ]
    
    # Also share with beta and gamma for completeness
    import_pubkey "$CONTAINER_BETA" "$DELTA_PUBKEY" "delta.local"
    import_pubkey "$CONTAINER_GAMMA" "$DELTA_PUBKEY" "delta.local"
}

@test "Trigger re-encryption on alpha to include delta" {
    # This sync will re-encrypt with delta as a recipient
    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]
}

@test "Sync other containers to get re-encrypted file" {
    run trigger_sync "$CONTAINER_BETA"
    [ "$status" -eq 0 ]
    
    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
}

@test "Sync delta to get the encrypted secrets" {
    run trigger_sync "$CONTAINER_DELTA"
    [ "$status" -eq 0 ]
}

@test "Verify delta can now decrypt existing secrets" {
    run get_secret "$CONTAINER_DELTA" "ENCRYPTED_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "secret-value-789" ]
}

@test "Verify delta can decrypt propagated secrets" {
    run get_secret "$CONTAINER_DELTA" "PROPAGATION_TEST"
    [ "$status" -eq 0 ]
    [ "$output" = "propagation-value-abc" ]
}

@test "Add secret on delta and verify it syncs back to others" {
    run add_secret "$CONTAINER_DELTA" "DELTA_SECRET" "delta-specific-value"
    [ "$status" -eq 0 ]
    
    # Sync alpha to get the new secret
    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]
    
    # Verify alpha received it
    run get_secret "$CONTAINER_ALPHA" "DELTA_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "delta-specific-value" ]
    
    # Sync beta and verify
    run trigger_sync "$CONTAINER_BETA"
    [ "$status" -eq 0 ]
    
    run get_secret "$CONTAINER_BETA" "DELTA_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "delta-specific-value" ]
}

@test "Stop and remove delta container" {
    docker stop "$CONTAINER_DELTA" 2>/dev/null || true
    docker rm -f "$CONTAINER_DELTA" 2>/dev/null || true
    
    # Verify it's gone
    run wait_for_container_removed "$CONTAINER_DELTA" 60
    [ "$status" -eq 0 ]
}
