#!/usr/bin/env bats

load 'test_helper'

@test "Clear existing secrets and initialize alpha with encryption" {
    run init_container "$CONTAINER_ALPHA" true
    [ "$status" -eq 0 ]
}

@test "Add encrypted secret to alpha" {
    run add_secret "$CONTAINER_ALPHA" "ENCRYPTED_SECRET" "secret-value-789"
    [ "$status" -eq 0 ]
}

@test "Verify encrypted file exists on alpha" {
    run container_exec "$CONTAINER_ALPHA" test -f /home/envsync/.secrets.env
    [ "$status" -eq 0 ]
    
    run container_exec "$CONTAINER_ALPHA" grep -q "ENCRYPTED: true" /home/envsync/.secrets.env
    [ "$status" -eq 0 ]
}

@test "Get alpha's AGE public key" {
    ALPHA_PUBKEY=$(get_pubkey "$CONTAINER_ALPHA")
    [ -n "$ALPHA_PUBKEY" ]
    echo "Alpha pubkey: $ALPHA_PUBKEY" >&3
    export ALPHA_PUBKEY
}

@test "Initialize beta with encryption" {
    run init_container "$CONTAINER_BETA" true
    [ "$status" -eq 0 ]
}

@test "Initialize gamma with encryption" {
    run init_container "$CONTAINER_GAMMA" true
    [ "$status" -eq 0 ]
}

@test "Get beta's and gamma's AGE public keys" {
    BETA_PUBKEY=$(get_pubkey "$CONTAINER_BETA")
    [ -n "$BETA_PUBKEY" ]
    export BETA_PUBKEY
    echo "Beta pubkey: $BETA_PUBKEY" >&3
    
    GAMMA_PUBKEY=$(get_pubkey "$CONTAINER_GAMMA")
    [ -n "$GAMMA_PUBKEY" ]
    export GAMMA_PUBKEY
    echo "Gamma pubkey: $GAMMA_PUBKEY" >&3
}

@test "Exchange public keys between all containers" {
    # Import keys to alpha
    run import_pubkey "$CONTAINER_ALPHA" "$(get_pubkey "$CONTAINER_BETA")" "beta.local"
    [ "$status" -eq 0 ]
    
    run import_pubkey "$CONTAINER_ALPHA" "$(get_pubkey "$CONTAINER_GAMMA")" "gamma.local"
    [ "$status" -eq 0 ]
    
    # Import keys to beta
    run import_pubkey "$CONTAINER_BETA" "$(get_pubkey "$CONTAINER_ALPHA")" "alpha.local"
    [ "$status" -eq 0 ]
    
    run import_pubkey "$CONTAINER_BETA" "$(get_pubkey "$CONTAINER_GAMMA")" "gamma.local"
    [ "$status" -eq 0 ]
    
    # Import keys to gamma
    run import_pubkey "$CONTAINER_GAMMA" "$(get_pubkey "$CONTAINER_ALPHA")" "alpha.local"
    [ "$status" -eq 0 ]
    
    run import_pubkey "$CONTAINER_GAMMA" "$(get_pubkey "$CONTAINER_BETA")" "beta.local"
    [ "$status" -eq 0 ]
}

@test "Trigger sync on all containers to exchange encrypted secrets" {
    run trigger_sync "$CONTAINER_ALPHA"
    [ "$status" -eq 0 ]
    
    run trigger_sync "$CONTAINER_BETA"
    [ "$status" -eq 0 ]
    
    run trigger_sync "$CONTAINER_GAMMA"
    [ "$status" -eq 0 ]
}

@test "Verify beta can decrypt the secret" {
    run get_secret "$CONTAINER_BETA" "ENCRYPTED_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "secret-value-789" ]
}

@test "Verify gamma can decrypt the secret" {
    run get_secret "$CONTAINER_GAMMA" "ENCRYPTED_SECRET"
    [ "$status" -eq 0 ]
    [ "$output" = "secret-value-789" ]
}

@test "Verify all containers have encrypted files" {
    for container in "$CONTAINER_ALPHA" "$CONTAINER_BETA" "$CONTAINER_GAMMA"; do
        run container_exec "$container" grep -q "ENCRYPTED: true" /home/envsync/.secrets.env
        [ "$status" -eq 0 ]
    done
}
