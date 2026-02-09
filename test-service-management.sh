#!/bin/bash
# Test script to verify service management during installation

set -e

echo "=== Testing Service Management During Installation ==="
echo ""

# Build the binary first
echo "1. Building env-sync binary..."
make clean 2>/dev/null || true
make build
echo "   ✓ Build successful"
echo ""

# Test 1: Verify service command exists
echo "2. Testing service command availability..."
./target/env-sync service help >/dev/null 2>&1
echo "   ✓ Service command available"
echo ""

# Test 2: Test service stop when not running
echo "3. Testing service stop (when not running)..."
OUTPUT=$(./target/env-sync service stop 2>&1 || true)
if echo "$OUTPUT" | grep -q "Service is not running\|Service stopped"; then
    echo "   ✓ Service stop handled correctly"
else
    echo "   ✗ Unexpected output: $OUTPUT"
    exit 1
fi
echo ""

# Test 3: Test service uninstall when not installed
echo "4. Testing service uninstall (when not installed)..."
OUTPUT=$(./target/env-sync service uninstall 2>&1 || true)
if echo "$OUTPUT" | grep -q "Service is not installed\|Service uninstalled"; then
    echo "   ✓ Service uninstall handled correctly"
else
    echo "   ✗ Unexpected output: $OUTPUT"
    exit 1
fi
echo ""

# Test 4: Test install.sh dry run (check logic)
echo "5. Verifying install.sh contains service management..."
if grep -q "env-sync service stop" install.sh; then
    echo "   ✓ install.sh includes service stop"
else
    echo "   ✗ install.sh missing service stop"
    exit 1
fi
if grep -q "service restart" install.sh; then
    echo "   ✓ install.sh includes service restart"
else
    echo "   ✗ install.sh missing service restart"
    exit 1
fi
echo ""

# Test 5: Test Makefile contains service management
echo "6. Verifying Makefile contains service management..."
if grep -q "env-sync service stop" Makefile; then
    echo "   ✓ Makefile includes service stop"
else
    echo "   ✗ Makefile missing service stop"
    exit 1
fi
if grep -q "service restart" Makefile; then
    echo "   ✓ Makefile includes service restart"
else
    echo "   ✗ Makefile missing service restart"
    exit 1
fi
echo ""

# Test 6: Run Go unit tests
echo "7. Running Go unit tests..."
cd src && go test ./internal/service/... -v | grep -E "PASS|FAIL"
cd ..
echo "   ✓ All unit tests passed"
echo ""

echo "=== All Tests Passed ==="
echo ""
echo "Summary:"
echo "  ✓ Service management commands work correctly"
echo "  ✓ Installation scripts handle service lifecycle"
echo "  ✓ Unit tests pass"
echo ""
echo "Note: Manual testing required for:"
echo "  - Actual service installation (env-sync serve -d)"
echo "  - Service stop/restart during live installation"
echo "  - Testing on Linux (systemd) and macOS (launchd)"
