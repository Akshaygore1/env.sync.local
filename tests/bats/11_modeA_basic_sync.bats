#!/usr/bin/env bats

# 11_modeA_basic_sync.bats — Mode A: basic HTTP sync
# Tests plaintext sync over HTTP between containers

load 'test_helper'

@test "modeA: Add test secret to alpha" {
    run add_secret "$CONTAINER_ALPHA" "HTTP_TEST_KEY" "http-value-123"
    [ "$status" -eq 0 ]
}

@test "modeA: Verify secret exists on alpha" {
    run get_secret "$CONTAINER_ALPHA" "HTTP_TEST_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "http-value-123" ]
}

@test "modeA: Sync beta from network (HTTP)" {
    run trigger_sync_http "$CONTAINER_BETA"
    [ "$status" -eq 0 ]
}

@test "modeA: Verify secret synced to beta via HTTP" {
    run get_secret "$CONTAINER_BETA" "HTTP_TEST_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "http-value-123" ]
}

@test "modeA: Sync gamma from network (HTTP)" {
    run trigger_sync_http "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
}

@test "modeA: Verify secret synced to gamma via HTTP" {
    run get_secret "$CONTAINER_GAMMA" "HTTP_TEST_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "http-value-123" ]
}

@test "modeA: Add secret on beta and propagate via HTTP" {
    run add_secret "$CONTAINER_BETA" "HTTP_SECOND_KEY" "http-beta-456"
    [ "$status" -eq 0 ]

    run trigger_sync_http "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]

    run trigger_sync_http "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]

    run get_secret "$CONTAINER_ALPHA" "HTTP_SECOND_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "http-beta-456" ]

    run get_secret "$CONTAINER_GAMMA" "HTTP_SECOND_KEY"
    [ "$status" -eq 0 ]
    [ "$output" = "http-beta-456" ]
}

@test "modeA: path command works in HTTP mode" {
    run container_exec "$CONTAINER_ALPHA" env-sync path
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^/.+\.secrets\.env$ ]]
}

@test "modeA: path --backup returns backup directory" {
    run container_exec "$CONTAINER_ALPHA" env-sync path --backup
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^/.+/backups$ ]]
}
