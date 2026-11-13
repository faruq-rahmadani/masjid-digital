# Deploy

masjid-digital runs on one Ubuntu server (24.04 tested) behind nginx:

| Domain (example) | What | Served by |
| --- | --- | --- |
| `masjid.example.com` | public site and docs (`web_public/`) | nginx, static files |
| `admin.masjid.example.com` | admin panel and TV at `/tv` (`web/`) | nginx, single-page app |
| `api.masjid.example.com` | API (`cmd/http`) | systemd service on `127.0.0.1:3100`, nginx in front |

The database is the server's local PostgreSQL. Releases are built on the server from a git
checkout of `main`. The API is one long-running process: the TV keeps a server-sent events
stream open and its events live in that process's memory, so run exactly one.

## Once: prepare the server

1. Point the three domains (DNS `A` records) at the server.
2. Clone the repo as the user that will deploy (no root needed):

   ```sh
   git clone https://github.com/<owner>/masjid-digital.git ~/apps/masjid-digital
   ```

3. Run the setup as root. It installs Go, creates the Postgres database, a config with fresh
   secrets, the systemd service, the nginx site and a daily backup. Read
   [`deploy/setup-server.sh`](../deploy/setup-server.sh) first; it is safe to run again.

   ```sh
   sudo PUBLIC_DOMAIN=masjid.example.com ADMIN_DOMAIN=admin.masjid.example.com \
        API_DOMAIN=api.masjid.example.com DEPLOY_USER=$USER \
        bash ~/apps/masjid-digital/deploy/setup-server.sh
   ```

4. Ship the first release (below), then turn on HTTPS once DNS resolves: the same command
   with `HTTPS=1` (and optionally `CERTBOT_EMAIL=...`) in front of `bash`.
5. Create the first super admin. The password is asked on the terminal:

   ```sh
   sudo -u masjid-digital ENV=production /opt/masjid-digital/admin \
        create-super-admin -name "Nama Lengkap" -phone 081234567890
   ```

## Every release

Merge to `main`, then on the server:

```sh
~/apps/masjid-digital/deploy/release.sh
```

It resets the checkout to `origin/main`, builds the API and both web apps, swaps them in,
restarts the API (migrations run on start) and waits for `/health`. `ONLY=api|admin|public`
ships one part; `BRANCH=...` ships another branch.

## Where things are

| Path | What |
| --- | --- |
| `/etc/masjid-digital-config/config.yaml` | API config with the database password and JWT secret (root and the service only) |
| `/opt/masjid-digital/` | API binaries (`http`, `migrate`, `admin`) and `release.env` |
| `/var/www/masjid-digital/{public,admin}` | built web apps |
| `/var/backups/masjid-digital/` | daily `pg_dump`, kept 14 days |
| `/etc/nginx/sites-available/masjid-digital.conf` | nginx site |

Logs: `sudo journalctl -u masjid-digital -f`. Forgotten password:
`sudo -u masjid-digital ENV=production /opt/masjid-digital/admin reset-password -phone 0812...`.
Restore a backup: `sudo -u postgres pg_restore --clean -d masjid_digital <file>.dump`.

## Other hosts

The [`Dockerfile`](../Dockerfile) builds the API as a container configured by environment
variables: every key of `config.yaml.example` is its path in capitals with `_`
(`app.database.dsn` is `APP_DATABASE_DSN`), and `PORT` sets the port. Web clients on another
origin go in `APP_HTTP_SERVER_ALLOWED_ORIGINS`.
