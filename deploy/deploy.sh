#!/bin/bash
# RADIANT Website — Deploy from zip package
# Run on the target server after uploading and extracting the zip.
# Usage: sudo bash deploy.sh

set -e

INSTALL_DIR="/opt/radiant"
SERVICE_NAME="radiant"
APACHE_SITE="radiant"

echo "=== Deploying RADIANT Website ==="
echo ""

if [ "$EUID" -ne 0 ]; then
  echo "Error: Please run as root (sudo bash deploy.sh)"
  exit 1
fi

# ─── Detect script location (inside extracted zip) ─────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PACKAGE_DIR="$(dirname "$SCRIPT_DIR")"

echo "Deploying from: $PACKAGE_DIR"

# ─── Install Deno if not present ───────────────────────────────────────────────

if ! command -v deno &> /dev/null; then
  echo "Installing Deno..."
  curl -fsSL https://deno.land/install.sh | DENO_INSTALL=/usr/local sh
else
  echo "Deno already installed: $(deno --version | head -1)"
fi

# ─── Copy website files ───────────────────────────────────────────────────────

echo "Copying website files..."
mkdir -p "$INSTALL_DIR/website"
mkdir -p "$INSTALL_DIR/data"

# Sync website directory (preserve data)
rm -rf "$INSTALL_DIR/website.tmp"
cp -a "$PACKAGE_DIR/website" "$INSTALL_DIR/website.tmp"
rm -rf "$INSTALL_DIR/website.tmp/data" "$INSTALL_DIR/website.tmp/tests"
rm -rf "$INSTALL_DIR/website"
mv "$INSTALL_DIR/website.tmp" "$INSTALL_DIR/website"

echo "Website files deployed to $INSTALL_DIR/website/"

# ─── Set permissions ───────────────────────────────────────────────────────────

chown -R www-data:www-data "$INSTALL_DIR"

# ─── Cache Deno dependencies ──────────────────────────────────────────────────

echo "Caching Deno dependencies..."
cd "$INSTALL_DIR/website"
DENO_DIR="$INSTALL_DIR/.deno" deno cache main.ts

# ─── Install systemd service (if not already installed or if updated) ──────────

echo "Installing systemd service..."
cp "$PACKAGE_DIR/deploy/radiant.service" /etc/systemd/system/
systemctl daemon-reload
systemctl enable "$SERVICE_NAME" > /dev/null 2>&1 || true

# ─── Install Apache config (if not already installed) ──────────────────────────

if [ ! -f "/etc/apache2/sites-available/${APACHE_SITE}.conf" ]; then
  echo "Installing Apache site config..."
  a2enmod proxy proxy_http ssl rewrite headers > /dev/null 2>&1 || true
  cp "$PACKAGE_DIR/deploy/apache-radiant.conf" "/etc/apache2/sites-available/${APACHE_SITE}.conf"
  a2ensite "$APACHE_SITE" > /dev/null 2>&1 || true
  systemctl restart apache2
else
  echo "Apache site config already exists (skipping)."
fi

# ─── Restart the service ───────────────────────────────────────────────────────

echo "Restarting RADIANT service..."
systemctl restart "$SERVICE_NAME"

# ─── Done ──────────────────────────────────────────────────────────────────────

echo ""
echo "=== Deployment Complete ==="
echo ""
systemctl status "$SERVICE_NAME" --no-pager -l | head -5
echo ""
echo "Site: https://cislunar-project.amsat-uk.org"
echo "Logs: journalctl -u $SERVICE_NAME -f"
