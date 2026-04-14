#!/bin/bash
# Test ION-DTN bping over KISS CLA (Task 7)
# This script helps validate ION-DTN ping functionality between two nodes.
#
# PREREQUISITES:
# - Two nodes (Node A and Node B) with ION-DTN installed
# - TNC4 hardware connected via USB on both nodes
# - FT-817 radios configured for 9600 baud operation
# - Both nodes must run this test from separate terminals/machines
#
# USAGE:
#   On Node A: ./scripts/test-ion-bping.sh node-a
#   On Node B: ./scripts/test-ion-bping.sh node-b

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
NC='\033[0m' # No Color

# Parse arguments
NODE_TYPE="${1:-}"

if [[ "$NODE_TYPE" != "node-a" && "$NODE_TYPE" != "node-b" ]]; then
    echo "Usage: $0 <node-a|node-b>"
    echo ""
    echo "This script performs Task 7 testing: ION-DTN bping over KISS CLA"
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
echo "ION-DTN bping Test - Task 7"
echo "========================================"
echo "Node: $NODE_NAME"
echo "Local EID: $LOCAL_EID"
echo "Remote EID: $REMOTE_EID"
echo ""

# Task 7.1: Start ION-DTN and verify initialization
echo "=== Task 7.1: Starting ION-DTN on $NODE_NAME ==="
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

# Check ion.log for errors
echo ""
echo "=== Checking ion.log for initialization errors ==="
if [[ -f "ion.log" ]]; then
    ERROR_COUNT=$(grep -i "error" ion.log | wc -l | tr -d ' ')
    if [[ "$ERROR_COUNT" -eq 0 ]]; then
        echo -e "${GREEN}✓${NC} No errors found in ion.log"
    else
        echo -e "${YELLOW}⚠${NC} Found $ERROR_COUNT error(s) in ion.log:"
        grep -i "error" ion.log | tail -5
    fi
else
    echo -e "${YELLOW}⚠${NC} ion.log not found"
fi

echo ""
if [[ "$PROCESSES_OK" == true ]]; then
    echo -e "${GREEN}✓ Task 7.1 PASSED: ION-DTN initialized successfully${NC}"
else
    echo -e "${RED}✗ Task 7.1 FAILED: Some ION-DTN processes are not running${NC}"
    exit 1
fi

# Task 7.2 / 7.3: Run bping test
echo ""
echo "========================================"
if [[ "$NODE_TYPE" == "node-a" ]]; then
    echo "=== Task 7.2: Run bping from Node A to Node B ==="
else
    echo "=== Task 7.3: Run bping from Node B to Node A ==="
fi
echo "========================================"
echo ""
echo "Local endpoint: $LOCAL_EID"
echo "Remote endpoint: $REMOTE_EID"
echo ""
echo "IMPORTANT: Ensure the remote node ($REMOTE_EID) is running before proceeding!"
echo ""
read -p "Press Enter when the remote node is ready, or Ctrl+C to abort..."
echo ""

# Run bping
echo "Executing: bping $LOCAL_EID $REMOTE_EID -c 5"
echo ""

# Run bping and capture output
BPING_OUTPUT=$(mktemp)
if bping "$LOCAL_EID" "$REMOTE_EID" -c 5 2>&1 | tee "$BPING_OUTPUT"; then
    BPING_SUCCESS=true
else
    BPING_SUCCESS=false
fi

echo ""
echo "=== Analyzing bping results ==="

# Check for successful ping responses
RESPONSE_COUNT=$(grep -c "bytes from" "$BPING_OUTPUT" || true)
if [[ "$RESPONSE_COUNT" -gt 0 ]]; then
    echo -e "${GREEN}✓${NC} Received $RESPONSE_COUNT ping response(s) from $REMOTE_EID"
    
    # Extract and display round-trip times
    echo ""
    echo "Round-trip times:"
    grep "bytes from" "$BPING_OUTPUT" | grep -oE "time=[0-9.]+ ms" || true
    
    if [[ "$NODE_TYPE" == "node-a" ]]; then
        echo -e "\n${GREEN}✓ Task 7.2 PASSED: bping from Node A to Node B successful${NC}"
    else
        echo -e "\n${GREEN}✓ Task 7.3 PASSED: bping from Node B to Node A successful${NC}"
    fi
else
    echo -e "${RED}✗${NC} No ping responses received from $REMOTE_EID"
    echo ""
    echo "Troubleshooting tips:"
    echo "1. Verify the remote node is running (check with: ps aux | grep ion)"
    echo "2. Check TNC4 USB connections on both nodes"
    echo "3. Verify FT-817 radios are powered on and configured for 9600 baud"
    echo "4. Check ion.log on both nodes for errors"
    echo "5. Verify contact plan allows communication (check node.ionrc)"
    
    if [[ "$NODE_TYPE" == "node-a" ]]; then
        echo -e "\n${RED}✗ Task 7.2 FAILED: bping from Node A to Node B failed${NC}"
    else
        echo -e "\n${RED}✗ Task 7.3 FAILED: bping from Node B to Node A failed${NC}"
    fi
fi

# Cleanup
rm -f "$BPING_OUTPUT"

echo ""
echo "========================================"
echo "Test Complete"
echo "========================================"
echo ""
echo "To stop this node, run: ./scripts/stop-node.sh"
echo ""
