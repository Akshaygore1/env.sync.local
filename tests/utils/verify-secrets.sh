#!/bin/bash
# Verify secrets match across all containers

set -e

KEY="$1"
EXPECTED_VALUE="$2"

if [ -z "$KEY" ] || [ -z "$EXPECTED_VALUE" ]; then
    echo "Usage: $0 <key> <expected_value>"
    exit 1
fi

FAILED=0

echo "Verifying secret '$KEY' across all containers..."
echo "  Expected value: $EXPECTED_VALUE"
echo ""

for container in env-sync-alpha env-sync-beta env-sync-gamma; do
    if ! docker ps --format "{{.Names}}" | grep -q "^${container}$"; then
        echo "  ✗ $container: NOT RUNNING"
        FAILED=1
        continue
    fi
    
    # Get secret value from container
    ACTUAL=$(docker exec "$container" env-sync show "$KEY" 2>/dev/null || echo "")
    
    if [ "$ACTUAL" != "$EXPECTED_VALUE" ]; then
        echo "  ✗ $container: MISMATCH"
        echo "      Expected: '$EXPECTED_VALUE'"
        echo "      Actual:   '$ACTUAL'"
        FAILED=1
    else
        echo "  ✓ $container: OK"
    fi
done

echo ""
if [ $FAILED -eq 0 ]; then
    echo "✓ All containers have matching secret: $KEY"
    exit 0
else
    echo "✗ Secret verification FAILED"
    exit 1
fi
