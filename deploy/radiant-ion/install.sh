#!/bin/bash
# radiant-ion Installer
# Builds from source and puts the binary in ./bin/
# Usage: bash deploy/radiant-ion/install.sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SRC_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

if [ ! -f "$SRC_DIR/radiant-ion/Cargo.toml" ]; then
    echo "Error: Cannot find radiant-ion/Cargo.toml"
    exit 1
fi

echo "=== radiant-ion Installer ==="
echo ""

# ─── Verify ION-DTN ──────────────────────────────────────────────────────────

echo "[1/3] Checking ION-DTN..."
if ! command -v ionadmin &>/dev/null; then
    echo "  ERROR: ION-DTN not found on PATH."
    exit 1
fi
echo "  Found: $(which ionadmin)"

# ─── Check Rust ───────────────────────────────────────────────────────────────

echo "[2/3] Checking Rust..."
if ! command -v cargo &>/dev/null; then
    echo "  Rust not found. Installing..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source "$HOME/.cargo/env"
fi
echo "  $(rustc --version)"

# ─── Build ────────────────────────────────────────────────────────────────────

echo "[3/3] Building..."
cd "$SRC_DIR/radiant-ion"
cargo build --release 2>&1 | tail -3

mkdir -p "$SRC_DIR/bin"
cp target/release/radiant-ion "$SRC_DIR/bin/"

echo ""
echo "=== Done ==="
echo ""
echo "Binary: $SRC_DIR/bin/radiant-ion"
echo ""
echo "Usage:"
echo "  ./bin/radiant-ion serve examples/g4dpz-ground-station.yaml --port 3000"
echo "  ./bin/radiant-ion start examples/g4dpz-ground-station.yaml"
echo "  ./bin/radiant-ion status"
echo "  ./bin/radiant-ion stop"
echo ""
