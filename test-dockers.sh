#!/bin/bash
#
# test-dockers.sh - Main entry point for Docker-based integration tests
#
# This script sets up bats-core and runs all Docker-based integration tests
# to verify env-sync functionality across multiple containers.
#
# Usage: ./test-dockers.sh [options]
#   --no-cleanup    Keep containers running after tests (for debugging)
#   --setup-only    Only setup the test environment, don't run tests
#   --filter PATTERN Run only tests matching the pattern
#   --help          Show this help message
#

set -e

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTS_DIR="$SCRIPT_DIR/tests"
BATS_DIR="$TESTS_DIR/bats-core"
BATS_TEST_DIR="$TESTS_DIR/bats"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse arguments
NO_CLEANUP=0
SETUP_ONLY=0
FILTER=""
SHOW_HELP=0

while [[ $# -gt 0 ]]; do
    case $1 in
        --no-cleanup)
            NO_CLEANUP=1
            shift
            ;;
        --setup-only)
            SETUP_ONLY=1
            shift
            ;;
        --filter)
            FILTER="$2"
            shift 2
            ;;
        --help|-h)
            SHOW_HELP=1
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Show help
if [ $SHOW_HELP -eq 1 ]; then
    echo "env-sync Docker Integration Tests"
    echo ""
    echo "Usage: ./test-dockers.sh [options]"
    echo ""
    echo "Options:"
    echo "  --no-cleanup      Keep containers running after tests (for debugging)"
    echo "  --setup-only      Only setup the test environment, don't run tests"
    echo "  --filter PATTERN  Run only tests matching the pattern"
    echo "  --help, -h        Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./test-dockers.sh                                    # Run all tests"
    echo "  ./test-dockers.sh --no-cleanup                       # Run tests, keep containers"
    echo "  ./test-dockers.sh --filter basic                     # Run only basic sync tests"
    echo "  ./test-dockers.sh --setup-only                       # Just setup, then exit"
    echo ""
    exit 0
fi

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Function to setup bats-core
setup_bats() {
    print_info "Setting up bats-core testing framework..."
    
    if [ -d "$BATS_DIR" ]; then
        print_info "bats-core already exists, updating..."
        cd "$BATS_DIR"
        git pull --quiet 2>/dev/null || true
    else
        print_info "Cloning bats-core repository..."
        git clone --depth 1 --quiet https://github.com/bats-core/bats-core.git "$BATS_DIR"
    fi
    
    # Verify bats is available
    if [ -f "$BATS_DIR/bin/bats" ]; then
        print_success "bats-core is ready"
        export BATS_BIN="$BATS_DIR/bin/bats"
    else
        print_error "Failed to setup bats-core"
        exit 1
    fi
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        exit 1
    fi
    print_success "Docker is available"
    
    # Check docker-compose
    if command -v docker-compose &> /dev/null; then
        export DOCKER_COMPOSE="docker-compose"
    elif docker compose version &> /dev/null; then
        export DOCKER_COMPOSE="docker compose"
    else
        print_error "docker-compose is not installed"
        exit 1
    fi
    print_success "Docker Compose is available"
    
    # Check git (for bats-core)
    if ! command -v git &> /dev/null; then
        print_error "git is required to fetch bats-core"
        exit 1
    fi
    print_success "git is available"
}

# Function to setup test environment
setup_environment() {
    print_info "Setting up test environment..."
    
    # Generate SSH keys
    print_info "Generating SSH keys..."
    "$TESTS_DIR/utils/generate-ssh-keys.sh"
    
    # Build Docker image
    print_info "Building Docker image..."
    cd "$TESTS_DIR"
    $DOCKER_COMPOSE -f "$TESTS_DIR/docker/docker-compose.yml" build
    
    print_success "Test environment is ready"
}

# Function to cleanup
cleanup() {
    if [ $NO_CLEANUP -eq 1 ]; then
        print_warning "Skipping cleanup (--no-cleanup flag set)"
        print_info "Containers are still running. To clean up later, run:"
        print_info "  cd tests && docker-compose -f docker/docker-compose.yml down -v"
        return
    fi
    
    print_info "Cleaning up..."
    cd "$TESTS_DIR"
    $DOCKER_COMPOSE -f "$TESTS_DIR/docker/docker-compose.yml" down -v 2>/dev/null || true
    docker stop env-sync-delta 2>/dev/null || true
    docker rm env-sync-delta 2>/dev/null || true
    print_success "Cleanup complete"
}

# Main execution
echo ""
echo "=============================================="
echo "env-sync Docker Integration Tests"
echo "=============================================="
echo ""

# Check prerequisites
check_prerequisites

# Setup bats-core
setup_bats

# Setup environment
setup_environment

# If setup-only, exit here
if [ $SETUP_ONLY -eq 1 ]; then
    print_info "Setup complete (--setup-only mode)"
    print_info "Test environment is ready. Containers are not running."
    print_info "To run tests, execute: ./test-dockers.sh"
    exit 0
fi

# Set trap for cleanup on exit
trap cleanup EXIT

# Run tests
echo ""
print_info "Running tests..."
echo ""

BATS_ARGS="--timing"

if [ -n "$FILTER" ]; then
    BATS_ARGS="$BATS_ARGS --filter '$FILTER'"
fi

if [ $NO_CLEANUP -eq 1 ]; then
    # Skip teardown if --no-cleanup is set
    BATS_TEST_PATTERN="01_setup.bats 10_basic_sync.bats 20_encrypted_sync.bats 30_propagation.bats 40_add_machine.bats"
else
    BATS_TEST_PATTERN="$BATS_TEST_DIR"
fi

cd "$SCRIPT_DIR"

# Run bats tests
$BATS_BIN $BATS_ARGS $BATS_TEST_PATTERN

TEST_EXIT_CODE=$?

echo ""
echo "=============================================="
if [ $TEST_EXIT_CODE -eq 0 ]; then
    print_success "All tests passed!"
else
    print_error "Some tests failed!"
fi
echo "=============================================="
echo ""

exit $TEST_EXIT_CODE
