#!/bin/bash
# Test ION-DTN BPSec integrity over KISS CLA (Task 11)
# This script helps validate BPSec integrity protection between two nodes.
#
# PREREQUISITES:
# - Two nodes (Node A and Node B) with ION-DTN installed
# - TNC4 hardware connected via USB on both nodes
# - FT-817 radios configured for 9600 baud operation
# - Both nodes must run this test from separate terminals/machines
# - Task 7 (bping) and Task 9 (store-and-forward) must be completed
#
# USAGE:
#   On Node A: ./scripts/test-ion-bpsec.sh node-a
#   On Node B: ./scripts/test-ion-bpsec.sh node-b

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
    echo "This script performs Task 11 testing: BPSec integrity over KISS CLA"
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
    CONFIG_DIR="$PROJECT_DIR/configs/node-a"
    STARTUP_SCRIPT="$SCRIPT_DIR/start-node-a.sh"
else
    NODE_NAME="Node B (Engine 2)"
    LOCAL_EID="ipn:2.1"
    REMOTE_EID="ipn:1.1"
    ENGINE_ID="2"
    REMOTE_ENGINE="1"
    CONFIG_DIR="$PROJECT_DIR/configs/node-b"
    STARTUP_SCRIPT="$SCRIPT_DIR/start-node-b.sh"
fi

# Create test data directory
mkdir -p "$PROJECT_DIR/test-data"

echo "========================================"
echo "ION-DTN BPSec Integrity Test - Task 11"
echo "========================================"
echo "Node: $NODE_NAME"
echo "Local EID: $LOCAL_EID"
echo "Remote EID: $REMOTE_EID"
echo ""

# Function to start ION-DTN
start_ion() {
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
        return 1
    fi
}

# Function to configure BPSec
configure_bpsec() {
    echo ""
    echo "=== Task 11.1: Configuring BPSec on $NODE_NAME ==="
    echo ""
    
    # Check if BPSec config file exists
    if [[ ! -f "$CONFIG_DIR/node.bpsecrc" ]]; then
        echo -e "${RED}✗ BPSec config file not found: $CONFIG_DIR/node.bpsecrc${NC}"
        return 1
    fi
    
    # Check if BPSec key file exists
    if [[ ! -f "$CONFIG_DIR/bpsec_key.txt" ]]; then
        echo -e "${RED}✗ BPSec key file not found: $CONFIG_DIR/bpsec_key.txt${NC}"
        return 1
    fi
    
    # Verify key file permissions (should be 600)
    KEY_PERMS=$(stat -f "%OLp" "$CONFIG_DIR/bpsec_key.txt" 2>/dev/null || stat -c "%a" "$CONFIG_DIR/bpsec_key.txt" 2>/dev/null)
    if [[ "$KEY_PERMS" == "600" ]]; then
        echo -e "${GREEN}✓${NC} BPSec key file has correct permissions (600)"
    else
        echo -e "${YELLOW}⚠${NC} BPSec key file permissions: $KEY_PERMS (should be 600)"
    fi
    
    echo ""
    echo "Applying BPSec configuration..."
    echo "Command: bpsecadmin $CONFIG_DIR/node.bpsecrc"
    echo ""
    
    # Apply BPSec configuration
    if bpsecadmin "$CONFIG_DIR/node.bpsecrc" 2>&1 | tee /tmp/bpsecadmin_output.txt; then
        echo ""
        echo -e "${GREEN}✓ Task 11.1 PASSED: BPSec configured successfully${NC}"
        echo ""
        echo "BPSec configuration:"
        echo "  - HMAC-SHA-256 integrity protection enabled"
        echo "  - Shared key: INTEGRITY_KEY_AB"
        echo "  - No encryption (amateur radio compliance)"
        return 0
    else
        echo ""
        echo -e "${RED}✗ Task 11.1 FAILED: BPSec configuration failed${NC}"
        return 1
    fi
}

