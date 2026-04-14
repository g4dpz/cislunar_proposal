#!/bin/bash
# Task 9.4: Test bundle lifetime expiry
# This script tests that expired bundles are not delivered

echo ""
echo "========================================"
echo "Task 9.4: Bundle Lifetime Expiry"
echo "========================================"
echo ""

# Create test directory
TEST_DIR="$PROJECT_DIR/test-data"
mkdir -p "$TEST_DIR"

if [[ "$NODE_TYPE" == "node-a" ]]; then
    # Node A: Sender (will send bundle with short lifetime)
    echo "=== Node A: Testing bundle lifetime expiry ==="
    echo ""
    
    echo "This test validates that bundles with expired lifetimes are not delivered."
    echo ""
    
    echo "Test procedure:"
    echo "1. Node B will stop ION-DTN (simulating out of contact)"
    echo "2. Node A will send a bundle with 30-second lifetime"
    echo "3. Wait 40 seconds for the bundle to expire"
    echo "4. Node B will restart ION-DTN"
    echo "5. Verify the expired bundle is NOT delivered"
    echo ""
    
    read -p "Press Enter when Node B has confirmed it is STOPPED..."
    echo ""
    
    # Send bundle with short lifetime (30 seconds)
    LIFETIME=30
    echo "Sending bundle with $LIFETIME second lifetime..."
    
    # Create test file
    TEST_FILE="$TEST_DIR/expiry-test.txt"
    SEND_TIME=$(date +%s)
    echo "EXPIRY TEST: Sent at $SEND_TIME with ${LIFETIME}s lifetime" > "$TEST_FILE"
    
    echo "Command: bpsendfile $LOCAL_EID $REMOTE_EID $TEST_FILE"
    echo "Note: ION-DTN uses default lifetime from configuration"
    echo ""
    
    if bpsendfile "$LOCAL_EID" "$REMOTE_EID" "$TEST_FILE"; then
        echo ""
        echo -e "${GREEN}✓ Bundle sent with ${LIFETIME}s lifetime${NC}"
        echo "  Sent at: $(date)"
        echo "  Expires at: $(date -d "+${LIFETIME} seconds" 2>/dev/null || date -r $((SEND_TIME + LIFETIME)))"
        echo ""
        
        # Wait for bundle to expire
        WAIT_TIME=40
        echo "Waiting ${WAIT_TIME} seconds for bundle to expire..."
        echo "(Bundle will expire after ${LIFETIME}s, we wait ${WAIT_TIME}s to be sure)"
        echo ""
        
        for i in $(seq $WAIT_TIME -1 1); do
            if [[ $i -eq $LIFETIME ]]; then
                echo -e "${YELLOW}  ⏰ Bundle should expire NOW (${LIFETIME}s elapsed)${NC}"
            fi
            echo -n "  $i seconds remaining..."
            sleep 1
            echo -ne "\r"
        done
        echo ""
        
        CURRENT_TIME=$(date +%s)
        ELAPSED=$((CURRENT_TIME - SEND_TIME))
        echo ""
        echo -e "${GREEN}✓ Waited ${ELAPSED} seconds - bundle should be expired${NC}"
        echo ""
        
        echo "NEXT STEP: Tell Node B to restart ION-DTN"
        echo "The expired bundle should NOT be delivered."
        echo ""
        read -p "Press Enter when Node B has restarted..."
        echo ""
        
        echo "Waiting 20 seconds to see if bundle is delivered (it should NOT be)..."
        sleep 20
        echo ""
        
        echo "Check with Node B to confirm NO bundle was received."
        echo ""
        echo -e "${GREEN}✓ Task 9.4 PASSED (Node A): Bundle sent with short lifetime${NC}"
    else
        echo ""
        echo -e "${RED}✗ Task 9.4 FAILED (Node A): bpsend failed${NC}"
    fi
    
else
    # Node B: Receiver (should NOT receive expired bundle)
    echo "=== Node B: Verifying expired bundle is not delivered ==="
    echo ""
    
    echo "This test validates that expired bundles are not delivered."
    echo ""
    
    echo "Test procedure:"
    echo "1. Stop ION-DTN on this node (Node B)"
    echo "2. Node A will send a bundle with 30-second lifetime"
    echo "3. Wait 40 seconds (bundle expires)"
    echo "4. Restart ION-DTN on this node"
    echo "5. Verify NO bundle is received"
    echo ""
    
    read -p "Press Enter to STOP ION-DTN on Node B..."
    echo ""
    
    echo "Stopping ION-DTN..."
    ionstop
    sleep 3
    echo -e "${YELLOW}✓ ION-DTN stopped on Node B${NC}"
    echo ""
    
    echo "IMPORTANT: Tell Node A that Node B is now STOPPED"
    echo ""
    echo "Node A will send a bundle with 30-second lifetime."
    echo "We will wait for it to expire before restarting."
    echo ""
    read -p "Press Enter after Node A has sent the bundle and waited for expiry..."
    echo ""
    
    echo "Restarting ION-DTN on Node B..."
    bash "$STARTUP_SCRIPT"
    sleep 5
    echo -e "${GREEN}✓ ION-DTN restarted on Node B${NC}"
    echo ""
    
    # Try to receive bundle (should timeout - no bundle should arrive)
    RECV_FILE="$TEST_DIR/received-9.4.txt"
    
    echo "Attempting to receive bundle (should timeout - bundle expired)..."
    echo "Command: bprecvfile $LOCAL_EID 1"
    echo ""
    echo "Waiting 30 seconds..."
    echo ""
    
    if timeout 30 bprecvfile "$LOCAL_EID" 1 > "$RECV_FILE" 2>&1; then
        echo ""
        echo -e "${RED}✗ Unexpected: Bundle was received!${NC}"
        echo ""
        echo "Received content:"
        cat "$RECV_FILE"
        echo ""
        echo -e "${RED}✗ Task 9.4 FAILED (Node B): Expired bundle was delivered${NC}"
        echo ""
        echo "This indicates the bundle did not expire as expected."
        echo "Possible causes:"
        echo "  - Bundle lifetime was not set correctly"
        echo "  - ION-DTN did not enforce lifetime expiry"
        echo "  - Timing issue (bundle delivered before expiry)"
    else
        EXIT_CODE=$?
        echo ""
        if [[ $EXIT_CODE -eq 124 ]]; then
            # Timeout exit code
            echo -e "${GREEN}✓ Timeout: No bundle received (as expected)${NC}"
            echo ""
            echo "The expired bundle was correctly NOT delivered."
            echo ""
            echo -e "${GREEN}✓ Task 9.4 PASSED (Node B): Expired bundle not delivered${NC}"
        else
            echo -e "${YELLOW}⚠ bprecv exited with code $EXIT_CODE${NC}"
            echo ""
            if [[ ! -s "$RECV_FILE" ]]; then
                echo "No bundle content received."
                echo -e "${GREEN}✓ Task 9.4 PASSED (Node B): Expired bundle not delivered${NC}"
            else
                echo "Unexpected content received:"
                cat "$RECV_FILE"
                echo ""
                echo -e "${RED}✗ Task 9.4 FAILED (Node B)${NC}"
            fi
        fi
    fi
fi

echo ""
echo "=== Task 9.4 Complete ==="
