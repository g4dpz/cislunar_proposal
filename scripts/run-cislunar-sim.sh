#!/bin/bash
# Three-node cislunar DTN simulation with TRUE packet-level delay
# Uses HDTN's udp-delay-sim to add real 1.3-second propagation delay
#
# Port mapping:
#   Ground (2113) → proxy:1114 → [1300ms delay] → Orbiter:1113
#   Orbiter reports → proxy:1115 → [1300ms delay] → Ground:2113

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SIM_DIR="$PROJECT_DIR/configs/simulation"
RECEIVED_DIR="$PROJECT_DIR/cislunar_received"
DELAY_MS=1300

echo "=== Cislunar DTN with TRUE ${DELAY_MS}ms Packet-Level Delay ==="
echo ""
echo "  Ground (2113) → proxy:1114 →[${DELAY_MS}ms]→ Orbiter:1113"
echo "  Orbiter reports → proxy:1115 →[${DELAY_MS}ms]→ Ground:2113"
echo ""

mkdir -p "$RECEIVED_DIR"

# Start delay proxies
echo "Starting delay proxies..."

# Data proxy: receives on 1114, delays, forwards to orbiter 1113
udp-delay-sim \
    --remote-udp-hostname=localhost --remote-udp-port=1113 \
    --my-bound-udp-port=1114 \
    --num-rx-udp-packets-buffer-size=10000 \
    --max-rx-udp-packet-size-bytes=1500 \
    --send-delay-ms=$DELAY_MS &
PID_PROXY_DATA=$!
echo "  [Data proxy]   :1114 → :1113 (${DELAY_MS}ms) PID $PID_PROXY_DATA"

# Report proxy: receives on 1115, delays, forwards to ground 2113
udp-delay-sim \
    --remote-udp-hostname=localhost --remote-udp-port=2113 \
    --my-bound-udp-port=1115 \
    --num-rx-udp-packets-buffer-size=10000 \
    --max-rx-udp-packet-size-bytes=1500 \
    --send-delay-ms=$DELAY_MS &
PID_PROXY_REPORT=$!
echo "  [Report proxy] :1115 → :2113 (${DELAY_MS}ms) PID $PID_PROXY_REPORT"
sleep 1

echo ""
echo "Starting HDTN nodes..."

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
echo "=== Ready — TRUE ${DELAY_MS}ms propagation delay active ==="
echo ""
echo "Send a file:"
echo "  rm -f /tmp/hdtn-send/*; echo 'Hello Moon' > /tmp/hdtn-send/lunar.dat"
echo "  bpsendfile --my-uri-eid=ipn:1.1 --dest-uri-eid=ipn:3.1 --use-bp-version-7 --outducts-config-file=$SIM_DIR/bping-outducts.json --file-or-folder-path=/tmp/hdtn-send"
echo ""
echo "File arrives after ~1.3s real delay. Received: $RECEIVED_DIR"
echo "Press Ctrl+C to stop"

cleanup() {
    echo ""
    echo "Stopping..."
    kill $PID_RECV $PID_GROUND $PID_ORBITER $PID_LANDER $PID_PROXY_DATA $PID_PROXY_REPORT 2>/dev/null || true
    wait $PID_RECV $PID_GROUND $PID_ORBITER $PID_LANDER $PID_PROXY_DATA $PID_PROXY_REPORT 2>/dev/null || true
    echo "Done."
}
trap cleanup INT TERM

wait
