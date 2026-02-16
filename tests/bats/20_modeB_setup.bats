#!/usr/bin/env bats

# 20_modeB_setup.bats — Mode B: trusted-owner-ssh
# Set up containers for SCP/SSH-based sync with encrypted secrets

load 'test_helper'

@test "modeB: Reset containers to clean state" {
    run reset_containers
    [ "$status" -eq 0 ]
}

@test "modeB: Set all containers to trusted-owner-ssh mode" {
    for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
        run set_mode "$container" "trusted-owner-ssh"
        [ "$status" -eq 0 ]
    done
}

@test "modeB: Verify mode is trusted-owner-ssh on all containers" {
    for container in $CONTAINER_ALPHA $CONTAINER_BETA $CONTAINER_GAMMA; do
        run container_exec "$container" env-sync mode get
        [ "$status" -eq 0 ]
        [[ "$output" =~ "trusted-owner-ssh" ]]
    done
}

@test "modeB: SSH connectivity between containers works" {
    run parallel_run \
        "container_exec \"$CONTAINER_ALPHA\" ssh -o ConnectTimeout=10 -o StrictHostKeyChecking=no beta.local echo \"OK\" 2>/dev/null | grep -qx \"OK\"" \
        "container_exec \"$CONTAINER_ALPHA\" ssh -o ConnectTimeout=10 -o StrictHostKeyChecking=no gamma.local echo \"OK\" 2>/dev/null | grep -qx \"OK\""
    [ "$status" -eq 0 ]
}

@test "modeB: Initialize alpha with plaintext secrets" {
    run init_container "$CONTAINER_ALPHA" false
    [ "$status" -eq 0 ]
}

@test "modeB: Initialize beta with plaintext secrets" {
    run init_container "$CONTAINER_BETA" false
    [ "$status" -eq 0 ]
}

@test "modeB: Initialize gamma with plaintext secrets" {
    run init_container "$CONTAINER_GAMMA" false
    [ "$status" -eq 0 ]
}
