#!/usr/bin/env bats

# 25_modeB_force_pull.bats — Mode B: force-pull via SSH

load 'test_helper'

@test "modeB: Add force-pull test secret to alpha" {
    run add_secret "$CONTAINER_ALPHA" "SSH_FORCE_PULL" "alpha-ssh-initial"
    [ "$status" -eq 0 ]
}

@test "modeB: Wait and add conflicting secret to beta" {
    sleep 2
    run add_secret "$CONTAINER_BETA" "SSH_FORCE_PULL" "beta-ssh-newer"
    [ "$status" -eq 0 ]
}

@test "modeB: Force pull from alpha to beta (SSH)" {
    run container_exec "$CONTAINER_BETA" env-sync sync --force-pull alpha.local
    [ "$status" -eq 0 ]
}

@test "modeB: Verify beta has alpha's value after force-pull" {
    run get_secret "$CONTAINER_BETA" "SSH_FORCE_PULL"
    [ "$status" -eq 0 ]
    [ "$output" = "alpha-ssh-initial" ]
}

@test "modeB: Alpha still has its original value" {
    run get_secret "$CONTAINER_ALPHA" "SSH_FORCE_PULL"
    [ "$status" -eq 0 ]
    [ "$output" = "alpha-ssh-initial" ]
}

@test "modeB: Force-pull without hostname fails" {
    run container_exec "$CONTAINER_BETA" env-sync sync --force-pull
    [ "$status" -ne 0 ]
}

@test "modeB: Backup was created during force-pull" {
    run container_exec "$CONTAINER_BETA" ls /home/envsync/.config/env-sync/backups/
    [ "$status" -eq 0 ]
    [ -n "$output" ]
}
