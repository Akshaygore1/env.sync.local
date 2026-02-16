#!/usr/bin/env bats

# 21_modeB_basic_sync.bats — Mode B: basic SSH sync (plaintext)
# Tests basic SCP/SSH sync between containers

load 'test_helper'

@test "modeB: Add test secret to alpha" {
    run add_secret "$CONTAINER_ALPHA" "SSH_TEST_KEY" "ssh-value-123"
    [ "$status" -eq 0 ]
}

@test "modeB: Verify secret exists on alpha" {
    run get_secret "$CONTAINER_ALPHA" "SSH_TEST_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "ssh-value-123" ]
}

@test "modeB: Trigger sync on beta (SSH)" {
    run trigger_sync "$CONTAINER_BETA"
    [ "$status" -eq 0 ]
}

@test "modeB: Verify secret synced to beta via SSH" {
    run get_secret "$CONTAINER_BETA" "SSH_TEST_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "ssh-value-123" ]
}

@test "modeB: Trigger sync on gamma (SSH)" {
    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
}

@test "modeB: Verify secret synced to gamma via SSH" {
    run get_secret "$CONTAINER_GAMMA" "SSH_TEST_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "ssh-value-123" ]
}

@test "modeB: Bidirectional sync — add on beta, sync to others" {
    run add_secret "$CONTAINER_BETA" "SSH_SECOND_KEY" "ssh-beta-456"
    [ "$status" -eq 0 ]

    run parallel_run \
        "trigger_sync \"$CONTAINER_ALPHA\"" \
        "trigger_sync \"$CONTAINER_GAMMA\""
    [ "$status" -eq 0 ]

    run parallel_run \
        "verify_secret \"$CONTAINER_ALPHA\" \"SSH_SECOND_KEY\" \"ssh-beta-456\"" \
        "verify_secret \"$CONTAINER_GAMMA\" \"SSH_SECOND_KEY\" \"ssh-beta-456\""
    [ "$status" -eq 0 ]
}
