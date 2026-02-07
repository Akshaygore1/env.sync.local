#!/usr/bin/env bats

load 'test_helper'

@test "Add new secret to beta in encrypted setup" {
    run add_secret "$CONTAINER_BETA" "PROPAGATION_TEST" "propagation-value-abc"
    [ "$status" -eq 0 ]
}

@test "Trigger sync on alpha to receive new secret" {
    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]
}

@test "Trigger sync on gamma to receive new secret" {
    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
}

@test "Verify alpha received the propagated secret" {
    run get_secret "$CONTAINER_ALPHA" "PROPAGATION_TEST"
    [ "$status" -eq 0 ]
    [ "$output" = "propagation-value-abc" ]
}

@test "Verify gamma received the propagated secret" {
    run get_secret "$CONTAINER_GAMMA" "PROPAGATION_TEST"
    [ "$status" -eq 0 ]
    [ "$output" = "propagation-value-abc" ]
}

@test "Verify original encrypted secret still exists on all containers" {
    run parallel_run \
        "verify_secret \"$CONTAINER_ALPHA\" \"ENCRYPTED_SECRET\" \"secret-value-789\"" \
        "verify_secret \"$CONTAINER_BETA\" \"ENCRYPTED_SECRET\" \"secret-value-789\"" \
        "verify_secret \"$CONTAINER_GAMMA\" \"ENCRYPTED_SECRET\" \"secret-value-789\""
    [ "$status" -eq 0 ]
}

@test "Add multiple secrets to gamma" {
    run add_secret "$CONTAINER_GAMMA" "MULTI_1" "value-1"
    [ "$status" -eq 0 ]
    
    run add_secret "$CONTAINER_GAMMA" "MULTI_2" "value-2"
    [ "$status" -eq 0 ]
    
    run add_secret "$CONTAINER_GAMMA" "MULTI_3" "value-3"
    [ "$status" -eq 0 ]
}

@test "Sync all containers and verify multiple secrets propagated" {
    run parallel_run \
        "trigger_sync \"$CONTAINER_ALPHA\"" \
        "trigger_sync \"$CONTAINER_BETA\""
    [ "$status" -eq 0 ]
    
    run parallel_run \
        "verify_secret \"$CONTAINER_ALPHA\" \"MULTI_1\" \"value-1\"" \
        "verify_secret \"$CONTAINER_BETA\" \"MULTI_2\" \"value-2\"" \
        "verify_secret \"$CONTAINER_ALPHA\" \"MULTI_3\" \"value-3\""
    [ "$status" -eq 0 ]
}
