#!/bin/bash
# Stop ION-DTN node (works for either Node A or Node B)
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ION_BIN="$PROJECT_DIR/ion-build/ion-install/bin"
ION_LIB="$PROJECT_DIR/ion-build/ion-install/lib"

export PATH="$ION_BIN:$PATH"
export DYLD_LIBRARY_PATH="$ION_LIB:$DYLD_LIBRARY_PATH"
export LD_LIBRARY_PATH="$ION_LIB:$LD_LIBRARY_PATH"

echo "=== Stopping ION-DTN node ==="
ionstop
echo "Node stopped."
