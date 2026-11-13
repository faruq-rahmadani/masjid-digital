#!/usr/bin/env bash
# One-time setup of an Ubuntu server for masjid-digital, behind the server's existing nginx.
# Releases are then built on the server from a git checkout by deploy/release.sh.
# Run it as root (sudo) from that checkout; it is safe to run again, and keeps what exists
# (database password, JWT secret, config file).
#
#   git clone https://github.com/<owner>/masjid-digital.git ~/apps/masjid-digital
#   sudo PUBLIC_DOMAIN=masjid.example.com ADMIN_DOMAIN=admin.masjid.example.com \
#        API_DOMAIN=api.masjid.example.com DEPLOY_USER=$USER \
#        bash ~/apps/masjid-digital/deploy/setup-server.sh
#
# It creates:
#   - Go in /usr/local/go (the version go.mod asks for), to build releases
#   - system user masjid-digital, which runs the API
#   - Postgres role and database masjid_digital (local Postgres, password generated here)
#   - /etc/masjid-digital-config/config.yaml (generated secrets, readable by the service only)
#   - /opt/masjid-digital (binaries) and /var/www/masjid-digital/{public,admin}, owned by
#     DEPLOY_USER so releases need no root
#   - systemd service masjid-digital on 127.0.0.1:$API_PORT
#   - sudoers rule: DEPLOY_USER may restart that one service without a password
#   - nginx site for the three domains (HTTP; then run with HTTPS=1 for certificates)
#   - daily pg_dump to /var/backups/masjid-digital, kept 14 days
#
# HTTPS=1 additionally asks Let's Encrypt for certificates with certbot (the domains' DNS
# must already point here; it accepts Let's Encrypt's terms). CERTBOT_EMAIL is optional.
set -euo pipefail

: "${PUBLIC_DOMAIN:?set PUBLIC_DOMAIN}"
: "${ADMIN_DOMAIN:?set ADMIN_DOMAIN}"
: "${API_DOMAIN:?set API_DOMAIN}"
: "${DEPLOY_USER:?set DEPLOY_USER (the SSH user that uploads releases)}"
API_PORT="${API_PORT:-3100}"
HTTPS="${HTTPS:-0}"

[[ $EUID -eq 0 ]] || { echo "run as root: sudo ... bash $0" >&2; exit 1; }
id "$DEPLOY_USER" >/dev/null || { echo "no user $DEPLOY_USER" >&2; exit 1; }
if ss -ltnH "sport = :$API_PORT" | grep -q . && ! systemctl is-active --quiet masjid-digital; then
  echo "port $API_PORT is taken by something else; pick another API_PORT" >&2
  exit 1
fi

step() { printf '\n== %s\n' "$*"; }

