#!/usr/bin/env bats

# 32_modeC_sync.bats — Mode C: secure-peer encrypted sync via mTLS
# Tests secret sync between approved peers using mTLS transport

load 'test_helper'

@test "modeC: Add encrypted secret to alpha" {
    run add_secret "$CONTAINER_ALPHA" "MTLS_SECRET" "mtls-value-secure-789"
    [ "$status" -eq 0 ]
}

@test "modeC: Verify secret on alpha" {
    run get_secret "$CONTAINER_ALPHA" "MTLS_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "mtls-value-secure-789" ]
}

@test "modeC: Sync beta (secure peer)" {
    run trigger_sync "$CONTAINER_BETA"
    [ "$status" -eq 0 ]
}

@test "modeC: Verify secret synced to beta via mTLS" {
    run get_secret "$CONTAINER_BETA" "MTLS_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "mtls-value-secure-789" ]
}

@test "modeC: Sync gamma (secure peer)" {
    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
}

@test "modeC: Verify secret synced to gamma via mTLS" {
    run get_secret "$CONTAINER_GAMMA" "MTLS_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "mtls-value-secure-789" ]
}

@test "modeC: Add secret on beta and propagate via secure peer sync" {
    run add_secret "$CONTAINER_BETA" "MTLS_SECOND" "mtls-beta-456"
    [ "$status" -eq 0 ]

    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]

    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]

    run get_secret "$CONTAINER_ALPHA" "MTLS_SECOND"
    [ "$status" -eq 0 ]
    [ "$output" = "mtls-beta-456" ]

    run get_secret "$CONTAINER_GAMMA" "MTLS_SECOND"
    [ "$status" -eq 0 ]
    [ "$output" = "mtls-beta-456" ]
}

@test "modeC: Verify all containers have encrypted files" {
    run parallel_run \
        "container_exec \"$CONTAINER_ALPHA\" grep -q \"ENCRYPTED: true\" /home/envsync/.secrets.env" \
        "container_exec \"$CONTAINER_BETA\" grep -q \"ENCRYPTED: true\" /home/envsync/.secrets.env" \
        "container_exec \"$CONTAINER_GAMMA\" grep -q \"ENCRYPTED: true\" /home/envsync/.secrets.env"
    [ "$status" -eq 0 ]
}
