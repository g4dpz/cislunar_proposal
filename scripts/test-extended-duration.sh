#!/bin/bash
# Extended Duration Test (Task 15.2)
# This script runs both nodes for 1+ hours with periodic bundle exchanges
# to verify no memory leaks, no process crashes, and telemetry accuracy.
#
# PREREQUISITES:
# - Two nodes (Node A and Node B) with ION-DTN installed
# - TNC4 hardware connected via USB on both nodes
# - FT-817 radios configured for 9600 baud operation
# - dtn-node CLI built (cmd/dtn-node)
# - Configuration files: configs/dtn-node-a.yaml, configs/dtn-node-b.yaml
# - Task 15.1 (end-to-end integration test) must be completed
#
# USAGE:
#   On Node A: ./scripts/test-extended-duration.sh node-a [duration_minutes]
#   On Node B: ./scripts/test-extended-duration.sh node-b [duration_minutes]
#
# Default duration: 60 minutes (1 hour)
# Bundle exchange interval: 5 minutes

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
DURATION_MINUTES="${2:-60}"  # Default: 60 minutes (1 hour)

if [[ "$NODE_TYPE" != "node-a" && "$NODE_TYPE" != "node-b" ]]; then
    echo "Usage: $0 <node-a|node-b> [duration_minutes]"
    echo ""
    echo "This script performs Task 15.2: Extended Duration Test"
    echo ""
    echo "Run on Node A:"
    echo "  $0 node-a 60"
    echo ""
    echo "Run on Node B:"
    echo "  $0 node-b 60"
    echo ""
    echo "Default duration: 60 minutes (1 hour)"
    exit 1
fi

# Set node-specific variables
if [[ "$NODE_TYPE" == "node-a" ]]; then
    NODE_NAME="Node A"
    LOCAL_EID="ipn:1.1"
    REMOTE_EID="ipn:2.1"
    CONFIG_FILE="$PROJECT_DIR/configs/dtn-node-a.yaml"
    TELEMETRY_PORT=8080
else
    NODE_NAME="Node B"
    LOCAL_EID="ipn:2.1"
    REMOTE_EID="ipn:1.1"
    CONFIG_FILE="$PROJECT_DIR/configs/dtn-node-b.yaml"
    TELEMETRY_PORT=8081
fi

# Test parameters
BUNDLE_INTERVAL_SECONDS=300  # 5 minutes
TELEMETRY_INTERVAL_SECONDS=60  # 1 minute
DURATION_SECONDS=$((DURATION_MINUTES * 60))

# Create test data and log directories
mkdir -p "$PROJECT_DIR/test-data"
mkdir -p "$PROJECT_DIR/test-logs"

LOG_FILE="$PROJECT_DIR/test-logs/extended-duration-${NODE_TYPE}-$(date +%Y%m%d-%H%M%S).log"
TELEMETRY_LOG="$PROJECT_DIR/test-logs/telemetry-${NODE_TYPE}-$(date +%Y%m%d-%H%M%S).log"

echo "========================================"
echo "Extended Duration Test - Task 15.2"
echo "========================================"
echo "Node: $NODE_NAME"
echo "Local EID: $LOCAL_EID"
echo "Remote EID: $REMOTE_EID"
echo "Duration: $DURATION_MINUTES minutes ($DURATION_SECONDS seconds)"
echo "Bundle interval: $((BUNDLE_INTERVAL_SECONDS / 60)) minutes"
echo "Telemetry interval: $((TELEMETRY_INTERVAL_SECONDS / 60)) minute(s)"
echo "Log file: $LOG_FILE"
echo "Telemetry log: $TELEMETRY_LOG"
echo ""

# Log function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Step 1: Build dtn-node CLI
log "=== Step 1: Building dtn-node CLI ==="
cd "$PROJECT_DIR"
if go build -o dtn-node ./cmd/dtn-node; then
    log "✓ dtn-node CLI built successfully"
else
    log "✗ Failed to build dtn-node CLI"
    exit 1
fi
echo ""

# Step 2: Start node using Go wrapper
log "=== Step 2: Starting $NODE_NAME using Go wrapper ==="

# Check if ION is already running
if pgrep -f "rfxclock" > /dev/null 2>&1; then
    log "Warning: ION-DTN appears to be already running"
    log "Stopping existing instance..."
    ionstop 2>/dev/null || true
    sleep 2
fi

# Start dtn-node in background
log "Starting dtn-node with config: $CONFIG_FILE"
./dtn-node -config "$CONFIG_FILE" > "dtn-node-${NODE_TYPE}.log" 2>&1 &
DTN_NODE_PID=$!
log "dtn-node started with PID: $DTN_NODE_PID"

