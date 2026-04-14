#!/bin/bash
# Task 9.2: Test store-and-forward with delayed contact
# This script tests bundle storage and delayed delivery

echo ""
echo "========================================"
echo "Task 9.2: Store-and-Forward with Delayed Contact"
echo "========================================"
echo ""

# Create test directory
TEST_DIR="$PROJECT_DIR/test-data"
mkdir -p "$TEST_DIR"

if [[ "$NODE_TYPE" == "node-a" ]]; then
    # Node A: Sender (will send while Node B is offline)
    echo "=== Node A: Testing delayed delivery ==="
    echo ""
    
    echo "This test validates that bundles are stored when the destination"
    echo "is unreachable and delivered when contact is re-established."
    echo ""
    
    echo "Test procedure:"
    echo "1. Node B will stop ION-DTN (simulating offline/out of contact)"
    echo "2. Node A will send a bundle"
    echo "3. Verify bundle is stored in Node A's bundle store"
    echo "4. Node B will restart ION-DTN"
    echo "5. Verify bundle is delivered to Node B"
    echo ""
    
    read -p "Press Enter when Node B has confirmed it is STOPPED..."
    echo ""
    
    # Create test file
    TEST_FILE="$TEST_DIR/testfile-9.2.txt"
    TIMESTAMP=$(date +%s)
    cat > "$TEST_FILE" << EOF
Task 9.2 Delayed Delivery Test
================================
Sent at: $(date)
Timestamp: $TIMESTAMP
From: Node A ($LOCAL_EID)
To: Node B ($REMOTE_EID)

This bundle was sent while Node B was offline.
It should be stored at Node A and delivered when Node B comes back online.
EOF
    
    echo "Test file created: $TEST_FILE"
    echo "Size: $(wc -c < "$TEST_FILE") bytes"
    echo ""
    
    # Send the file (should be queued since Node B is offline)
    echo "Sending bundle while Node B is offline..."
    echo "Command: bpsendfile $LOCAL_EID $REMOTE_EID $TEST_FILE"
    echo ""
    
    if bpsendfile "$LOCAL_EID" "$REMOTE_EID" "$TEST_FILE"; then
        echo ""
        echo -e "${GREEN}✓ Bundle queued for delivery${NC}"
        echo ""
        
        # Check bundle store
        echo "Checking bundle store status..."
        sleep 2
        
        # Use bpstats to check pending bundles
        echo "Bundle statistics:"
        bpstats || echo "(bpstats not available)"
        echo ""
        
        echo "The bundle should now be stored in Node A's bundle store,"
        echo "waiting for contact with Node B."
        echo ""
        echo "NEXT STEP: Tell Node B to restart ION-DTN"
        echo ""
        read -p "Press Enter when Node B has restarted and is ready to receive..."
        echo ""
        
        echo "Waiting for bundle delivery (30 seconds)..."
        sleep 30
        echo ""
        
        echo "Check with Node B to confirm bundle was received."
        echo ""
        echo -e "${GREEN}✓ Task 9.2 PASSED (Node A): Bundle stored and queued for delayed delivery${NC}"
    else
        echo ""
        echo -e "${RED}✗ Task 9.2 FAILED (Node A): bpsendfile failed${NC}"
    fi
    
else
    # Node B: Receiver (will be stopped, then restarted)
    echo "=== Node B: Testing delayed reception ==="
    echo ""
    
    echo "This test validates bundle reception after being offline."
    echo ""
    
    echo "Test procedure:"
    echo "1. Stop ION-DTN on this node (Node B)"
    echo "2. Node A will send a bundle"
    echo "3. Restart ION-DTN on this node"
    echo "4. Verify bundle is received"
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
    read -p "Press Enter after Node A has sent the bundle..."
    echo ""
    
    echo "Restarting ION-DTN on Node B..."
    bash "$STARTUP_SCRIPT"
    sleep 5
    echo -e "${GREEN}✓ ION-DTN restarted on Node B${NC}"
    echo ""
    
    # Start bprecvfile to receive the delayed bundle
    RECV_FILE="$TEST_DIR/received-9.2.txt"
    
    echo "Starting bprecvfile to receive delayed bundle..."
    echo "Command: bprecvfile $LOCAL_EID 1"
    echo ""
    echo "Waiting for bundle (timeout: 60 seconds)..."
    echo ""
    
    if timeout 60 bprecvfile "$LOCAL_EID" 1 > "$RECV_FILE" 2>&1; then
        echo ""
        echo -e "${GREEN}✓ Delayed bundle received${NC}"
        echo ""
        
        if [[ -f "$RECV_FILE" && -s "$RECV_FILE" ]]; then
            echo "Received file content:"
            cat "$RECV_FILE"
            echo ""
            
            # Extract timestamp to verify it was sent earlier
            if grep -q "Timestamp:" "$RECV_FILE"; then
                SENT_TIME=$(grep "Timestamp:" "$RECV_FILE" | awk '{print $2}')
                RECV_TIME=$(date +%s)
                DELAY=$((RECV_TIME - SENT_TIME))
                echo "Delivery delay: $DELAY seconds"
                echo ""
            fi
            
            echo -e "${GREEN}✓ Task 9.2 PASSED (Node B): Delayed bundle received after restart${NC}"
        else
            echo -e "${RED}✗ Received file is empty${NC}"
            echo -e "${RED}✗ Task 9.2 FAILED (Node B)${NC}"
        fi
    else
        echo ""
        echo -e "${RED}✗ No bundle received within timeout${NC}"
        echo -e "${RED}✗ Task 9.2 FAILED (Node B)${NC}"
    fi
fi

echo ""
echo "=== Task 9.2 Complete ==="