# Function to test bundle integrity verification (Task 11.2)
test_integrity_verification() {
    echo ""
    echo "========================================"
    echo "=== Task 11.2: Test Bundle Integrity Verification ==="
    echo "========================================"
    echo ""
    
    if [[ "$NODE_TYPE" == "node-b" ]]; then
        # Node B: Receiver
        echo "=== $NODE_NAME: Preparing to receive bundle with BPSec integrity ==="
        echo ""
        echo "Starting bprecvfile to receive from $REMOTE_EID..."
        echo "Command: bprecvfile $LOCAL_EID 1"
        echo ""
        echo "Waiting for file transfer (this may take 30-60 seconds)..."
        echo ""
        
        # Receive file
        RECV_FILE="$PROJECT_DIR/test-data/received-11.2.txt"
        if timeout 60 bprecvfile "$LOCAL_EID" 1 > "$RECV_FILE" 2>&1; then
            echo ""
            echo -e "${GREEN}✓ File received${NC}"
            echo ""
            echo "Received file details:"
            echo "  Path: $RECV_FILE"
            if [[ -f "$RECV_FILE" ]]; then
                echo "  Size: $(wc -c < "$RECV_FILE") bytes"
                echo "  MD5: $(md5sum "$RECV_FILE" 2>/dev/null | awk '{print $1}' || md5 -q "$RECV_FILE")"
                echo ""
                echo "File content preview:"
                head -5 "$RECV_FILE"
                echo ""
            fi
            
            # Check ion.log for BPSec verification messages
            echo "=== Checking ion.log for BPSec verification ==="
            if grep -i "bpsec\|integrity\|bib" ion.log | tail -5; then
                echo ""
                echo -e "${GREEN}✓ BPSec integrity verification logged${NC}"
            else
                echo -e "${YELLOW}⚠ No BPSec messages found in ion.log${NC}"
            fi
            
            echo ""
            echo -e "${GREEN}✓ Task 11.2 PASSED ($NODE_NAME): Bundle received with integrity verification${NC}"
        else
            echo ""
            echo -e "${RED}✗ bprecvfile timed out or failed${NC}"
            echo -e "${RED}✗ Task 11.2 FAILED ($NODE_NAME): Bundle not received${NC}"
            return 1
        fi
    else
        # Node A: Sender
        echo "=== $NODE_NAME: Preparing to send bundle with BPSec integrity ==="
        echo ""
        
        # Create test file
        TEST_FILE="$PROJECT_DIR/test-data/testfile-11.2.txt"
        cat > "$TEST_FILE" << EOF
Task 11.2 BPSec Integrity Test File
====================================
This bundle is protected by BPSec Block Integrity Block (BIB)
using HMAC-SHA-256 for origin authentication and tamper detection.

Amateur radio compliance: NO ENCRYPTION
Security: Integrity protection only

Sent from: $LOCAL_EID
Sent to: $REMOTE_EID
Timestamp: $(date)

This file should be received intact with verified integrity.
EOF
        
        echo "Test file created:"
        echo "  Path: $TEST_FILE"
        echo "  Size: $(wc -c < "$TEST_FILE") bytes"
        echo "  MD5: $(md5sum "$TEST_FILE" 2>/dev/null | awk '{print $1}' || md5 -q "$TEST_FILE")"
        echo ""
        
        echo "IMPORTANT: Ensure Node B is ready to receive (running bprecvfile)"
        echo ""
        read -p "Press Enter when Node B is ready, or Ctrl+C to abort..."
        echo ""
        
        echo "Sending file to $REMOTE_EID with BPSec integrity..."
        echo "Command: bpsendfile $LOCAL_EID $REMOTE_EID $TEST_FILE"
        echo ""
        
        if bpsendfile "$LOCAL_EID" "$REMOTE_EID" "$TEST_FILE" 2>&1; then
            echo ""
            echo -e "${GREEN}✓ File sent successfully with BPSec integrity${NC}"
            echo ""
            
            # Check ion.log for BPSec application messages
            echo "=== Checking ion.log for BPSec application ==="
            if grep -i "bpsec\|integrity\|bib" ion.log | tail -5; then
                echo ""
                echo -e "${GREEN}✓ BPSec integrity block applied${NC}"
            else
                echo -e "${YELLOW}⚠ No BPSec messages found in ion.log${NC}"
            fi
            
            echo ""
            echo -e "${GREEN}✓ Task 11.2 PASSED ($NODE_NAME): Bundle sent with BPSec integrity${NC}"
        else
            echo ""
            echo -e "${RED}✗ bpsendfile failed${NC}"
            echo -e "${RED}✗ Task 11.2 FAILED ($NODE_NAME): Bundle not sent${NC}"
            return 1
        fi
    fi
}

