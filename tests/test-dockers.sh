#!/bin/bash
#
# test-dockers.sh - Main entry point for Docker-based integration tests
#
# This script sets up bats-core and runs all Docker-based integration tests
# to verify env-sync functionality across multiple containers.
#
# Usage: ./tests/test-dockers.sh [options]
#   --no-cleanup    Keep containers running after tests (for debugging)
#   --setup-only    Only setup the test environment, don't run tests
#   --debug         Enable debug mode (print outputs of failures)
#   --filter PATTERN Run only tests matching the pattern
#   --mode MODE     Run only tests for a specific mode (A, B, C, or all) [default: all]
#   --formatter FMT  Output format (pretty, tap, junit, etc.) [default: pretty]
#   --help          Show this help message
#

set -e

# Script directory
TESTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_DIR="$TESTS_DIR/.."
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
FORMATTER="pretty"
DEBUG_MODE=0
SHOW_HELP=0
TEST_MODE="all"

# Default to tap in any CI environment
if [ "${CI:-}" = "true" ]; then
    FORMATTER="tap"
fi

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
        --debug)
            DEBUG_MODE=1
            shift
            ;;
        --filter)
            FILTER="$2"
            shift 2
            ;;
        --mode)
            TEST_MODE="$2"
            shift 2
            ;;
        --formatter)
            FORMATTER="$2"
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
    echo "env-sync Docker Integration Tests (v3.0 — 3-mode support)"
    echo ""
    echo "Usage: ./tests/test-dockers.sh [options]"
    echo ""
    echo "Options:"
    echo "  --no-cleanup      Keep containers running after tests (for debugging)"
    echo "  --setup-only      Only setup the test environment, don't run tests"
    echo "  --filter PATTERN  Run only tests matching the pattern"
    echo "  --mode MODE       Run tests for a specific mode:"
    echo "                      A    dev-plaintext-http tests (10-12)"
    echo "                      B    trusted-owner-ssh tests (20-25)"
    echo "                      C    secure-peer tests (30-33)"
    echo "                      all  All tests (default)"
    echo "  --formatter FMT   Output format (pretty, tap, junit, etc.) [default: $FORMATTER]"
    echo "  --debug           Print output on test failures"
    echo "  --help, -h        Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./tests/test-dockers.sh                       # Run all tests"
    echo "  ./tests/test-dockers.sh --mode A              # Run only Mode A (HTTP) tests"
    echo "  ./tests/test-dockers.sh --mode B              # Run only Mode B (SSH) tests"
    echo "  ./tests/test-dockers.sh --mode C              # Run only Mode C (mTLS) tests"
    echo "  ./tests/test-dockers.sh --no-cleanup          # Run tests, keep containers"
    echo "  ./tests/test-dockers.sh --filter modeA        # Run only modeA tests"
    echo "  ./tests/test-dockers.sh --formatter tap       # Run tests with TAP output"
    echo "  ./tests/test-dockers.sh --setup-only          # Just setup, then exit"
    echo ""
    exit 0
fi

# Set up binary paths
ENV_SYNC_GO_BIN="${ENV_SYNC_GO_BIN:-target/env-sync}"
ENV_SYNC_BIN="$ENV_SYNC_GO_BIN"

export ENV_SYNC_GO_BIN
export ENV_SYNC_BIN

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

# Function to check for bats
check_bats() {
    print_info "Checking for bats testing framework..."

    if command -v bats &> /dev/null; then
        print_success "bats is installed"
        export BATS_BIN="bats"
    else
        print_error "bats is not installed. Please install bats-core on your system."
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
}

# Function to build Go binary with Docker
build_go_binary() {
    local go_image="${ENV_SYNC_GO_DOCKER_IMAGE:-golang:1.24}"

    print_info "Building Go binary with Docker (${go_image})..."
    mkdir -p "$SCRIPT_DIR/target"

    docker run --rm \
        -u "$(id -u):$(id -g)" \
        -e CGO_ENABLED=0 \
        -e HOME=/tmp \
        -e GOPATH=/tmp/go \
        -e GOCACHE=/tmp/go-cache \
        -e GOMODCACHE=/tmp/go-mod \
        -v "$SCRIPT_DIR":/workspace \
        -w /workspace/src \
        "$go_image" \
        bash -c 'mkdir -p /workspace/target && /usr/local/go/bin/go build -o /workspace/target/env-sync ./cmd/env-sync'
}

