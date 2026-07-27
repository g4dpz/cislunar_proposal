#!/bin/bash
# api-demo.sh — Demonstrates the radiant-ion HTTP API using curl.
#
# Prerequisites:
#   radiant-ion serve examples/g4dpz-ground-station.yaml --port 3000
#
# Run this script in a separate terminal while the server is running.

BASE_URL="http://localhost:3000"

echo "=== radiant-ion API Demo ==="
echo "  Server: $BASE_URL"
echo

# 1. Health check
echo "[1] GET /lifecycle/health"
curl -s "$BASE_URL/lifecycle/health" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/lifecycle/health"
echo
echo

# 2. Get engine state
echo "[2] GET /lifecycle/state"
curl -s "$BASE_URL/lifecycle/state" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/lifecycle/state"
echo
echo

# 3. Get current config
echo "[3] GET /config"
curl -s "$BASE_URL/config" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/config"
echo
echo

# 4. Get bundle statistics
echo "[4] GET /stats"
curl -s "$BASE_URL/stats" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/stats"
echo
echo

# 5. Get link states
echo "[5] GET /stats/links"
curl -s "$BASE_URL/stats/links" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/stats/links"
echo
echo

# 6. Get capabilities
echo "[6] GET /capabilities"
curl -s "$BASE_URL/capabilities" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/capabilities"
echo
echo

# 7. Stop the engine
echo "[7] POST /lifecycle/stop"
curl -s -X POST "$BASE_URL/lifecycle/stop" | python3 -m json.tool 2>/dev/null || curl -s -X POST "$BASE_URL/lifecycle/stop"
echo
echo

# 8. Verify stopped
echo "[8] GET /lifecycle/health (after stop)"
curl -s "$BASE_URL/lifecycle/health" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/lifecycle/health"
echo
echo

# 9. Start the engine again
echo "[9] POST /lifecycle/start"
curl -s -X POST "$BASE_URL/lifecycle/start" | python3 -m json.tool 2>/dev/null || curl -s -X POST "$BASE_URL/lifecycle/start"
echo
echo

# 10. Verify running
echo "[10] GET /lifecycle/health (after restart)"
curl -s "$BASE_URL/lifecycle/health" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/lifecycle/health"
echo
echo

echo "=== Done ==="