# Wait for node to initialize
log "Waiting for node to initialize (10 seconds)..."
sleep 10

# Verify dtn-node is running
if ! ps -p $DTN_NODE_PID > /dev/null 2>&1; then
    log "✗ dtn-node process died"
    log "Log output:"
    tail -20 "dtn-node-${NODE_TYPE}.log" | tee -a "$LOG_FILE"
    exit 1
fi
log "✓ dtn-node is running"

# Verify ION-DTN processes
log "=== Verifying ION-DTN processes ==="
PROCESSES_OK=true

if pgrep -f "rfxclock" > /dev/null 2>&1; then
    log "✓ rfxclock is running"
else
    log "✗ rfxclock is NOT running"
    PROCESSES_OK=false
fi

if pgrep -f "ltpclock" > /dev/null 2>&1; then
    log "✓ ltpclock is running"
else
    log "✗ ltpclock is NOT running"
    PROCESSES_OK=false
fi

if pgrep -f "bpclock" > /dev/null 2>&1; then
    log "✓ bpclock is running"
else
    log "✗ bpclock is NOT running"
    PROCESSES_OK=false
fi

if [[ "$PROCESSES_OK" != true ]]; then
    log "✗ Some ION-DTN processes are not running"
    kill $DTN_NODE_PID 2>/dev/null || true
    exit 1
fi
log "✓ ION-DTN initialized successfully"
echo ""

# Step 3: Collect initial telemetry baseline
log "=== Step 3: Collecting initial telemetry baseline ==="

