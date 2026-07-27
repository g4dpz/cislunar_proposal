#!/bin/bash
# Package radiant-ion for deployment to an Ubuntu VM.
# Creates a tarball with source, dependencies, and install scripts.
# Usage: bash deploy/radiant-ion/package.sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
OUTPUT="$PROJECT_DIR/dist/radiant-ion-${TIMESTAMP}.tar.gz"

echo "=== Packaging radiant-ion for deployment ==="

mkdir -p "$PROJECT_DIR/dist"

cd "$PROJECT_DIR"

tar czf "$OUTPUT" \
    --exclude='*/target' \
    --exclude='*/.git' \
    --exclude='*.DS_Store' \
    radiant-ion/Cargo.toml \
    radiant-ion/Cargo.lock \
    radiant-ion/src/ \
    radiant-ion/examples/ \
    radiant-ion/README.md \
    radiant-dtn-abstraction/Cargo.toml \
    radiant-dtn-abstraction/src/ \
    deploy/radiant-ion/

echo ""
echo "Package: $OUTPUT"
echo "Size: $(du -h "$OUTPUT" | cut -f1)"
echo ""
echo "Deploy to your VM:"
echo "  scp $OUTPUT user@server:/tmp/"
echo "  ssh user@server"
echo "  cd /opt && sudo tar xzf /tmp/$(basename $OUTPUT)"
echo "  cd radiant-ion && sudo bash deploy/radiant-ion/install.sh"
