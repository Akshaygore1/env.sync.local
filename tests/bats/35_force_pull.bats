#!/usr/bin/env bats

load 'test_helper'

@test "Add FORCE_PULL_TEST secret to alpha with initial value" {
    run add_secret "$CONTAINER_ALPHA" "FORCE_PULL_TEST" "alpha-value-initial"
    [ "$status" -eq 0 ]
}

@test "Wait and add FORCE_PULL_TEST secret to beta with different value" {
    # Sleep to ensure beta's timestamp is newer
    sleep 2
    run add_secret "$CONTAINER_BETA" "FORCE_PULL_TEST" "beta-value-newer"
    [ "$status" -eq 0 ]
}

@test "Verify beta has its own value before force-pull" {
    run get_secret "$CONTAINER_BETA" "FORCE_PULL_TEST"
    [ "$status" -eq 0 ]
    [ "$output" = "beta-value-newer" ]
}

@test "Verify alpha has its original value before force-pull" {
    run get_secret "$CONTAINER_ALPHA" "FORCE_PULL_TEST"
    [ "$status" -eq 0 ]
    [ "$output" = "alpha-value-initial" ]
}

@test "Force pull from alpha to beta (should overwrite beta's newer value)" {
    run container_exec "$CONTAINER_BETA" env-sync sync --force-pull alpha.local
    [ "$status" -eq 0 ]
}

@test "Verify beta now has alpha's value after force-pull" {
    run get_secret "$CONTAINER_BETA" "FORCE_PULL_TEST"
    [ "$status" -eq 0 ]
    [ "$output" = "alpha-value-initial" ]
}

@test "Verify alpha still has its original value (unchanged)" {
    run get_secret "$CONTAINER_ALPHA" "FORCE_PULL_TEST"
    [ "$status" -eq 0 ]
    [ "$output" = "alpha-value-initial" ]
}

@test "Add another key to alpha and beta for second force-pull test" {
    run add_secret "$CONTAINER_ALPHA" "FORCE_PULL_TEST2" "alpha-value-2"
    [ "$status" -eq 0 ]

    # Sleep to ensure beta's timestamp is newer
    sleep 2
    run add_secret "$CONTAINER_BETA" "FORCE_PULL_TEST2" "beta-value-2-newer"
    [ "$status" -eq 0 ]
}

@test "Normal sync from beta to alpha should keep alpha's value (timestamp-based merge)" {
    # First verify beta has the newer value
    run get_secret "$CONTAINER_BETA" "FORCE_PULL_TEST2"
    [ "$status" -eq 0 ]
    [ "$output" = "beta-value-2-newer" ]

    # Normal sync should merge based on timestamps
    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]

    # Alpha should now have beta's newer value due to timestamp-based merge
    run get_secret "$CONTAINER_ALPHA" "FORCE_PULL_TEST2"
    [ "$status" -eq 0 ]
    [ "$output" = "beta-value-2-newer" ]
}

@test "Force-pull without hostname should fail with error" {
    run container_exec "$CONTAINER_BETA" env-sync sync --force-pull
    [ "$status" -ne 0 ]
}

@test "Verify backup was created during force-pull" {
    run container_exec "$CONTAINER_BETA" ls /home/envsync/.config/env-sync/backups/
    [ "$status" -eq 0 ]
    # Should have at least one backup file
    [ -n "$output" ]
}
