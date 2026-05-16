# Deployment Guide — RADIANT Website

## Prerequisites

- Linux server (Debian/Ubuntu)
- Apache 2.4+ with `mod_proxy`, `mod_proxy_http`, `mod_ssl`, `mod_rewrite`, `mod_headers`
- Deno runtime installed (`curl -fsSL https://deno.land/install.sh | sh`)
- Certbot for Let's Encrypt TLS certificates

## 1. Install Deno

```bash
curl -fsSL https://deno.land/install.sh | sh
sudo ln -s ~/.deno/bin/deno /usr/bin/deno
```

## 2. Deploy the application

```bash
sudo mkdir -p /opt/radiant/data
sudo git clone https://github.com/g4dpz/cislunar_proposal.git /opt/radiant/repo
sudo ln -s /opt/radiant/repo/website /opt/radiant/website
sudo chown -R www-data:www-data /opt/radiant
```

## 3. Install the systemd service

```bash
sudo cp /opt/radiant/repo/deploy/radiant.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable radiant
sudo systemctl start radiant
```

Check status:
```bash
sudo systemctl status radiant
sudo journalctl -u radiant -f
```

## 4. Configure Apache

Enable required modules:
```bash
sudo a2enmod proxy proxy_http ssl rewrite headers
```

Install the site config:
```bash
sudo cp /opt/radiant/repo/deploy/apache-radiant.conf /etc/apache2/sites-available/radiant.conf
sudo a2ensite radiant
```

## 5. Set up TLS with Let's Encrypt

```bash
sudo certbot certonly --webroot -w /var/www/html -d radiant.amsat-uk.org
```

Or if Apache is already running:
```bash
sudo certbot --apache -d radiant.amsat-uk.org
```

## 6. Restart Apache

```bash
sudo systemctl restart apache2
```

## 7. Verify

```bash
curl -I https://radiant.amsat-uk.org
```

## Updating

```bash
cd /opt/radiant/repo
sudo git pull
sudo systemctl restart radiant
```

## Default Admin Login

- Email: `admin@radiant.radio`
- Password: `admin123!`

**Change this immediately after first login.**

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8000` | HTTP port for the Deno server |
| `DB_PATH` | `./data/radiant.db` | Path to SQLite database file |
| `DENO_DIR` | system default | Deno cache directory |
