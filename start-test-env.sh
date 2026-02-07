#!/bin/bash
#
# start-test-env.sh - Start the Docker test environment for manual testing
#
# This script starts the test containers without running tests,
# allowing you to manually explore and debug the environment.
#
# Usage: ./start-test-env.sh
#   --clean     Clean up existing containers and start fresh
#   --stop      Stop the test environment
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTS_DIR="$SCRIPT_DIR/tests"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Parse arguments
CLEAN=0
STOP=0

while [[ $# -gt 0 ]]; do
    case $1 in
        --clean)
            CLEAN=1
            shift
            ;;
        --stop)
            STOP=1
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--clean] [--stop]"
            exit 1
            ;;
    esac
done

# Stop mode
if [ $STOP -eq 1 ]; then
    print_info "Stopping test environment..."
    cd "$TESTS_DIR"
    if command -v docker-compose &> /dev/null; then
        docker-compose -f docker/docker-compose.yml down -v
    else
        docker compose -f docker/docker-compose.yml down -v
    fi
    print_success "Test environment stopped"
    exit 0
fi

# Clean mode - remove existing containers
if [ $CLEAN -eq 1 ]; then
    print_warning "Cleaning up existing containers..."
    cd "$TESTS_DIR"
    if command -v docker-compose &> /dev/null; then
        docker-compose -f docker/docker-compose.yml down -v 2>/dev/null || true
    else
        docker compose -f docker/docker-compose.yml down -v 2>/dev/null || true
    fi
fi

echo ""
echo "=============================================="
echo "Starting env-sync Test Environment"
echo "=============================================="
echo ""

# Generate SSH keys
print_info "Generating SSH keys..."
"$TESTS_DIR/utils/generate-ssh-keys.sh"

# Build and start containers
print_info "Building and starting containers..."
cd "$TESTS_DIR"

if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE="docker compose"
fi

$DOCKER_COMPOSE -f docker/docker-compose.yml up -d --build

# Wait for containers
print_info "Waiting for containers to be ready (this may take a moment)..."
sleep 10

# Check health
print_info "Checking container health..."
"$TESTS_DIR/utils/check-health.sh"

echo ""
print_success "Test environment is ready!"
echo ""
echo "Containers:"
echo "  - alpha (env-sync-alpha):    SSH/mDNS ready"
echo "  - beta  (env-sync-beta):     SSH/mDNS ready"
echo "  - gamma (env-sync-gamma):    SSH/mDNS ready"
echo ""
echo "Quick commands:"
echo "  docker exec -it env-sync-alpha bash    # Access alpha"
echo "  docker exec -it env-sync-beta bash     # Access beta"
echo "  docker exec -it env-sync-gamma bash    # Access gamma"
echo ""
echo "Example workflow:"
echo "  # Initialize alpha with encrypted secrets"
echo "  docker exec env-sync-alpha env-sync init --encrypted"
echo "  docker exec env-sync-alpha env-sync add MY_SECRET='test-value'"
echo ""
echo "  # Sync to beta"
echo "  docker exec env-sync-beta env-sync init --encrypted"
echo "  docker exec env-sync-beta env-sync --force"
echo ""
echo "  # View logs"
echo "  docker logs env-sync-alpha -f"
echo ""
echo "To stop the environment:"
echo "  ./start-test-env.sh --stop"
echo ""