step "go"
GO_WANT="$(sed -n 's/^go \([0-9.]*\)$/\1/p' "$(dirname "$0")/../go.mod")"
GO_HAVE="$(/usr/local/go/bin/go env GOVERSION 2>/dev/null | sed 's/^go//' || true)"
if [[ -z $GO_HAVE ]] || [[ "$(printf '%s\n%s\n' "$GO_WANT" "$GO_HAVE" | sort -V | head -1)" != "$GO_WANT" ]]; then
  # newest stable release of go.mod's minor version, checked against its published sha256
  read -r GO_FILE GO_SHA < <(curl -fsSL 'https://go.dev/dl/?mode=json' | python3 -c '
import json, sys
minor = ".".join(sys.argv[1].split(".")[:2])
for r in json.load(sys.stdin):
    if r["stable"] and (r["version"] + ".").startswith("go" + minor + "."):
        f = next(f for f in r["files"] if f["os"] == "linux" and f["arch"] == "amd64" and f["kind"] == "archive")
        print(f["filename"], f["sha256"])
        break
' "$GO_WANT")
  curl -fsSL -o "/tmp/$GO_FILE" "https://go.dev/dl/$GO_FILE"
  echo "$GO_SHA  /tmp/$GO_FILE" | sha256sum -c --quiet
  rm -rf /usr/local/go && tar -C /usr/local -xzf "/tmp/$GO_FILE" && rm "/tmp/$GO_FILE"
fi
/usr/local/go/bin/go version

step "system user and folders"
id masjid-digital >/dev/null 2>&1 || useradd --system --no-create-home --shell /usr/sbin/nologin masjid-digital
install -d -o "$DEPLOY_USER" -g "$DEPLOY_USER" -m 755 /opt/masjid-digital
install -d -o "$DEPLOY_USER" -g "$DEPLOY_USER" -m 755 /var/www/masjid-digital /var/www/masjid-digital/public /var/www/masjid-digital/admin
install -d -o postgres -g postgres -m 700 /var/backups/masjid-digital
install -d -o root -g masjid-digital -m 750 /etc/masjid-digital-config
cat >/opt/masjid-digital/release.env <<EOF
# Written by deploy/setup-server.sh, read by deploy/release.sh
PUBLIC_DOMAIN=$PUBLIC_DOMAIN
ADMIN_DOMAIN=$ADMIN_DOMAIN
API_DOMAIN=$API_DOMAIN
API_PORT=$API_PORT
EOF
chown "$DEPLOY_USER:$DEPLOY_USER" /opt/masjid-digital/release.env

step "postgres role and database"
CONFIG=/etc/masjid-digital-config/config.yaml
if [[ ! -f $CONFIG ]]; then
  DB_PASSWORD="$(openssl rand -hex 24)"
  if sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='masjid_digital'" | grep -q 1; then
    sudo -u postgres psql -q -c "ALTER ROLE masjid_digital WITH LOGIN PASSWORD '$DB_PASSWORD'"
  else
    sudo -u postgres psql -q -c "CREATE ROLE masjid_digital WITH LOGIN PASSWORD '$DB_PASSWORD'"
  fi
  sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='masjid_digital'" | grep -q 1 ||
    sudo -u postgres createdb -O masjid_digital masjid_digital

  step "config file (new secrets)"
  umask 027
  cat >"$CONFIG" <<EOF
# Written by deploy/setup-server.sh. Every option: config.yaml.example in the repo.
app:
  name: masjid-digital
  http_server:
    listen_address: 127.0.0.1
    port: "$API_PORT"
    graceful_timeout: 10s
    read_timeout: 30s
    write_timeout: 30s
    idle_timeout: 120s
    allowed_origins: ["https://$ADMIN_DOMAIN"]
  logger:
    level: info
    debug: false
  database:
    driver: postgres
    dsn: postgres://masjid_digital:$DB_PASSWORD@127.0.0.1:5432/masjid_digital?sslmode=disable
    auto_migrate: true
  auth:
    jwt_secret: "$(openssl rand -base64 48 | tr -d '\n')"
    token_ttl: 24h
  cache:
    driver: memory
client: {}
EOF
  chown root:masjid-digital "$CONFIG"
  chmod 640 "$CONFIG"
  unset DB_PASSWORD
else
  echo "$CONFIG exists, kept as it is"
fi

step "systemd service"
cat >/etc/systemd/system/masjid-digital.service <<EOF
[Unit]
Description=masjid-digital API
After=network-online.target postgresql.service
Wants=network-online.target

[Service]
User=masjid-digital
Group=masjid-digital
Environment=ENV=production
ExecStart=/opt/masjid-digital/http
Restart=always
RestartSec=3
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable masjid-digital >/dev/null
[[ -x /opt/masjid-digital/http ]] && systemctl restart masjid-digital

step "sudoers: $DEPLOY_USER may restart the service"
SUDOERS=/etc/sudoers.d/masjid-digital
echo "$DEPLOY_USER ALL=(root) NOPASSWD: /usr/bin/systemctl restart masjid-digital, /usr/bin/systemctl status masjid-digital" >"$SUDOERS.tmp"
visudo -cf "$SUDOERS.tmp" >/dev/null && chmod 440 "$SUDOERS.tmp" && mv "$SUDOERS.tmp" "$SUDOERS"

step "nginx site"
SITE=/etc/nginx/sites-available/masjid-digital.conf
if [[ ! -f $SITE ]] || ! grep -q "managed by Certbot" "$SITE"; then
  cat >"$SITE" <<EOF
# Written by deploy/setup-server.sh

# public site and docs (web_public/)
server {
    listen 80;
    listen [::]:80;
    server_name $PUBLIC_DOMAIN;
    root /var/www/masjid-digital/public;
    location / { try_files \$uri \$uri/ \$uri.html =404; }
}

# admin panel and TV (web/), a single-page app
server {
    listen 80;
    listen [::]:80;
    server_name $ADMIN_DOMAIN;
    root /var/www/masjid-digital/admin;
    location /assets/ { add_header Cache-Control "public, max-age=31536000, immutable"; try_files \$uri =404; }
    location / { add_header Cache-Control "no-cache"; try_files \$uri /index.html; }
}

# API
server {
    listen 80;
    listen [::]:80;
    server_name $API_DOMAIN;
    client_max_body_size 7m; # largest upload is a 6 MB slide

    location / {
        proxy_pass http://127.0.0.1:$API_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # the TV's live connection (server-sent events): unbuffered and long-lived
    location ~ ^/api/public/v1/screens/[^/]+/events\$ {
        proxy_pass http://127.0.0.1:$API_PORT;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 1h;
    }
}
EOF
else
  echo "$SITE already has certificates (managed by Certbot), kept as it is"
fi
ln -sf "$SITE" /etc/nginx/sites-enabled/masjid-digital.conf
nginx -t
systemctl reload nginx

step "daily database backup"
cat >/etc/cron.d/masjid-digital-backup <<'EOF'
# Written by deploy/setup-server.sh: daily dump, kept 14 days
30 2 * * * postgres pg_dump -Fc masjid_digital > /var/backups/masjid-digital/masjid_digital-$(date +\%F).dump && find /var/backups/masjid-digital -name '*.dump' -mtime +14 -delete
EOF
chmod 644 /etc/cron.d/masjid-digital-backup

if [[ $HTTPS == 1 ]]; then
  step "certificates (Let's Encrypt)"
  EMAIL_ARGS=(--register-unsafely-without-email)
  [[ -n ${CERTBOT_EMAIL:-} ]] && EMAIL_ARGS=(-m "$CERTBOT_EMAIL")
  certbot --nginx --non-interactive --agree-tos --redirect "${EMAIL_ARGS[@]}" \
    -d "$PUBLIC_DOMAIN" -d "$ADMIN_DOMAIN" -d "$API_DOMAIN"
fi

step "done"
systemctl --no-pager --lines=0 status masjid-digital || true
echo "next: as $DEPLOY_USER, run deploy/release.sh from the checkout"
