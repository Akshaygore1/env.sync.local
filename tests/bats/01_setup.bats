#!/usr/bin/env bats

load 'test_helper'

@test "Docker is installed and running" {
    run check_docker
    [ "$status" -eq 0 ]
}

@test "Docker Compose is installed" {
    run check_docker_compose
    [ "$status" -eq 0 ]
    [ -n "$DOCKER_COMPOSE" ]
}

@test "SSH keys are generated" {
    run "$UTILS_DIR/generate-ssh-keys.sh"
    [ "$status" -eq 0 ]
    [ -f "$DOCKER_DIR/ssh-keys/id_ed25519" ]
    [ -f "$DOCKER_DIR/ssh-keys/authorized_keys" ]
}

@test "Docker image builds successfully" {
    cd "$DOCKER_DIR/.."
    run $DOCKER_COMPOSE -f "$DOCKER_DIR/docker-compose.yml" build --no-cache
    [ "$status" -eq 0 ]
}

@test "Containers start successfully" {
    cd "$DOCKER_DIR/.."
    $DOCKER_COMPOSE -f "$DOCKER_DIR/docker-compose.yml" up -d
    
    run wait_for_containers 90
    [ "$status" -eq 0 ]
}

@test "All containers are healthy" {
    run "$UTILS_DIR/check-health.sh"
    [ "$status" -eq 0 ]
}

@test "Containers can reach each other via mDNS" {
    # Test alpha can resolve beta and gamma
    run container_exec "$CONTAINER_ALPHA" getent hosts beta.local
    [ "$status" -eq 0 ]
    
    run container_exec "$CONTAINER_ALPHA" getent hosts gamma.local
    [ "$status" -eq 0 ]
    
    # Test beta can resolve alpha and gamma
    run container_exec "$CONTAINER_BETA" getent hosts alpha.local
    [ "$status" -eq 0 ]
}

@test "SSH connectivity between containers works" {
    # Test alpha can SSH to beta
    run container_exec "$CONTAINER_ALPHA" ssh -o ConnectTimeout=10 -o StrictHostKeyChecking=no beta.local echo "OK"
    [ "$status" -eq 0 ]
    [ "$output" = "OK" ]
    
    # Test alpha can SSH to gamma
    run container_exec "$CONTAINER_ALPHA" ssh -o ConnectTimeout=10 -o StrictHostKeyChecking=no gamma.local echo "OK"
    [ "$status" -eq 0 ]
    [ "$output" = "OK" ]
}
