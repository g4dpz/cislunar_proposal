#!/bin/bash
# Task 9.1: Test bpsendfile / bprecvfile
# This script tests basic file transfer using ION-DTN store-and-forward

echo ""
echo "========================================"
echo "Task 9.1: Test bpsendfile / bprecvfile"
echo "========================================"
echo ""

# Create test directory
TEST_DIR="$PROJECT_DIR/test-data"
mkdir -p "$TEST_DIR"

if [[ "$NODE_TYPE" == "node-a" ]]; then
    # Node A: Sender
    echo "=== Node A: Preparing to send file ==="
    echo ""
    
    # Create a test file
    TEST_FILE="$TEST_DIR/testfile-9.1.txt"
    echo "Creating test file: $TEST_FILE"
    cat > "$TEST_FILE" << 'EOF'
Task 9.1 Test File
==================
This is a test file for ION-DTN store-and-forward validation.
Testing bpsendfile from Node A (ipn:1.1) to Node B (ipn:2.1).

Test data:
- Timestamp: $(date)
- Node: Node A (Engine 1)
- Destination: Node B (Engine 2)
- Protocol: BPv7 over LTP over AX.25 over KISS
- Hardware: Mobilinkd TNC4 + Yaesu FT-817 @ 9600 baud

Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
EOF
    
    # Calculate checksum
    if command -v md5sum &> /dev/null; then
        CHECKSUM=$(md5sum "$TEST_FILE" | awk '{print $1}')
    elif command -v md5 &> /dev/null; then
        CHECKSUM=$(md5 -q "$TEST_FILE")
    else
        CHECKSUM="N/A"
    fi
    
    echo "Test file created:"
    echo "  Path: $TEST_FILE"
    echo "  Size: $(wc -c < "$TEST_FILE") bytes"
    echo "  MD5: $CHECKSUM"
    echo ""
    
    echo "IMPORTANT: Ensure Node B is running bprecvfile before proceeding!"
    echo "On Node B, run: bprecvfile $REMOTE_EID 1"
    echo ""
    read -p "Press Enter when Node B is ready to receive, or Ctrl+C to abort..."
    echo ""
    
    # Send the file
    echo "Sending file to $REMOTE_EID..."
    echo "Command: bpsendfile $LOCAL_EID $REMOTE_EID $TEST_FILE"
    echo ""
    
    if bpsendfile "$LOCAL_EID" "$REMOTE_EID" "$TEST_FILE"; then
        echo ""
        echo -e "${GREEN}✓ File sent successfully${NC}"
        echo ""
        echo "File details for Node B to verify:"
        echo "  Expected MD5: $CHECKSUM"
        echo "  Expected size: $(wc -c < "$TEST_FILE") bytes"
        echo ""
        echo -e "${GREEN}✓ Task 9.1 PASSED (Node A): File sent via bpsendfile${NC}"
    else
        echo ""
        echo -e "${RED}✗ Task 9.1 FAILED (Node A): bpsendfile failed${NC}"
    fi
    
else
    # Node B: Receiver
    echo "=== Node B: Preparing to receive file ==="
    echo ""
    
    RECV_FILE="$TEST_DIR/received-9.1.txt"
    
    echo "Starting bprecvfile to receive from $REMOTE_EID..."
    echo "Command: bprecvfile $LOCAL_EID 1"
    echo ""
    echo "Waiting for file transfer (this may take 30-60 seconds)..."
    echo "Press Ctrl+C to abort if no file arrives after 2 minutes"
    echo ""
    
    # Run bprecvfile with timeout
    if timeout 120 bprecvfile "$LOCAL_EID" 1 > "$RECV_FILE" 2>&1; then
        echo ""
        echo -e "${GREEN}✓ File received${NC}"
        echo ""
        
        # Verify file
        if [[ -f "$RECV_FILE" && -s "$RECV_FILE" ]]; then
            # Calculate checksum
            if command -v md5sum &> /dev/null; then
                RECV_CHECKSUM=$(md5sum "$RECV_FILE" | awk '{print $1}')
            elif command -v md5 &> /dev/null; then
                RECV_CHECKSUM=$(md5 -q "$RECV_FILE")
            else
                RECV_CHECKSUM="N/A"
            fi
            
            echo "Received file details:"
            echo "  Path: $RECV_FILE"
            echo "  Size: $(wc -c < "$RECV_FILE") bytes"
            echo "  MD5: $RECV_CHECKSUM"
            echo ""
            
            echo "File content preview:"
            head -10 "$RECV_FILE"
            echo ""
            
            echo "Ask Node A for the expected MD5 checksum to verify integrity."
            echo ""
            echo -e "${GREEN}✓ Task 9.1 PASSED (Node B): File received via bprecvfile${NC}"
        else
            echo -e "${RED}✗ Received file is empty or missing${NC}"
            echo -e "${RED}✗ Task 9.1 FAILED (Node B): File not received properly${NC}"
        fi
    else
        echo ""
        echo -e "${RED}✗ bprecvfile timed out or failed${NC}"
        echo -e "${RED}✗ Task 9.1 FAILED (Node B): No file received${NC}"
        echo ""
        echo "Troubleshooting:"
        echo "1. Verify Node A sent the file"
        echo "2. Check ion.log for errors"
        echo "3. Verify contact window is active"
        echo "4. Check TNC4 and radio connections"
    fi
fi

echo ""
echo "=== Task 9.1 Complete ==="
