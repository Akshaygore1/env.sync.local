#!/bin/bash
# Check if all containers are healthy and services are running

set -e

echo "Checking container health..."

ALL_HEALTHY=1

for container in env-sync-alpha env-sync-beta env-sync-gamma; do
    echo ""
    echo "Checking $container..."
    
    # Check if container is running
    if ! docker ps --format "{{.Names}}" | grep -q "^${container}$"; then
        echo "  ✗ Container is not running"
        ALL_HEALTHY=0
        continue
    fi
    echo "  ✓ Container is running"
    
    # Check SSH daemon
    if docker exec "$container" pgrep -x sshd > /dev/null 2>&1; then
        echo "  ✓ SSH daemon is running"
    else
        echo "  ✗ SSH daemon is NOT running"
        ALL_HEALTHY=0
    fi
    
    # Check Avahi daemon
    if docker exec "$container" pgrep -x avahi-daemon > /dev/null 2>&1; then
        echo "  ✓ Avahi daemon is running"
    else
        echo "  ✗ Avahi daemon is NOT running"
        ALL_HEALTHY=0
    fi
    
    # Check SSH connectivity to other containers
    if [ "$container" = "env-sync-alpha" ]; then
        if docker exec --user envsync "$container" ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no beta.local echo "OK" > /dev/null 2>&1; then
            echo "  ✓ SSH to beta.local: OK"
        else
            echo "  ✗ SSH to beta.local: FAILED"
            ALL_HEALTHY=0
        fi
    fi
done

echo ""
if [ $ALL_HEALTHY -eq 1 ]; then
    echo "✓ All containers are healthy"
    exit 0
else
    echo "✗ Some containers or services are not healthy"
    exit 1
fi
