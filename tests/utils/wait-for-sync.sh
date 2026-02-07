#!/bin/bash
# Wait for env-sync to complete on a container

set -e

CONTAINER="$1"
TIMEOUT="${2:-20}"

if [ -z "$CONTAINER" ]; then
    echo "Usage: $0 <container_name> [timeout_seconds]"
    exit 1
fi

echo "Waiting for sync to complete on $CONTAINER (timeout: ${TIMEOUT}s)..."

# Wait for the sync log to indicate completion
START_TIME=$(date +%s)
while true; do
    CURRENT_TIME=$(date +%s)
    ELAPSED=$((CURRENT_TIME - START_TIME))

    if [ $ELAPSED -ge $TIMEOUT ]; then
        echo "  ✗ TIMEOUT after ${TIMEOUT} seconds"
        exit 1
    fi

    # Check if log file exists and has completion marker
    if docker exec --user envsync "$CONTAINER" test -f /home/envsync/.config/env-sync/logs/env-sync.log 2>/dev/null; then
        LOG_CONTENT=$(docker exec --user envsync "$CONTAINER" cat /home/envsync/.config/env-sync/logs/env-sync.log 2>/dev/null || echo "")

        if echo "$LOG_CONTENT" | grep -qE "sync.*completed|Sync.*complete"; then
            echo "  ✓ Sync completed on $CONTAINER (${ELAPSED}s)"
            exit 0
        fi

        if echo "$LOG_CONTENT" | grep -qE "sync.*failed|Sync.*failed|ERROR"; then
            echo "  ✗ Sync FAILED on $CONTAINER"
            echo "  Log excerpt:"
            echo "$LOG_CONTENT" | tail -5 | sed 's/^/    /'
            exit 1
        fi
    fi

    sleep 1
done
