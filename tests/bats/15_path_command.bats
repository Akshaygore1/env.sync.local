#!/usr/bin/env bats

load 'test_helper'

@test "env-sync path returns secrets file path" {
    run docker exec "$CONTAINER_ALPHA" env-sync path
    [ "$status" -eq 0 ]
    # Should return absolute path to secrets file
    [[ "$output" =~ ^/.+\.secrets\.env$ ]]
}

@test "env-sync path --backup returns backup directory path" {
    run docker exec "$CONTAINER_ALPHA" env-sync path --backup
    [ "$status" -eq 0 ]
    # Should return absolute path to backup directory
    [[ "$output" =~ ^/.+/backups$ ]]
}

@test "env-sync path output is absolute path" {
    run docker exec "$CONTAINER_ALPHA" env-sync path
    [ "$status" -eq 0 ]
    # Path should start with /
    [[ "$output" =~ ^/ ]]
}

@test "env-sync path can be used in command substitution" {
    # Test that the path command can be used with $(env-sync path)
    run docker exec "$CONTAINER_ALPHA" bash -c 'test -f "$(env-sync path)" && echo "EXISTS"'
    [ "$status" -eq 0 ]
    [ "$output" = "EXISTS" ]
}

@test "env-sync path --help shows usage" {
    run docker exec "$CONTAINER_ALPHA" env-sync path --help
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Usage: env-sync path" ]]
    [[ "$output" =~ "--backup" ]]
}

@test "env-sync path with invalid option fails" {
    run docker exec "$CONTAINER_ALPHA" env-sync path --invalid
    [ "$status" -ne 0 ]
}

# Test bash version if ENV_SYNC_USE_BASH is set
@test "bash version: env-sync path returns secrets file path" {
    if [[ "${ENV_SYNC_USE_BASH}" != "true" ]]; then
        skip "Not testing bash version"
    fi
    run docker exec "$CONTAINER_ALPHA" env-sync path
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^/.+\.secrets\.env$ ]]
}

@test "bash version: env-sync path --backup returns backup directory path" {
    if [[ "${ENV_SYNC_USE_BASH}" != "true" ]]; then
        skip "Not testing bash version"
    fi
    run docker exec "$CONTAINER_ALPHA" env-sync path --backup
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^/.+/backups$ ]]
}