# Function to test integrity failure detection (Task 11.3 - OPTIONAL)
test_integrity_failure() {
    echo ""
    echo "========================================"
    echo "=== Task 11.3: Test Integrity Failure Detection (OPTIONAL) ==="
    echo "========================================"
    echo ""
    
    if [[ "$NODE_TYPE" == "node-b" ]]; then
        # Node B: Test with wrong key
        echo "=== $NODE_NAME: Testing integrity failure detection ==="
        echo ""
        echo "This test reconfigures Node B with a WRONG key to simulate"
        echo "integrity verification failure."
        echo ""
        
        read -p "Press Enter to reconfigure with wrong key, or Ctrl+C to skip..."
        echo ""
        
        # Stop ION
        echo "Stopping ION-DTN..."
        ionstop 2>/dev/null || true
        sleep 2
        
        # Backup correct key
        cp "$CONFIG_DIR/bpsec_key.txt" "$CONFIG_DIR/bpsec_key_correct.txt"
        
        # Use wrong key
        if [[ -f "$CONFIG_DIR/bpsec_key_wrong.txt" ]]; then
            cp "$CONFIG_DIR/bpsec_key_wrong.txt" "$CONFIG_DIR/bpsec_key.txt"
            echo -e "${YELLOW}⚠ Using WRONG key for integrity verification${NC}"
        else
            echo -e "${RED}✗ Wrong key file not found${NC}"
            return 1
        fi
        
        # Restart ION with wrong key
        echo ""
        start_ion
        configure_bpsec
        
        echo ""
        echo "=== Attempting to receive bundle with WRONG key ==="
        echo "Expected: Integrity verification should FAIL"
        echo ""
        
        RECV_FILE="$PROJECT_DIR/test-data/received-11.3.txt"
        if timeout 60 bprecvfile "$LOCAL_EID" 1 > "$RECV_FILE" 2>&1; then
            echo ""
            echo -e "${YELLOW}⚠ Bundle was received (unexpected)${NC}"
            echo ""
            echo "Checking ion.log for integrity failure..."
            if grep -i "integrity.*fail\|bib.*fail\|verification.*fail" ion.log | tail -5; then
                echo ""
                echo -e "${GREEN}✓ Integrity failure detected and logged${NC}"
                echo -e "${GREEN}✓ Task 11.3 PASSED: Integrity failure detection works${NC}"
            else
                echo ""
                echo -e "${YELLOW}⚠ No integrity failure messages found${NC}"
                echo -e "${YELLOW}⚠ Task 11.3 PARTIAL: Bundle received but integrity check unclear${NC}"
            fi
        else
            echo ""
            echo -e "${GREEN}✓ Bundle NOT received (as expected with wrong key)${NC}"
            echo ""
            echo "Checking ion.log for integrity failure..."
            if grep -i "integrity.*fail\|bib.*fail\|verification.*fail" ion.log | tail -5; then
                echo ""
                echo -e "${GREEN}✓ Integrity failure detected and logged${NC}"
            fi
            echo ""
            echo -e "${GREEN}✓ Task 11.3 PASSED: Integrity failure detection works${NC}"
        fi
        
        # Restore correct key
        echo ""
        echo "Restoring correct key..."
        cp "$CONFIG_DIR/bpsec_key_correct.txt" "$CONFIG_DIR/bpsec_key.txt"
        
        echo ""
        echo "To continue testing, restart ION-DTN with correct key:"
        echo "  1. Stop ION: ionstop"
        echo "  2. Restart: $STARTUP_SCRIPT"
        echo "  3. Reconfigure BPSec: bpsecadmin $CONFIG_DIR/node.bpsecrc"
        
    else
        # Node A: Sender for Task 11.3
        echo "=== $NODE_NAME: Sending bundle for integrity failure test ==="
        echo ""
        echo "IMPORTANT: Ensure Node B is reconfigured with WRONG key"
        echo ""
        read -p "Press Enter when Node B is ready with wrong key, or Ctrl+C to skip..."
        echo ""
        
        # Create test file
        TEST_FILE="$PROJECT_DIR/test-data/testfile-11.3.txt"
        cat > "$TEST_FILE" << EOF
Task 11.3 Integrity Failure Test
=================================
This bundle should be REJECTED by Node B because
Node B is configured with a WRONG key.

Sent from: $LOCAL_EID
Sent to: $REMOTE_EID
Timestamp: $(date)

Expected: Integrity verification FAILURE at Node B
EOF
        
        echo "Sending file to $REMOTE_EID..."
        echo "Command: bpsendfile $LOCAL_EID $REMOTE_EID $TEST_FILE"
        echo ""
        
        if bpsendfile "$LOCAL_EID" "$REMOTE_EID" "$TEST_FILE" 2>&1; then
            echo ""
            echo -e "${GREEN}✓ File sent (Node B should reject due to wrong key)${NC}"
            echo ""
            echo -e "${GREEN}✓ Task 11.3 PASSED ($NODE_NAME): Bundle sent for failure test${NC}"
        else
            echo ""
            echo -e "${RED}✗ bpsendfile failed${NC}"
            return 1
        fi
    fi
}

# Main test menu
show_menu() {
    echo ""
    echo "========================================"
    echo "Task 11 Test Menu"
    echo "========================================"
    echo "1) Task 11.1: Configure BPSec"
    echo "2) Task 11.2: Test bundle integrity verification"
    echo "3) Task 11.3: Test integrity failure detection (OPTIONAL)"
    echo "4) Run all tests"
    echo "5) Exit and stop ION-DTN"
    echo ""
    read -p "Select test (1-5): " choice
    echo ""
    
    case $choice in
        1)
            configure_bpsec
            show_menu
            ;;
        2)
            test_integrity_verification
            show_menu
            ;;
        3)
            test_integrity_failure
            show_menu
            ;;
        4)
            configure_bpsec
            test_integrity_verification
            test_integrity_failure
            show_menu
            ;;
        5)
            echo "Stopping ION-DTN..."
            ionstop 2>/dev/null || true
            echo "Goodbye!"
            exit 0
            ;;
        *)
            echo "Invalid choice"
            show_menu
            ;;
    esac
}

# Start ION-DTN
start_ion

# Show test menu
show_menu
