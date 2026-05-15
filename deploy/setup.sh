#!/bin/bash
# RADIANT Website — Service Setup Script
# Run as root on the target Linux server.
# Usage: sudo bash setup.sh

set -e

REPO_URL="https://github.com/g4dpz/cislunar_proposal.git"
INSTALL_DIR="/opt/radiant"
SERVICE_NAME="radiant"
APACHE_SITE="radiant"
DOMAIN="cislunar-project.amsat-uk.org"

echo "=== RADIANT Service Setup ==="
echo ""

# ─── Check root ────────────────────────────────────────────────────────────────

if [ "$EUID" -ne 0 ]; then
  echo "Error: Please run as root (sudo bash setup.sh)"
  exit 1
fi

# ─── Install Deno if not present ───────────────────────────────────────────────

if ! command -v deno &> /dev/null; then
  echo "Installing Deno..."
  curl -fsSL https://deno.land/install.sh | DENO_INSTALL=/usr/local sh
  echo "Deno installed at /usr/local/bin/deno"
else
  echo "Deno already installed: $(deno --version | head -1)"
fi

# ─── Enable Apache modules ────────────────────────────────────────────────────

echo "Enabling Apache modules..."
a2enmod proxy proxy_http ssl rewrite headers > /dev/null 2>&1 || true

# ─── Clone or update repository ───────────────────────────────────────────────

if [ -d "$INSTALL_DIR/repo" ]; then
  echo "Updating existing repository..."
  cd "$INSTALL_DIR/repo"
  git pull
else
  echo "Cloning repository..."
  mkdir -p "$INSTALL_DIR"
  git clone "$REPO_URL" "$INSTALL_DIR/repo"
fi

# ─── Create data directory ─────────────────────────────────────────────────────

mkdir -p "$INSTALL_DIR/data"

# ─── Set permissions ───────────────────────────────────────────────────────────

chown -R www-data:www-data "$INSTALL_DIR"

# ─── Cache Deno dependencies ──────────────────────────────────────────────────

echo "Caching Deno dependencies..."
cd "$INSTALL_DIR/repo/website"
DENO_DIR="$INSTALL_DIR/.deno" deno cache main.ts

# ─── Install systemd service ──────────────────────────────────────────────────

echo "Installing systemd service..."
cp "$INSTALL_DIR/repo/deploy/radiant.service" /etc/systemd/system/
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"

# ─── Install Apache site config ────────────────────────────────────────────────

echo "Installing Apache site config..."
cp "$INSTALL_DIR/repo/deploy/apache-radiant.conf" "/etc/apache2/sites-available/${APACHE_SITE}.conf"
a2ensite "$APACHE_SITE" > /dev/null 2>&1 || true

# ─── Start the service ─────────────────────────────────────────────────────────

echo "Starting RADIANT service..."
systemctl restart "$SERVICE_NAME"

# ─── Restart Apache ────────────────────────────────────────────────────────────

echo "Restarting Apache..."
systemctl restart apache2

# ─── Done ──────────────────────────────────────────────────────────────────────

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Service status:"
systemctl status "$SERVICE_NAME" --no-pager -l | head -5
echo ""
echo "Site: https://$DOMAIN"
echo "Default admin: admin@radiant.radio / admin123!"
echo ""
echo "IMPORTANT: Change the default admin password immediately after first login."
echo ""
echo "To view logs:  journalctl -u $SERVICE_NAME -f"
echo "To restart:    systemctl restart $SERVICE_NAME"
echo "To update:     cd $INSTALL_DIR/repo && git pull && systemctl restart $SERVICE_NAME"
