#!/usr/bin/env bats

# 12_modeA_force_pull.bats — Mode A: force-pull via HTTP

load 'test_helper'

@test "modeA: Add FORCE_PULL_TEST secret to alpha" {
    run add_secret "$CONTAINER_ALPHA" "HTTP_FORCE_PULL" "alpha-http-initial"
    [ "$status" -eq 0 ]
}

@test "modeA: Wait and add conflicting secret to beta" {
    sleep 2
    run add_secret "$CONTAINER_BETA" "HTTP_FORCE_PULL" "beta-http-newer"
    [ "$status" -eq 0 ]
}

@test "modeA: Verify values before force-pull" {
    run get_secret "$CONTAINER_ALPHA" "HTTP_FORCE_PULL"
    [ "$status" -eq 0 ]
    [ "$output" = "alpha-http-initial" ]

    run get_secret "$CONTAINER_BETA" "HTTP_FORCE_PULL"
    [ "$status" -eq 0 ]
    [ "$output" = "beta-http-newer" ]
}

@test "modeA: Force pull from alpha to beta" {
    run container_exec "$CONTAINER_BETA" env-sync sync --force-pull alpha.local
    [ "$status" -eq 0 ]
}

@test "modeA: Verify beta has alpha's value after force-pull" {
    run get_secret "$CONTAINER_BETA" "HTTP_FORCE_PULL"
    [ "$status" -eq 0 ]
    [ "$output" = "alpha-http-initial" ]
}

@test "modeA: Force-pull without hostname fails" {
    run container_exec "$CONTAINER_BETA" env-sync sync --force-pull
    [ "$status" -ne 0 ]
}

@test "modeA: Cleanup — stop all HTTP servers" {
    stop_server "$CONTAINER_ALPHA"
    stop_server "$CONTAINER_BETA"
    stop_server "$CONTAINER_GAMMA"
}