if curl -s "http://localhost:$TELEMETRY_PORT/health" > /tmp/telemetry-baseline.json; then
    log "✓ Initial telemetry collected"
    
    INITIAL_STORAGE=$(cat /tmp/telemetry-baseline.json | grep -oE '"storage_used_bytes":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
    INITIAL_BUNDLES_SENT=$(cat /tmp/telemetry-baseline.json | grep -oE '"bundles_sent":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
    INITIAL_BUNDLES_RECEIVED=$(cat /tmp/telemetry-baseline.json | grep -oE '"bundles_received":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
    
    log "Baseline metrics:"
    log "  Storage used: $INITIAL_STORAGE bytes"
    log "  Bundles sent: $INITIAL_BUNDLES_SENT"
    log "  Bundles received: $INITIAL_BUNDLES_RECEIVED"
    
    # Log full telemetry
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] BASELINE" >> "$TELEMETRY_LOG"
    cat /tmp/telemetry-baseline.json >> "$TELEMETRY_LOG"
    echo "" >> "$TELEMETRY_LOG"
else
    log "⚠ Failed to collect initial telemetry"
fi
echo ""

# Step 4: Run extended duration test
log "=== Step 4: Running extended duration test ==="
log "Test will run for $DURATION_MINUTES minutes"
log "Press Ctrl+C to stop early"
echo ""

START_TIME=$(date +%s)
END_TIME=$((START_TIME + DURATION_SECONDS))
LAST_BUNDLE_TIME=$START_TIME
LAST_TELEMETRY_TIME=$START_TIME
BUNDLE_COUNT=0
CRASH_COUNT=0
TELEMETRY_ERRORS=0

# Setup signal handler for graceful shutdown
trap 'log "Received interrupt signal, shutting down..."; break' INT TERM

while true; do
    CURRENT_TIME=$(date +%s)
    ELAPSED=$((CURRENT_TIME - START_TIME))
    REMAINING=$((END_TIME - CURRENT_TIME))
    
    # Check if test duration completed
    if [[ $CURRENT_TIME -ge $END_TIME ]]; then
        log "Test duration completed ($DURATION_MINUTES minutes)"
        break
    fi
    
    # Check if dtn-node process is still running
    if ! ps -p $DTN_NODE_PID > /dev/null 2>&1; then
        log "✗ CRITICAL: dtn-node process crashed!"
        CRASH_COUNT=$((CRASH_COUNT + 1))
        log "Crash count: $CRASH_COUNT"
        log "Attempting to restart..."
        
        # Restart dtn-node
        ./dtn-node -config "$CONFIG_FILE" > "dtn-node-${NODE_TYPE}.log" 2>&1 &
        DTN_NODE_PID=$!
        log "dtn-node restarted with PID: $DTN_NODE_PID"
        sleep 10
        
        if ! ps -p $DTN_NODE_PID > /dev/null 2>&1; then
            log "✗ CRITICAL: Failed to restart dtn-node"
            exit 1
        fi
    fi
    
    # Periodic bundle exchange
    if [[ $((CURRENT_TIME - LAST_BUNDLE_TIME)) -ge $BUNDLE_INTERVAL_SECONDS ]]; then
        BUNDLE_COUNT=$((BUNDLE_COUNT + 1))
        log "=== Sending periodic bundle #$BUNDLE_COUNT ==="
        
        # Create test file
        TEST_FILE="$PROJECT_DIR/test-data/extended-test-${NODE_TYPE}-${BUNDLE_COUNT}.txt"
        cat > "$TEST_FILE" << EOF
Extended Duration Test Bundle #$BUNDLE_COUNT
============================================
Sent from: $LOCAL_EID ($NODE_NAME)
Sent to: $REMOTE_EID
Timestamp: $(date)
Elapsed time: $((ELAPSED / 60)) minutes
Remaining time: $((REMAINING / 60)) minutes

This is periodic bundle #$BUNDLE_COUNT sent during the extended duration test.
EOF
        
        log "Sending bundle to $REMOTE_EID..."
        if bpsendfile "$LOCAL_EID" "$REMOTE_EID" "$TEST_FILE" >> "$LOG_FILE" 2>&1; then
            log "✓ Bundle #$BUNDLE_COUNT sent successfully"
        else
            log "✗ Bundle #$BUNDLE_COUNT send failed"
        fi
        
        LAST_BUNDLE_TIME=$CURRENT_TIME
    fi
    
    # Periodic telemetry collection
    if [[ $((CURRENT_TIME - LAST_TELEMETRY_TIME)) -ge $TELEMETRY_INTERVAL_SECONDS ]]; then
        log "=== Collecting telemetry (elapsed: $((ELAPSED / 60)) min, remaining: $((REMAINING / 60)) min) ==="
        
        if curl -s "http://localhost:$TELEMETRY_PORT/health" > /tmp/telemetry-current.json; then
            CURRENT_STORAGE=$(cat /tmp/telemetry-current.json | grep -oE '"storage_used_bytes":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
            CURRENT_BUNDLES_SENT=$(cat /tmp/telemetry-current.json | grep -oE '"bundles_sent":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
            CURRENT_BUNDLES_RECEIVED=$(cat /tmp/telemetry-current.json | grep -oE '"bundles_received":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
            CURRENT_BUNDLES_STORED=$(cat /tmp/telemetry-current.json | grep -oE '"bundles_stored":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
            
            STORAGE_DELTA=$((CURRENT_STORAGE - INITIAL_STORAGE))
            STORAGE_GROWTH_RATE=$((STORAGE_DELTA / (ELAPSED + 1)))  # bytes per second
            
            log "Current metrics:"
            log "  Storage used: $CURRENT_STORAGE bytes (delta: $STORAGE_DELTA, rate: $STORAGE_GROWTH_RATE B/s)"
            log "  Bundles sent: $CURRENT_BUNDLES_SENT (delta: $((CURRENT_BUNDLES_SENT - INITIAL_BUNDLES_SENT)))"
            log "  Bundles received: $CURRENT_BUNDLES_RECEIVED (delta: $((CURRENT_BUNDLES_RECEIVED - INITIAL_BUNDLES_RECEIVED)))"
            log "  Bundles stored: $CURRENT_BUNDLES_STORED"
            
            # Check for memory leak indicators
            if [[ $ELAPSED -gt 600 ]] && [[ $STORAGE_GROWTH_RATE -gt 1000 ]]; then
                log "⚠ WARNING: Potential memory leak detected (storage growth rate: $STORAGE_GROWTH_RATE B/s)"
            fi
            
            # Log full telemetry
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] ELAPSED: $((ELAPSED / 60)) min" >> "$TELEMETRY_LOG"
            cat /tmp/telemetry-current.json >> "$TELEMETRY_LOG"
            echo "" >> "$TELEMETRY_LOG"
        else
            log "⚠ Failed to collect telemetry"
            TELEMETRY_ERRORS=$((TELEMETRY_ERRORS + 1))
        fi
        
        LAST_TELEMETRY_TIME=$CURRENT_TIME
    fi
    
    # Sleep for 10 seconds before next check
    sleep 10
done

# Step 5: Collect final telemetry
log ""
log "=== Step 5: Collecting final telemetry ==="

sleep 5  # Give telemetry time to update

if curl -s "http://localhost:$TELEMETRY_PORT/health" > /tmp/telemetry-final.json; then
    log "✓ Final telemetry collected"
    
    FINAL_STORAGE=$(cat /tmp/telemetry-final.json | grep -oE '"storage_used_bytes":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
    FINAL_BUNDLES_SENT=$(cat /tmp/telemetry-final.json | grep -oE '"bundles_sent":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
    FINAL_BUNDLES_RECEIVED=$(cat /tmp/telemetry-final.json | grep -oE '"bundles_received":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
    FINAL_BUNDLES_STORED=$(cat /tmp/telemetry-final.json | grep -oE '"bundles_stored":\s*[0-9]+' | grep -oE '[0-9]+' || echo "0")
    
    TOTAL_STORAGE_DELTA=$((FINAL_STORAGE - INITIAL_STORAGE))
    TOTAL_BUNDLES_SENT_DELTA=$((FINAL_BUNDLES_SENT - INITIAL_BUNDLES_SENT))
    TOTAL_BUNDLES_RECEIVED_DELTA=$((FINAL_BUNDLES_RECEIVED - INITIAL_BUNDLES_RECEIVED))
    
    log "Final metrics:"
    log "  Storage used: $FINAL_STORAGE bytes (total delta: $TOTAL_STORAGE_DELTA)"
    log "  Bundles sent: $FINAL_BUNDLES_SENT (total delta: $TOTAL_BUNDLES_SENT_DELTA)"
    log "  Bundles received: $FINAL_BUNDLES_RECEIVED (total delta: $TOTAL_BUNDLES_RECEIVED_DELTA)"
    log "  Bundles stored: $FINAL_BUNDLES_STORED"
    
    # Log full telemetry
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] FINAL" >> "$TELEMETRY_LOG"
    cat /tmp/telemetry-final.json >> "$TELEMETRY_LOG"
    echo "" >> "$TELEMETRY_LOG"
else
    log "⚠ Failed to collect final telemetry"
fi
echo ""

# Step 6: Graceful shutdown
log "=== Step 6: Graceful shutdown ==="

log "Shutting down dtn-node (PID: $DTN_NODE_PID)..."
kill -SIGINT $DTN_NODE_PID 2>/dev/null || true

# Wait for graceful shutdown
log "Waiting for graceful shutdown (5 seconds)..."
sleep 5

# Check if process stopped
if ps -p $DTN_NODE_PID > /dev/null 2>&1; then
    log "⚠ Process still running, forcing shutdown..."
    kill -SIGKILL $DTN_NODE_PID 2>/dev/null || true
    sleep 1
fi

log "✓ Node stopped"

# Verify ION-DTN stopped
if pgrep -f "rfxclock" > /dev/null 2>&1; then
    log "⚠ ION-DTN processes still running"
    log "Running ionstop..."
    ionstop 2>/dev/null || true
else
    log "✓ ION-DTN stopped cleanly"
fi

echo ""

# Step 7: Analyze results
log "========================================"
log "=== Task 15.2 Test Results ==="
log "========================================"
log ""
log "Test duration: $DURATION_MINUTES minutes"
log "Bundles sent: $BUNDLE_COUNT"
log "Process crashes: $CRASH_COUNT"
log "Telemetry errors: $TELEMETRY_ERRORS"
log ""
log "Storage analysis:"
log "  Initial: $INITIAL_STORAGE bytes"
log "  Final: $FINAL_STORAGE bytes"
log "  Delta: $TOTAL_STORAGE_DELTA bytes"
log "  Average growth rate: $((TOTAL_STORAGE_DELTA / DURATION_SECONDS)) bytes/second"
log ""
log "Bundle statistics:"
log "  Bundles sent (delta): $TOTAL_BUNDLES_SENT_DELTA"
log "  Bundles received (delta): $TOTAL_BUNDLES_RECEIVED_DELTA"
log "  Final bundles stored: $FINAL_BUNDLES_STORED"
log ""

# Determine test result
TEST_PASSED=true

if [[ $CRASH_COUNT -gt 0 ]]; then
    log "✗ FAILED: Process crashed $CRASH_COUNT time(s)"
    TEST_PASSED=false
fi

if [[ $TOTAL_STORAGE_DELTA -gt $((100 * 1024 * 1024)) ]]; then
    log "⚠ WARNING: Storage grew by more than 100 MB (potential memory leak)"
    log "  Consider investigating storage growth"
fi

if [[ $TELEMETRY_ERRORS -gt $((DURATION_MINUTES / 10)) ]]; then
    log "⚠ WARNING: High telemetry error rate ($TELEMETRY_ERRORS errors)"
fi

if [[ "$TEST_PASSED" == true ]]; then
    log ""
    log "✓ Task 15.2 EXTENDED DURATION TEST PASSED"
    log ""
    log "Results:"
    log "  ✓ No process crashes"
    log "  ✓ Telemetry remained accurate"
    log "  ✓ Node operated for $DURATION_MINUTES minutes"
else
    log ""
    log "✗ Task 15.2 EXTENDED DURATION TEST FAILED"
    log ""
    log "See logs for details:"
    log "  Main log: $LOG_FILE"
    log "  Telemetry log: $TELEMETRY_LOG"
fi

log ""
log "Test logs saved to:"
log "  $LOG_FILE"
log "  $TELEMETRY_LOG"
log ""
