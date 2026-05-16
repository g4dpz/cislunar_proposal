#!/bin/bash
# Stop a running DTN node by sending SIGTERM
set -e

PID_FILE="${1:-/tmp/dtn-node.pid}"

if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    echo "Stopping DTN node (PID $PID)..."
    kill -TERM "$PID" 2>/dev/null || true
    rm -f "$PID_FILE"
else
    # Find by process name
    PID=$(pgrep -f "dtn-node" || true)
    if [ -n "$PID" ]; then
        echo "Stopping DTN node (PID $PID)..."
        kill -TERM $PID
    else
        echo "No running DTN node found"
    fi
fi
