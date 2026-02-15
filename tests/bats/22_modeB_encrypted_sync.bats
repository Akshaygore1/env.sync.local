#!/usr/bin/env bats

# 22_modeB_encrypted_sync.bats — Mode B: encrypted sync via SSH
# Tests AGE-encrypted sync with key exchange between containers

load 'test_helper'

@test "modeB: Re-initialize alpha with encryption" {
    run init_container "$CONTAINER_ALPHA" true
    [ "$status" -eq 0 ]
}

@test "modeB: Add encrypted secret to alpha" {
    run add_secret "$CONTAINER_ALPHA" "ENCRYPTED_SECRET" "secret-value-789"
    [ "$status" -eq 0 ]
}

@test "modeB: Verify encrypted file exists on alpha" {
    run container_exec "$CONTAINER_ALPHA" grep -q "ENCRYPTED: true" /home/envsync/.secrets.env
    [ "$status" -eq 0 ]
}

@test "modeB: Initialize beta and gamma with encryption" {
    run init_container "$CONTAINER_BETA" true
    [ "$status" -eq 0 ]

    run init_container "$CONTAINER_GAMMA" true
    [ "$status" -eq 0 ]
}

@test "modeB: Exchange public keys between all containers" {
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

@test "modeB: Sync all containers to exchange encrypted secrets" {
    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]

    run trigger_sync "$CONTAINER_BETA"
    [ "$status" -eq 0 ]

    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
}

@test "modeB: Verify beta can decrypt the secret" {
    run get_secret "$CONTAINER_BETA" "ENCRYPTED_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "secret-value-789" ]
}

@test "modeB: Verify gamma can decrypt the secret" {
    run get_secret "$CONTAINER_GAMMA" "ENCRYPTED_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "secret-value-789" ]
}

@test "modeB: Verify all containers have encrypted files" {
    run parallel_run \
        "container_exec \"$CONTAINER_ALPHA\" grep -q \"ENCRYPTED: true\" /home/envsync/.secrets.env" \
        "container_exec \"$CONTAINER_BETA\" grep -q \"ENCRYPTED: true\" /home/envsync/.secrets.env" \
        "container_exec \"$CONTAINER_GAMMA\" grep -q \"ENCRYPTED: true\" /home/envsync/.secrets.env"
    [ "$status" -eq 0 ]
}
