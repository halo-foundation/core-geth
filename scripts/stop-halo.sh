#!/bin/bash
# Stop Halo Chain node
# This script safely stops the running geth node

set -e

echo "🛑 Stopping Halo Chain Node..."
echo "================================"

# Find geth process
GETH_PID=$(pgrep -f "geth.*halo" || true)

if [ -z "$GETH_PID" ]; then
    echo "⚠️  No running Halo Chain node found"
    exit 0
fi

echo "Found Halo Chain node with PID: $GETH_PID"
echo "Sending SIGTERM for graceful shutdown..."

# Send SIGTERM for graceful shutdown
kill -TERM $GETH_PID

# Wait for process to stop (max 30 seconds)
COUNTER=0
while [ $COUNTER -lt 30 ]; do
    if ! ps -p $GETH_PID > /dev/null 2>&1; then
        echo "✅ Halo Chain node stopped successfully"
        exit 0
    fi
    echo "Waiting for node to stop... ($COUNTER/30)"
    sleep 1
    COUNTER=$((COUNTER + 1))
done

# If still running, force kill
if ps -p $GETH_PID > /dev/null 2>&1; then
    echo "⚠️  Node did not stop gracefully, forcing shutdown..."
    kill -9 $GETH_PID
    sleep 1
    if ! ps -p $GETH_PID > /dev/null 2>&1; then
        echo "✅ Halo Chain node force stopped"
    else
        echo "❌ Failed to stop node"
        exit 1
    fi
fi
