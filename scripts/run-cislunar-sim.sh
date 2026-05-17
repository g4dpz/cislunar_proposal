#!/bin/bash
# Three-node cislunar DTN simulation using HDTN
# Ground Station (Earth) → Lunar Orbiter (relay) → Lunar Lander
#
# Topology:
#   bping/bpsendfile (ipn:1.1)
#     → [STCP :4556] → Ground Station HDTN (nodeId=10)
#     → [LTP/UDP :2113 → :1113, 1300ms OWLT, 500 bps] → Orbiter HDTN (nodeId=20)
#     → [LTP/UDP :2123 → :3113, 10ms OWLT, 9600 bps] → Lander HDTN (nodeId=30)
#     → [STCP :4558] → bprecvfile (ipn:30.1)
#
# Return path:
#   Lander → [LTP/UDP :2143 → :3133] → Orbiter → [LTP/UDP :2133 → :1133] → Ground
#
# Usage:
#   ./scripts/run-cislunar-sim.sh
#
# To stop: Ctrl+C or kill the script

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
HDTN_BIN="${HDTN_BIN:-hdtn-one-process}"
SIM_DIR="$PROJECT_DIR/configs/simulation"
RECEIVED_DIR="/tmp/hdtn-sim/lander/received"

echo "=== Cislunar DTN Simulation (HDTN) ==="
echo ""
echo "Topology:"
echo "  Ground Station (nodeId=10) ←1.3s OWLT→ Orbiter (nodeId=20) ←10ms OWLT→ Lander (nodeId=30)"
echo ""
echo "Links:"
echo "  Earth → Orbiter:  LTP/UDP, 500 bps, 1300ms OWLT, port 2113→1113"
echo "  Orbiter → Lander: LTP/UDP, 9600 bps, 10ms OWLT, port 2123→3113"
echo "  Orbiter → Earth:  LTP/UDP, 500 bps, 1300ms OWLT, port 2133→1133"
echo "  Lander → Orbiter: LTP/UDP, 9600 bps, 10ms OWLT, port 2143→3133"
echo ""

# Create storage directories
mkdir -p /tmp/hdtn-sim/ground-station
mkdir -p /tmp/hdtn-sim/orbiter
mkdir -p /tmp/hdtn-sim/lander
mkdir -p "$RECEIVED_DIR"

# Check HDTN binary
if ! command -v "$HDTN_BIN" &> /dev/null; then
    echo "Error: $HDTN_BIN not found in PATH"
    echo "Set HDTN_BIN environment variable or install HDTN"
    exit 1
fi

echo "Starting nodes (reverse order: lander → orbiter → ground)..."
echo ""

# Start Lunar Lander (Node 3, nodeId=30) first
echo "[Lander]         Starting nodeId=30 (LTP induct :3113, STCP induct :4558)..."
$HDTN_BIN \
    --hdtn-config-file="$SIM_DIR/cislunar-lander.json" \
    --contact-plan-file="$SIM_DIR/cislunar-contact-plan.json" &
PID_LANDER=$!
echo "[Lander]         PID: $PID_LANDER"
sleep 5

# Start Lunar Orbiter (Node 2, nodeId=20)
echo "[Orbiter]        Starting nodeId=20 (LTP inducts :1113, :3133)..."
$HDTN_BIN \
    --hdtn-config-file="$SIM_DIR/cislunar-orbiter.json" \
    --contact-plan-file="$SIM_DIR/cislunar-contact-plan.json" &
PID_ORBITER=$!
echo "[Orbiter]        PID: $PID_ORBITER"
sleep 5

# Start Ground Station (Node 1, nodeId=10)
echo "[Ground Station] Starting nodeId=10 (STCP induct :4556, LTP induct :1133)..."
$HDTN_BIN \
    --hdtn-config-file="$SIM_DIR/cislunar-ground-station.json" \
    --contact-plan-file="$SIM_DIR/cislunar-contact-plan.json" &
PID_GROUND=$!
echo "[Ground Station] PID: $PID_GROUND"
sleep 5

# Start bprecvfile at the lander
echo "[bprecvfile]     Starting receiver at ipn:30.1..."
bprecvfile \
    --my-uri-eid=ipn:30.1 \
    --inducts-config-file="$SIM_DIR/bpsink-inducts.json" \
    --save-directory="$RECEIVED_DIR" \
    --max-rx-bundle-size-bytes=10000000 &
PID_RECV=$!
echo "[bprecvfile]     PID: $PID_RECV"
sleep 2

echo ""
echo "=== Simulation Ready ==="
echo ""
echo "To send a file from Ground Station to Lander:"
echo "  bpsendfile --my-uri-eid=ipn:10.1 --dest-uri-eid=ipn:30.1 \\"
echo "    --outducts-config-file=$SIM_DIR/bping-outducts.json \\"
echo "    --file-or-folder-path=<file>"
echo ""
echo "To ping the Lander from Ground Station:"
echo "  bping --my-uri-eid=ipn:10.1 --dest-uri-eid=ipn:30.2047 \\"
echo "    --outducts-config-file=$SIM_DIR/bping-outducts.json \\"
echo "    --bundle-lifetime=300"
echo ""
echo "Expected round-trip time: ~2.62s (1.3s + 0.01s each way + processing)"
echo ""
echo "Received files will appear in: $RECEIVED_DIR"
echo ""
echo "Press Ctrl+C to stop all nodes"
echo ""

# Trap Ctrl+C to clean up
cleanup() {
    echo ""
    echo "Stopping simulation..."
    kill $PID_RECV $PID_GROUND $PID_ORBITER $PID_LANDER 2>/dev/null || true
    wait $PID_RECV $PID_GROUND $PID_ORBITER $PID_LANDER 2>/dev/null || true
    echo "All nodes stopped."
}
trap cleanup INT TERM

# Wait for any node to exit
wait
