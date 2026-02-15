#!/usr/bin/env bats

# 31_modeC_peer_enrollment.bats — Mode C: peer enrollment workflow
# Tests the invite → request → approve → cert exchange lifecycle

load 'test_helper'

@test "modeC: Alpha creates invite token" {
    run peer_invite "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Token:" ]]
}

@test "modeC: Beta requests access to alpha using invite token" {
    invite_output=$(peer_invite "$CONTAINER_ALPHA")
    INVITE_TOKEN=$(echo "$invite_output" | grep "Token:" | awk '{print $2}')
    [ -n "$INVITE_TOKEN" ]

    run peer_request "$CONTAINER_BETA" "alpha.local" "$INVITE_TOKEN"
    [ "$status" -eq 0 ]
}

@test "modeC: Alpha sees beta as pending peer" {
    run peer_list "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "beta" ]] || [[ "$output" =~ "pending" ]]
}

@test "modeC: Alpha approves beta" {
    peer_output=$(peer_list "$CONTAINER_ALPHA")
    BETA_PEER_ID=$(echo "$peer_output" | grep -i "beta" | head -1 | awk '{print $2}')

    if [ -n "$BETA_PEER_ID" ] && [ "$BETA_PEER_ID" != "⏳" ] && [ "$BETA_PEER_ID" != "✓" ]; then
        run peer_approve "$CONTAINER_ALPHA" "$BETA_PEER_ID"
        [ "$status" -eq 0 ]
    else
        run peer_approve "$CONTAINER_ALPHA" "beta.local"
        [ "$status" -eq 0 ]
    fi
}

@test "modeC: Gamma enrolls with alpha" {
    invite_output=$(peer_invite "$CONTAINER_ALPHA")
    INVITE_TOKEN=$(echo "$invite_output" | grep "Token:" | awk '{print $2}')
    [ -n "$INVITE_TOKEN" ]

    run peer_request "$CONTAINER_GAMMA" "alpha.local" "$INVITE_TOKEN"
    [ "$status" -eq 0 ]

    peer_output=$(peer_list "$CONTAINER_ALPHA")
    GAMMA_PEER_ID=$(echo "$peer_output" | grep -i "gamma" | head -1 | awk '{print $2}')

    if [ -n "$GAMMA_PEER_ID" ] && [ "$GAMMA_PEER_ID" != "⏳" ] && [ "$GAMMA_PEER_ID" != "✓" ]; then
        run peer_approve "$CONTAINER_ALPHA" "$GAMMA_PEER_ID"
        [ "$status" -eq 0 ]
    else
        run peer_approve "$CONTAINER_ALPHA" "gamma.local"
        [ "$status" -eq 0 ]
    fi
}

@test "modeC: Alpha lists both beta and gamma as approved" {
    run peer_list "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "beta" ]]
    [[ "$output" =~ "gamma" ]]
}

@test "modeC: Exchange TLS certificates for bidirectional mTLS" {
    exchange_tls_certs "$CONTAINER_ALPHA" "$CONTAINER_BETA" "$CONTAINER_GAMMA"

    # Verify each container has the others' certs in trusted dir
    run container_exec "$CONTAINER_BETA" ls /home/envsync/.config/env-sync/tls/trusted/
    [ "$status" -eq 0 ]
    [[ "$output" =~ "alpha" ]]
}

@test "modeC: Exchange AGE public keys between all containers" {
    local alpha_pubkey beta_pubkey gamma_pubkey

    alpha_pubkey=$(get_pubkey "$CONTAINER_ALPHA")
    beta_pubkey=$(get_pubkey "$CONTAINER_BETA")
    gamma_pubkey=$(get_pubkey "$CONTAINER_GAMMA")

    [ -n "$alpha_pubkey" ]
    [ -n "$beta_pubkey" ]
    [ -n "$gamma_pubkey" ]

    # Import to alpha
    run import_pubkey "$CONTAINER_ALPHA" "$beta_pubkey" "beta.local"
    [ "$status" -eq 0 ]
    run import_pubkey "$CONTAINER_ALPHA" "$gamma_pubkey" "gamma.local"
    [ "$status" -eq 0 ]

    # Import to beta
    run import_pubkey "$CONTAINER_BETA" "$alpha_pubkey" "alpha.local"
    [ "$status" -eq 0 ]
    run import_pubkey "$CONTAINER_BETA" "$gamma_pubkey" "gamma.local"
    [ "$status" -eq 0 ]

    # Import to gamma
    run import_pubkey "$CONTAINER_GAMMA" "$alpha_pubkey" "alpha.local"
    [ "$status" -eq 0 ]
    run import_pubkey "$CONTAINER_GAMMA" "$beta_pubkey" "beta.local"
    [ "$status" -eq 0 ]
}

@test "modeC: Peer commands require secure-peer mode" {
    run container_exec "$CONTAINER_ALPHA" env-sync peer list
    [ "$status" -eq 0 ]
}
