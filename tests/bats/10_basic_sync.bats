#!/usr/bin/env bats

load 'test_helper'

@test "Initialize alpha with plaintext secrets" {
    run init_container "$CONTAINER_ALPHA" false
    [ "$status" -eq 0 ]
}

@test "Add test secret to alpha" {
    run add_secret "$CONTAINER_ALPHA" "TEST_KEY" "alpha-value-123"
    [ "$status" -eq 0 ]
}

@test "Verify secret exists on alpha" {
    run get_secret "$CONTAINER_ALPHA" "TEST_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "alpha-value-123" ]
}

@test "Initialize beta with plaintext" {
    run init_container "$CONTAINER_BETA" false
    [ "$status" -eq 0 ]
}

@test "Initialize gamma with plaintext" {
    run init_container "$CONTAINER_GAMMA" false
    [ "$status" -eq 0 ]
}

@test "Trigger sync on beta" {
    run trigger_sync "$CONTAINER_BETA"
    [ "$status" -eq 0 ]
}

@test "Trigger sync on gamma" {
    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
}

@test "Verify secret synced to beta" {
    run get_secret "$CONTAINER_BETA" "TEST_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "alpha-value-123" ]
}

@test "Verify secret synced to gamma" {
    run get_secret "$CONTAINER_GAMMA" "TEST_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "alpha-value-123" ]
}

@test "Add second secret to beta and verify propagation" {
    run add_secret "$CONTAINER_BETA" "SECOND_KEY" "beta-value-456"
    [ "$status" -eq 0 ]
    
    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]
    
    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
    
    run get_secret "$CONTAINER_ALPHA" "SECOND_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "beta-value-456" ]
    
    run get_secret "$CONTAINER_GAMMA" "SECOND_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "beta-value-456" ]
}
