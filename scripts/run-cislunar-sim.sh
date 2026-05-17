#!/bin/bash
# Three-node cislunar DTN simulation using HDTN
# Based on the working HDTN 3-hop LTP delayed test
#
# Topology:
#   bpsendfile (ipn:1.1)
#     → [STCP :4556] → Node1 HDTN (nodeId=10, Ground Station)
#     → [LTP/UDP :2113 → :1113, 1s OWLT, 250ms tx delay] → Node2 HDTN (nodeId=20, Orbiter)
#     → [LTP/UDP :2123 → :3113, 1s OWLT, 250ms tx delay] → Node3 HDTN (nodeId=30, Lander)
#     → [STCP :4558] → bpreceivefile (ipn:3.1)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SIM_DIR="$PROJECT_DIR/configs/simulation"
RECEIVED_DIR="$PROJECT_DIR/cislunar_received"

echo "=== Cislunar DTN Simulation (HDTN 3-hop LTP) ==="
echo ""
echo "Topology:"
echo "  bpsendfile (ipn:1.1) → Ground (10) → Orbiter (20) → Lander (30) → bpreceivefile (ipn:3.1)"
echo ""
echo "Link delays: 1s OWLT per hop + 250ms TX delay = ~2.5s end-to-end"
echo ""

mkdir -p "$RECEIVED_DIR"

# Check HDTN binary
if ! command -v hdtn-one-process &> /dev/null; then
    echo "Error: hdtn-one-process not found in PATH"
    exit 1
fi

echo "Starting spacecraft receiver (bpreceivefile, ipn:3.1) on STCP port 4558..."
bpreceivefile \
    --my-uri-eid=ipn:3.1 \
    --inducts-config-file="$SIM_DIR/bpsink-inducts.json" \
    --save-directory="$RECEIVED_DIR" \
    --max-rx-bundle-size-bytes=10000000 &
PID_RECV=$!
sleep 2

echo "Starting Lander HDTN (nodeId=30)..."
hdtn-one-process \
    --hdtn-config-file="$SIM_DIR/cislunar-lander.json" \
    --contact-plan-file="$SIM_DIR/cislunar-contact-plan.json" &
PID_LANDER=$!
sleep 8

echo "Starting Orbiter HDTN (nodeId=20)..."
hdtn-one-process \
    --hdtn-config-file="$SIM_DIR/cislunar-orbiter.json" \
    --contact-plan-file="$SIM_DIR/cislunar-contact-plan.json" &
PID_ORBITER=$!
sleep 8

echo "Starting Ground Station HDTN (nodeId=10)..."
hdtn-one-process \
    --hdtn-config-file="$SIM_DIR/cislunar-ground-station.json" \
    --contact-plan-file="$SIM_DIR/cislunar-contact-plan.json" &
PID_GROUND=$!
sleep 8

echo ""
echo "=== Simulation Ready ==="
echo ""
echo "To send files from Earth to Lander:"
echo "  mkdir -p /tmp/hdtn-send"
echo "  echo 'Hello Moon' > /tmp/hdtn-send/message.dat"
echo "  bpsendfile --my-uri-eid=ipn:1.1 --dest-uri-eid=ipn:3.1 --use-bp-version-7 --outducts-config-file=$SIM_DIR/bping-outducts.json --file-or-folder-path=/tmp/hdtn-send"
echo ""
echo "Received files will appear in: $RECEIVED_DIR"
echo ""
echo "Press Ctrl+C to stop all nodes"
echo ""

cleanup() {
    echo ""
    echo "Stopping simulation..."
    kill $PID_RECV $PID_GROUND $PID_ORBITER $PID_LANDER 2>/dev/null || true
    wait $PID_RECV $PID_GROUND $PID_ORBITER $PID_LANDER 2>/dev/null || true
    echo "All nodes stopped."
}
trap cleanup INT TERM

wait
