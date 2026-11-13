#!/usr/bin/env bash
# Builds and installs a release ON the server, from the git checkout this script lives in.
# The server must be prepared once by setup-server.sh, which writes the settings this
# script reads (/opt/masjid-digital/release.env). Run it as the deploy user, no root:
#
#   ~/apps/masjid-digital/deploy/release.sh            # latest origin/main
#   BRANCH=feat/x ONLY=admin deploy/release.sh          # another branch, or one part
#
# ONLY=api|admin|public ships one part. The checkout is reset to origin/$BRANCH, so keep
# no local changes in it.
set -euo pipefail

# The braces make bash read the whole script before running it: the checkout below may
# replace this very file.
{
# shellcheck source=/dev/null
source /opt/masjid-digital/release.env # PUBLIC_DOMAIN ADMIN_DOMAIN API_DOMAIN API_PORT
BRANCH="${BRANCH:-main}"
ONLY="${ONLY:-all}"
export PATH="/usr/local/go/bin:$PATH"
# Node from nvm when the deploy user has it
if [[ -s ${NVM_DIR:-$HOME/.nvm}/nvm.sh ]]; then
  # shellcheck source=/dev/null
  source "${NVM_DIR:-$HOME/.nvm}/nvm.sh" >/dev/null
fi

cd "$(dirname "$0")/.."
git fetch --quiet origin "$BRANCH"
git checkout --quiet -B "$BRANCH" "origin/$BRANCH"
VERSION="$(git describe --always)"
OUT="$(mktemp -d)"
trap 'rm -rf "$OUT"' EXIT
want() { [[ $ONLY == all || $ONLY == "$1" ]]; }
echo "== $BRANCH at $VERSION"

if want api; then
  echo "== api"
  for cmd in http migrate admin; do
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$OUT/api/$cmd" "./cmd/$cmd"
  done
  # swap each binary in with one rename, then restart
  for cmd in http migrate admin; do
    install -m 755 "$OUT/api/$cmd" "/opt/masjid-digital/.$cmd.new"
    mv -f "/opt/masjid-digital/.$cmd.new" "/opt/masjid-digital/$cmd"
  done
  sudo -n /usr/bin/systemctl restart masjid-digital
fi

if want admin; then
  echo "== admin panel + TV"
  (cd web && npm ci --silent --no-audit --no-fund &&
    VITE_API_BASE_URL="https://$API_DOMAIN/api" VITE_APP_VERSION="$VERSION" \
      npx vite build --outDir "$OUT/admin" --emptyOutDir --logLevel warn)
  rsync -a --delete "$OUT/admin/" /var/www/masjid-digital/admin/
fi

if want public; then
  echo "== public site"
  (cd web_public && npm ci --silent --no-audit --no-fund &&
    VITE_DEMO_URL="https://$ADMIN_DOMAIN" \
      npx vite build --outDir "$OUT/public" --emptyOutDir --logLevel warn)
  rsync -a --delete "$OUT/public/" /var/www/masjid-digital/public/
fi

if want api; then
  echo "== health"
  for _ in $(seq 1 20); do
    curl -fsS -o /dev/null "http://127.0.0.1:$API_PORT/health" 2>/dev/null && { echo "API is up ($VERSION)"; exit 0; }
    sleep 2
  done
  echo "API did not answer /health; see: sudo journalctl -u masjid-digital -n 50" >&2
  exit 1
fi
exit 0
}
