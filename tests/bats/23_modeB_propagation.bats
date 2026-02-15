#!/usr/bin/env bats

# 23_modeB_propagation.bats — Mode B: secret propagation and multi-secret sync

load 'test_helper'

@test "modeB: Add new encrypted secret to beta" {
    run add_secret "$CONTAINER_BETA" "PROPAGATION_TEST" "propagation-value-abc"
    [ "$status" -eq 0 ]
}

@test "modeB: Sync alpha and gamma to receive new secret" {
    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]

    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
}

@test "modeB: Verify propagated secret on alpha" {
    run get_secret "$CONTAINER_ALPHA" "PROPAGATION_TEST"
    [ "$status" -eq 0 ]
    [ "$output" = "propagation-value-abc" ]
}

@test "modeB: Verify propagated secret on gamma" {
    run get_secret "$CONTAINER_GAMMA" "PROPAGATION_TEST"
    [ "$status" -eq 0 ]
    [ "$output" = "propagation-value-abc" ]
}

@test "modeB: Original encrypted secret still exists on all containers" {
    run parallel_run \
        "verify_secret \"$CONTAINER_ALPHA\" \"ENCRYPTED_SECRET\" \"secret-value-789\"" \
        "verify_secret \"$CONTAINER_BETA\" \"ENCRYPTED_SECRET\" \"secret-value-789\"" \
        "verify_secret \"$CONTAINER_GAMMA\" \"ENCRYPTED_SECRET\" \"secret-value-789\""
    [ "$status" -eq 0 ]
}

@test "modeB: Add multiple secrets to gamma" {
    run add_secret "$CONTAINER_GAMMA" "MULTI_1" "value-1"
    [ "$status" -eq 0 ]

    run add_secret "$CONTAINER_GAMMA" "MULTI_2" "value-2"
    [ "$status" -eq 0 ]

    run add_secret "$CONTAINER_GAMMA" "MULTI_3" "value-3"
    [ "$status" -eq 0 ]
}

@test "modeB: Sync all and verify multi-secret propagation" {
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
