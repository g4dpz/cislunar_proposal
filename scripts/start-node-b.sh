#!/bin/bash
# Start ION-DTN Node B (Engine 2) with KISS CLA
# Run from the project root directory.
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ION_BIN="$PROJECT_DIR/ion-build/ion-install/bin"
ION_LIB="$PROJECT_DIR/ion-build/ion-install/lib"
CONFIG_DIR="$PROJECT_DIR/configs/node-b"

export PATH="$ION_BIN:$PATH"
export DYLD_LIBRARY_PATH="$ION_LIB:$DYLD_LIBRARY_PATH"
export LD_LIBRARY_PATH="$ION_LIB:$LD_LIBRARY_PATH"

echo "=== Starting ION-DTN Node B (Engine 2) ==="
echo "Config: $CONFIG_DIR"
echo "ION bin: $ION_BIN"

# Copy kiss.ionconfig to working directory
cp "$CONFIG_DIR/kiss.ionconfig" .

# Initialize ION
echo "--- ionadmin ---"
ionadmin "$CONFIG_DIR/node.ionrc"

# Initialize LTP (with KISS CLA)
echo "--- ltpadmin ---"
ltpadmin "$CONFIG_DIR/node.ltprc"

# Initialize BP
echo "--- bpadmin ---"
bpadmin "$CONFIG_DIR/node.bprc"

# Initialize IPN routing
echo "--- ipnadmin ---"
ipnadmin "$CONFIG_DIR/node.ipnrc"

echo ""
echo "=== Node B (Engine 2) is running ==="
echo "Endpoints: ipn:2.0, ipn:2.1, ipn:2.2"
echo "KISS CLA: ltpkisscli 2 / ltpkissclo 1"
echo ""
echo "Test with: bping ipn:2.1 ipn:1.1"
echo "Stop with: ionstop"