# Function to setup test environment
setup_environment() {
    print_info "Setting up test environment..."

    build_go_binary

    # Generate SSH keys
    print_info "Generating SSH keys..."
    "$TESTS_DIR/utils/generate-ssh-keys.sh"

    # Build Docker image
    print_info "Building Docker image..."
    cd "$TESTS_DIR"
    $DOCKER_COMPOSE -f "$TESTS_DIR/docker/docker-compose.yml" build

    print_success "Test environment is ready"
}

# Function to select test files based on mode
select_test_files() {
    local mode="$1"
    local test_files=""

    # Always include setup
    test_files="$BATS_TEST_DIR/01_setup.bats"

    case "$mode" in
        A|a|modeA)
            print_info "Running Mode A (dev-plaintext-http) tests" >&2
            test_files="$test_files $BATS_TEST_DIR/10_modeA_setup.bats"
            test_files="$test_files $BATS_TEST_DIR/11_modeA_basic_sync.bats"
            test_files="$test_files $BATS_TEST_DIR/12_modeA_force_pull.bats"
            ;;
        B|b|modeB)
            print_info "Running Mode B (trusted-owner-ssh) tests" >&2
            test_files="$test_files $BATS_TEST_DIR/20_modeB_setup.bats"
            test_files="$test_files $BATS_TEST_DIR/21_modeB_basic_sync.bats"
            test_files="$test_files $BATS_TEST_DIR/22_modeB_encrypted_sync.bats"
            test_files="$test_files $BATS_TEST_DIR/23_modeB_propagation.bats"
            test_files="$test_files $BATS_TEST_DIR/24_modeB_add_machine.bats"
            test_files="$test_files $BATS_TEST_DIR/25_modeB_force_pull.bats"
            ;;
        C|c|modeC)
            print_info "Running Mode C (secure-peer) tests" >&2
            test_files="$test_files $BATS_TEST_DIR/30_modeC_setup.bats"
            test_files="$test_files $BATS_TEST_DIR/31_modeC_peer_enrollment.bats"
            test_files="$test_files $BATS_TEST_DIR/32_modeC_sync.bats"
            test_files="$test_files $BATS_TEST_DIR/33_modeC_revocation.bats"
            ;;
        all)
            print_info "Running all mode tests (A, B, C)" >&2
            test_files="$BATS_TEST_DIR"
            ;;
        *)
            print_error "Unknown mode: $mode. Use A, B, C, or all."
            exit 1
            ;;
    esac

    # Append teardown unless running all (bats handles sorting)
    if [ "$mode" != "all" ]; then
        test_files="$test_files $BATS_TEST_DIR/99_teardown.bats"
    fi

    echo "$test_files"
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
echo "env-sync Docker Integration Tests (v3.0)"
echo "=============================================="
echo ""

# Check prerequisites
check_prerequisites

# Check for bats
check_bats

# Setup environment
setup_environment

# If setup-only, exit here
if [ $SETUP_ONLY -eq 1 ]; then
    print_info "Setup complete (--setup-only mode)"
    print_info "Test environment is ready. Containers are not running."
    print_info "To run tests, execute: ./tests/test-dockers.sh"
    exit 0
fi

# Set trap for cleanup on exit
trap cleanup EXIT

# Select test files based on mode
BATS_TEST_PATTERN=$(select_test_files "$TEST_MODE")

# Run tests
echo ""
print_info "Running tests (mode: $TEST_MODE)..."
echo ""

BATS_ARGS="--timing --formatter $FORMATTER"

if [ -n "$FILTER" ]; then
    BATS_ARGS="$BATS_ARGS --filter '$FILTER'"
fi

if [ $DEBUG_MODE -eq 1 ]; then
    BATS_ARGS="$BATS_ARGS --print-output-on-failure"
fi

cd "$SCRIPT_DIR"

# Run bats tests
$BATS_BIN $BATS_ARGS $BATS_TEST_PATTERN

TEST_EXIT_CODE=$?

echo ""
echo "=============================================="
if [ $TEST_EXIT_CODE -eq 0 ]; then
    print_success "All tests passed! (mode: $TEST_MODE)"
else
    print_error "Some tests failed! (mode: $TEST_MODE)"
fi
echo "=============================================="
echo ""

exit $TEST_EXIT_CODE
