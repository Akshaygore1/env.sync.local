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
    run docker ps -a --format "{{.Names}}" | grep -E "^env-sync-(alpha|beta|gamma|delta)$"
    [ "$status" -ne 0 ]
}

@test "Clean up Docker volumes" {
    # Remove any lingering volumes
    docker volume rm -f env-sync-test_alpha-data 2>/dev/null || true
    docker volume rm -f env-sync-test_beta-data 2>/dev/null || true
    docker volume rm -f env-sync-test_gamma-data 2>/dev/null || true
    
    # Verify network is removed
    run docker network ls --format "{{.Name}}" | grep "^env-sync-test$"
    [ "$status" -ne 0 ]
}
