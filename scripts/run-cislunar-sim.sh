#!/bin/bash
# Three-node cislunar DTN simulation using HDTN
#
# Topology:
#   bpsendfile (ipn:1.1)
#     → [STCP :4556] → Ground Station HDTN (nodeId=10)
#     → [LTP/UDP, 1300ms OWLT, 500 bps] → Orbiter HDTN (nodeId=20)
#     → [LTP/UDP, 10ms OWLT, 9600 bps] → Lander HDTN (nodeId=30)
#     → [STCP :4558] → bpreceivefile (ipn:3.1)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SIM_DIR="$PROJECT_DIR/configs/simulation"
RECEIVED_DIR="$PROJECT_DIR/cislunar_received"

echo "=== Cislunar DTN Simulation (HDTN) ==="
echo ""
echo "Topology:"
echo "  Ground (10) →[1.3s OWLT, 500bps]→ Orbiter (20) →[10ms, 9.6kbps]→ Lander (30)"
echo ""

mkdir -p "$RECEIVED_DIR"

if ! command -v hdtn-one-process &> /dev/null; then
    echo "Error: hdtn-one-process not found in PATH"
    exit 1
fi

echo "Starting nodes..."

bpreceivefile \
    --my-uri-eid=ipn:3.1 \
    --inducts-config-file="$SIM_DIR/bpsink-inducts.json" \
    --save-directory="$RECEIVED_DIR" \
    --max-rx-bundle-size-bytes=10000000 &
PID_RECV=$!
sleep 2

hdtn-one-process \
    --hdtn-config-file="$SIM_DIR/cislunar-lander.json" \
    --contact-plan-file="$SIM_DIR/cislunar-contact-plan.json" &
PID_LANDER=$!
sleep 8

hdtn-one-process \
    --hdtn-config-file="$SIM_DIR/cislunar-orbiter.json" \
    --contact-plan-file="$SIM_DIR/cislunar-contact-plan.json" &
PID_ORBITER=$!
sleep 8

hdtn-one-process \
    --hdtn-config-file="$SIM_DIR/cislunar-ground-station.json" \
    --contact-plan-file="$SIM_DIR/cislunar-contact-plan.json" &
PID_GROUND=$!
sleep 8

echo ""
echo "=== Simulation Ready ==="
echo ""
echo "Send a file:"
echo "  rm -f /tmp/hdtn-send/*"
echo "  echo 'Hello Moon' > /tmp/hdtn-send/lunar.dat"
echo "  bpsendfile --my-uri-eid=ipn:1.1 --dest-uri-eid=ipn:3.1 --use-bp-version-7 --outducts-config-file=$SIM_DIR/bping-outducts.json --file-or-folder-path=/tmp/hdtn-send"
echo ""
echo "Received files: $RECEIVED_DIR"
echo "Press Ctrl+C to stop"

cleanup() {
    echo ""
    echo "Stopping..."
    kill $PID_RECV $PID_GROUND $PID_ORBITER $PID_LANDER 2>/dev/null || true
    wait $PID_RECV $PID_GROUND $PID_ORBITER $PID_LANDER 2>/dev/null || true
    echo "Done."
}
trap cleanup INT TERM

wait
