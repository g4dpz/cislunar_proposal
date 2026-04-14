#!/bin/bash
# Test ION-DTN store-and-forward over KISS CLA (Task 9)
# This script validates store-and-forward functionality between two nodes.
#
# PREREQUISITES:
# - Two nodes (Node A and Node B) with ION-DTN installed
# - TNC4 hardware connected via USB on both nodes
# - FT-817 radios configured for 9600 baud operation
# - Both nodes must run this test from separate terminals/machines
#
# USAGE:
#   On Node A: ./scripts/test-ion-store-forward.sh node-a
#   On Node B: ./scripts/test-ion-store-forward.sh node-b

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ION_BIN="$PROJECT_DIR/ion-install/bin"
ION_LIB="$PROJECT_DIR/ion-install/lib"

export PATH="$ION_BIN:$PATH"
export DYLD_LIBRARY_PATH="$ION_LIB:$DYLD_LIBRARY_PATH"
export LD_LIBRARY_PATH="$ION_LIB:$LD_LIBRARY_PATH"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse arguments
NODE_TYPE="${1:-}"

if [[ "$NODE_TYPE" != "node-a" && "$NODE_TYPE" != "node-b" ]]; then
    echo "Usage: $0 <node-a|node-b>"
    echo ""
    echo "This script performs Task 9 testing: ION-DTN store-and-forward over KISS CLA"
    echo ""
    echo "Run on Node A:"
    echo "  $0 node-a"
    echo ""
    echo "Run on Node B:"
    echo "  $0 node-b"
    exit 1
fi

# Set node-specific variables
if [[ "$NODE_TYPE" == "node-a" ]]; then
    NODE_NAME="Node A (Engine 1)"
    LOCAL_EID="ipn:1.1"
    REMOTE_EID="ipn:2.1"
    ENGINE_ID="1"
    REMOTE_ENGINE="2"
    STARTUP_SCRIPT="$SCRIPT_DIR/start-node-a.sh"
else
    NODE_NAME="Node B (Engine 2)"
    LOCAL_EID="ipn:2.1"
    REMOTE_EID="ipn:1.1"
    ENGINE_ID="2"
    REMOTE_ENGINE="1"
    STARTUP_SCRIPT="$SCRIPT_DIR/start-node-b.sh"
fi

echo "========================================"
echo "ION-DTN Store-and-Forward Test - Task 9"
echo "========================================"
echo "Node: $NODE_NAME"
echo "Local EID: $LOCAL_EID"
echo "Remote EID: $REMOTE_EID"
echo ""

# Start ION-DTN and verify initialization
echo "=== Starting ION-DTN on $NODE_NAME ==="
echo ""

# Check if ION is already running
if pgrep -f "rfxclock" > /dev/null 2>&1; then
    echo -e "${YELLOW}Warning: ION-DTN appears to be already running${NC}"
    echo "Stopping existing instance..."
    ionstop 2>/dev/null || true
    sleep 2
fi

# Start the node
echo "Starting $NODE_NAME..."
bash "$STARTUP_SCRIPT"
echo ""

# Wait for ION to initialize
echo "Waiting for ION-DTN to initialize (5 seconds)..."
sleep 5

# Verify ION-DTN processes are running
echo ""
echo "=== Verifying ION-DTN processes ==="
PROCESSES_OK=true

# Check for rfxclock (ION core process)
if pgrep -f "rfxclock" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} rfxclock is running"
else
    echo -e "${RED}✗${NC} rfxclock is NOT running"
    PROCESSES_OK=false
fi

# Check for ltpkisscli (LTP KISS receive)
if pgrep -f "ltpkisscli $ENGINE_ID" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} ltpkisscli $ENGINE_ID is running"
else
    echo -e "${RED}✗${NC} ltpkisscli $ENGINE_ID is NOT running"
    PROCESSES_OK=false
fi

# Check for ltpkissclo (LTP KISS transmit)
if pgrep -f "ltpkissclo $REMOTE_ENGINE" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} ltpkissclo $REMOTE_ENGINE is running"
else
    echo -e "${RED}✗${NC} ltpkissclo $REMOTE_ENGINE is NOT running"
    PROCESSES_OK=false
fi

echo ""
if [[ "$PROCESSES_OK" == true ]]; then
    echo -e "${GREEN}✓ ION-DTN initialized successfully${NC}"
else
    echo -e "${RED}✗ Some ION-DTN processes are not running${NC}"
    exit 1
fi

# Main test menu
while true; do
    echo ""
    echo "========================================"
    echo "Task 9 Test Menu"
    echo "========================================"
    echo "1) Task 9.1: Test bpsendfile / bprecvfile"
    echo "2) Task 9.2: Test store-and-forward with delayed contact"
    echo "3) Task 9.3: Test priority-based delivery"
    echo "4) Task 9.4: Test bundle lifetime expiry"
    echo "5) Run all tests"
    echo "6) Exit and stop ION-DTN"
    echo ""
    read -p "Select test (1-6): " choice
    
    case $choice in
        1)
            source "$SCRIPT_DIR/test-task-9.1.sh"
            ;;
        2)
            source "$SCRIPT_DIR/test-task-9.2.sh"
            ;;
        3)
            source "$SCRIPT_DIR/test-task-9.3.sh"
            ;;
        4)
            source "$SCRIPT_DIR/test-task-9.4.sh"
            ;;
        5)
            source "$SCRIPT_DIR/test-task-9.1.sh"
            source "$SCRIPT_DIR/test-task-9.2.sh"
            source "$SCRIPT_DIR/test-task-9.3.sh"
            source "$SCRIPT_DIR/test-task-9.4.sh"
            ;;
        6)
            echo ""
            echo "Stopping ION-DTN..."
            ionstop
            echo "Goodbye!"
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid choice${NC}"
            ;;
    esac
done
