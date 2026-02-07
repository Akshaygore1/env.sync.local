#!/usr/bin/env bats

load 'test_helper'

@test "Clean up all test containers and volumes" {
    cd "$DOCKER_DIR/.."
    
    # Stop and remove containers via docker-compose
    $DOCKER_COMPOSE -f "$DOCKER_DIR/docker-compose.yml" down -v 2>/dev/null || true
    
    # Remove delta if it still exists
    docker stop "$CONTAINER_DELTA" 2>/dev/null || true
    docker rm "$CONTAINER_DELTA" 2>/dev/null || true
    
    # Verify cleanup
    run wait_for_containers_removed 60 "$CONTAINER_ALPHA" "$CONTAINER_BETA" "$CONTAINER_GAMMA" "$CONTAINER_DELTA"
    [ "$status" -eq 0 ]
}

@test "Clean up Docker volumes" {
    # Remove any lingering volumes
    docker volume rm -f env-sync-test_alpha-data 2>/dev/null || true
    docker volume rm -f env-sync-test_beta-data 2>/dev/null || true
    docker volume rm -f env-sync-test_gamma-data 2>/dev/null || true
    
    # Verify network is removed
    run wait_for_network_removed "env-sync-test" 60
    [ "$status" -eq 0 ]
}
