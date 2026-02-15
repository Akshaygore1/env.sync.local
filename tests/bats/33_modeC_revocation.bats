#!/usr/bin/env bats

# 33_modeC_revocation.bats — Mode C: peer revocation
# Tests that revoking a peer cuts off access

load 'test_helper'

@test "modeC: Add a secret known only to current approved peers" {
    run add_secret "$CONTAINER_ALPHA" "PRE_REVOKE_SECRET" "before-revoke-value"
    [ "$status" -eq 0 ]

    run trigger_sync "$CONTAINER_BETA"
    [ "$status" -eq 0 ]

    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
}

@test "modeC: Verify gamma has the secret before revocation" {
    run get_secret "$CONTAINER_GAMMA" "PRE_REVOKE_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "before-revoke-value" ]
}

@test "modeC: Alpha revokes gamma" {
    peer_output=$(peer_list "$CONTAINER_ALPHA")
    # Peer ID is column 2 (format: "  ✓ PEER_ID (hostname)")
    GAMMA_PEER_ID=$(echo "$peer_output" | grep -i "gamma" | head -1 | awk '{print $2}')

    if [ -n "$GAMMA_PEER_ID" ] && [ "$GAMMA_PEER_ID" != "⏳" ] && [ "$GAMMA_PEER_ID" != "✓" ] && [ "$GAMMA_PEER_ID" != "✗" ]; then
        run peer_revoke "$CONTAINER_ALPHA" "$GAMMA_PEER_ID"
    else
        run peer_revoke "$CONTAINER_ALPHA" "gamma.local"
    fi
    [ "$status" -eq 0 ]
}

@test "modeC: Verify gamma is listed as revoked" {
    run peer_list "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "revoked" ]] || [[ "$output" =~ "gamma" ]]
}

@test "modeC: Add secret after revocation" {
    run add_secret "$CONTAINER_ALPHA" "POST_REVOKE_SECRET" "after-revoke-value"
    [ "$status" -eq 0 ]

    # Sync beta (still approved)
    run trigger_sync "$CONTAINER_BETA"
    [ "$status" -eq 0 ]
}

@test "modeC: Beta (approved) can still sync post-revocation secrets" {
    run get_secret "$CONTAINER_BETA" "POST_REVOKE_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "after-revoke-value" ]
}

@test "modeC: Cleanup — stop mTLS servers" {
    stop_server "$CONTAINER_ALPHA"
    stop_server "$CONTAINER_BETA"
    stop_server "$CONTAINER_GAMMA"
}
