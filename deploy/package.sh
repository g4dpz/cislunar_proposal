#!/bin/bash
# RADIANT Website — Package for deployment
# Creates a zip file containing only the website and deploy files.
# Usage: bash deploy/package.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="$PROJECT_DIR/dist"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
ZIP_NAME="radiant-website-${TIMESTAMP}.zip"

echo "=== Packaging RADIANT Website ==="

# Create dist directory
mkdir -p "$OUTPUT_DIR"

# Create zip from project root, including only website/ and deploy/
cd "$PROJECT_DIR"
zip -r "$OUTPUT_DIR/$ZIP_NAME" \
  website/main.ts \
  website/deno.json \
  website/content/ \
  website/db/ \
  website/email/ \
  website/middleware/ \
  website/routes/ \
  website/services/ \
  website/views/ \
  website/public/ \
  deploy/radiant.service \
  deploy/apache-radiant.conf \
  deploy/setup.sh \
  deploy/deploy.sh \
  deploy/README.md \
  -x "website/data/*" \
  -x "website/tests/*" \
  -x "*.db" \
  -x "*.db-shm" \
  -x "*.db-wal"

echo ""
echo "Package created: dist/$ZIP_NAME"
echo "Size: $(du -h "$OUTPUT_DIR/$ZIP_NAME" | cut -f1)"
echo ""
echo "To deploy, upload to the server and run:"
echo "  unzip $ZIP_NAME -d /opt/radiant"
echo "  sudo bash /opt/radiant/deploy/deploy.sh"
