#!/bin/bash
# Start HDTN DTN node B
set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

exec "$PROJECT_DIR/cmd/dtn-node/dtn-node" --config "$PROJECT_DIR/configs/dtn-node-b.yaml"
