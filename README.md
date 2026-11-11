# masjid-digital

A digital notice board and admin panel for a mosque. Pengurus (the mosque's board) manage
prayer times, iqamah, slides, running text and the Friday duty roster from their phone; every
TV in the mosque shows it full screen and updates within seconds of a change.

Each deployment serves **one mosque** (single tenant): there is no mosque table or `mosque_id`;
the mosque's profile and logo are stored once.

![TV display](docs/screenshots/tv.png)

## Features

### TV display (`/tv`)

A public 1920×1080 page for any TV with a browser (smart TV, set-top box, mini PC). No login.

- **Main screen:** clock and Gregorian + Hijri date, today's six prayer times with the next one
  highlighted, countdown to the next prayer, slideshow, running text, a Qur'an quote and the
  kas (treasury) header: saldo awal, pemasukan, pengeluaran and saldo akhir of each kas marked
  "tampil di layar TV", one kas at a time.
- **Prayer flow:** switches by itself to an adzan screen, an iqamah countdown, then a dimmed or
  black screen while the prayer is on, and back to normal.
- **Jumat:** before Dzuhur on Friday, a Jumat screen with khatib, imam, muadzin and the khutbah
  theme, then adzan, khutbah and sholat.
- **Alarm:** a beep or bell at adzan and in the last seconds before iqamah (tap the screen once to
  allow sound).
- **Live:** every TV keeps a server-sent events connection, so a saved change reaches all of them
  within seconds; the clock follows the server's time, not the device's.

| Adzan | Iqamah countdown |
| :---: | :---: |
| ![Adzan](docs/screenshots/tv-adzan.png) | ![Iqamah countdown](docs/screenshots/tv-iqamah.png) |
| **Jumat** | **Sholat in progress** |
| ![Jumat](docs/screenshots/tv-jumat.png) | ![Sholat in progress](docs/screenshots/tv-sholat.png) |

`/tv-preview` shows the same screen with sample data and a panel to jump to any mode (adzan,
iqamah, Jumat, sholat) and speed up time, so you can check the flow without waiting for a prayer.

### Admin panel

A mobile-first web app for pengurus. Each role sees only the menus it may use (see
[Auth and roles](#auth-and-roles)).

| | | |
| :---: | :---: | :---: |
| ![Login](docs/screenshots/login.png) | ![Home](docs/screenshots/home.png) | ![Users](docs/screenshots/users.png) |
| **Login** | **Home** | **Users** |
| ![Slides](docs/screenshots/slides.png) | ![Running text](docs/screenshots/running-texts.png) | ![Prayer times](docs/screenshots/prayer-times.png) |
| **Slides** | **Running text** | **Prayer times** |
| ![Iqamah](docs/screenshots/iqamah.png) | ![Friday duties](docs/screenshots/friday-duties.png) | ![Mosque profile](docs/screenshots/profile.png) |
| **Iqamah & alarm** | **Friday duties** | **Mosque profile** |
| ![TV screens](docs/screenshots/screens.png) | | |
| **TV screens** | | |

- **Home:** next prayer with a countdown, TV status at a glance, the menu, and recent activity
  (who changed what).
- **Users** (super admin): add, edit, disable or re-enable users, reset a password, filter by role
  and search. Users are never deleted. Anyone can change their own password.
- **Slides:** upload JPG/PNG/WebP images (max 5 MB, 16:9 suggested), set how long each shows and
  whether the title and caption appear, turn slides on or off, drag to reorder.
- **Running text:** short announcements (max 200 characters) that scroll along the bottom of the
  TV; turn on or off, drag to reorder, one scroll speed for all.
- **Prayer times:** computed with the Kemenag RI method (Subuh 20°, Isya 18°) from the mosque's
  location, with a 7-day preview. Correct each prayer by ±10 minutes, set the imsak offset, fix a
  time every week (e.g. Dzuhur at 12:00 on Fridays) or on one date, and shift the Hijri date.
- **Iqamah & alarm:** minutes from adzan to iqamah and length of each prayer, how long the adzan
  screen shows, dimmed or black screen during sholat, Jumat and khutbah timings, alarm sound and
  volume. The page shows the resulting timeline for the next prayer.
- **Friday duties:** khatib, imam, muadzin and khutbah theme per Friday; the next one shows on the
  TV's Jumat screen.
- **Mosque profile:** name, address, contact number, logo, location picked on a map (search or
  "use my location") and timezone (WIB, WITA, WIT).
- **TV screens:** a TV registers itself the first time it opens `/tv` (with a QR code to open the
  link from a phone). Name it, see whether it is online and which version it runs, force a mode
  (auto, normal, Jumat, black), reload it remotely, or block a screen that isn't yours.

- **Keuangan (kas):** money is tracked by account (Tunai, bank) and by kas (Kas Umum, Zakat Mal,
  Zakat Fitrah, Fidyah, and earmarked "terikat" kas such as Renovasi Atap). Pemasukan and
  pengeluaran with a receipt photo, one receipt split over several kas, Pindah dana between kas or
  accounts, edits and deletes with a reason and history, "Mulai pembukuan" with opening balances,
  monthly tutup buku after counting each account, kas targets and "Selesaikan program". Syariah
  rules are enforced: zakat and fidyah only reach their recipients, earmarked money moves only
  with the donors' consent, and no account or kas goes below Rp 0. Warnings show when amil
  spending passes 1/8 of the zakat income or Zakat Fitrah is still held after 1 Syawal.

**Not built yet:** the TV's Qur'an quote is sample data for now.

## Repository layout

```
cmd/
  http/          # HTTP server entrypoint and dependency wiring (wire.go)
  migrate/       # database migration CLI
internal/
  config/        # config.yaml schema and loader
  domain/        # entities, request structs, validation (no framework imports)
  usecase/       # business logic; repository interfaces in interface.go
  repository/    # database and cache implementations
  delivery/http/ # Echo router, controllers, middleware, error handler
migrations/      # goose SQL migrations, one folder per driver (sqlite, postgres)
mock/            # gomock mocks, generated
pkg/             # reusable packages: apperror, cache, database, hasher, token, env
web/             # React client (see web/README.md)
web_public/      # public product site and user docs, static (see web_public/README.md)
docs/            # API spec (openapi.yaml), design notes, README screenshots
```

`internal/domain/README.md` and `internal/repository/README.md` describe the rules of each layer.

## Stack

- **Backend:** Go 1.26, Echo v4, sqlx + squirrel, goose migrations, Argon2id passwords, JWT access tokens
- **Database:** SQLite (default, zero setup) or PostgreSQL
- **Cache / sessions:** in-memory (default) or Redis-compatible (Redis, Valkey, Dragonfly, KeyDB)
- **Frontend:** Vite, React 19, TypeScript, TanStack Query, Tailwind CSS v4

## Getting started

Requirements: Go 1.26+, Node 24 (for `web/`).

```sh
cp config.yaml.example config.yaml
# set app.auth.jwt_secret (at least 32 bytes):
openssl rand -base64 32

go run ./cmd/http    # http://localhost:3000, migrations run on startup
```

Check it is up with `curl localhost:3000/health`.

The public product site with the user documentation lives in `web_public/` (static HTML; set the
demo links in its `.env`, see [web_public/README.md](web_public/README.md)).

Frontend, in another terminal:

```sh
cd web
cp .env.example .env.local
npm install
npm run dev          # http://localhost:5173, /api is proxied to the Go server
```

> There is no way to create the first `super_admin` yet: every user endpoint requires one.

## Configuration

The config file depends on `ENV`:

| `ENV`                     | Config file                              |
| ------------------------- | ---------------------------------------- |
| empty or `development`    | `config.yaml` in the project root        |
| `production`              | `/etc/masjid-digital-config/config.yaml` |

Every option is documented in [`config.yaml.example`](config.yaml.example). The main ones:

| Key                             | Description                                                      |
| ------------------------------- | ---------------------------------------------------------------- |
| `app.http_server.port`          | HTTP port, default `3000`                                        |
| `app.database.driver` / `dsn`   | `sqlite` (`masjid.db`) or `postgres` (`postgres://...`)          |
| `app.database.migrate_dsn`      | optional separate user for schema changes; empty uses `dsn`      |
| `app.database.auto_migrate`     | apply pending migrations on startup                              |
| `app.auth.jwt_secret`           | required, at least 32 bytes                                      |
| `app.auth.token_ttl`            | access token lifetime, e.g. `24h`                                |
| `app.cache.driver` / `dsn`      | `memory` or `redis` (`redis://:password@localhost:6379/0`)       |

With the `memory` cache, sessions are lost on restart (everyone logs in again) and are not shared
between instances; use Redis when running more than one.

## Database migrations

Migrations live in `migrations/sqlite` and `migrations/postgres`; add every change to both.
With `auto_migrate: false`, run them yourself:

```sh
go run ./cmd/migrate up       # apply pending migrations
go run ./cmd/migrate down     # roll back the latest one
go run ./cmd/migrate status   # list migrations and whether they are applied
go run ./cmd/migrate version  # print the current schema version
```

## Auth and roles

`POST /api/v1/auth/login` with a phone number (E.164) or email returns a bearer token. A user has
one session at a time: logging in elsewhere, logging out, being disabled, a role change or an
admin password reset ends it.

| Role          | Can                                                          |
| ------------- | ------------------------------------------------------------ |
| `super_admin` | everything, including user management                        |
| `takmir`      | run the mosque's operations, edit the mosque profile and logo |
| `bendahara`   | record kas income and expenses                               |

Users are never deleted, only disabled, so records keep pointing at who made them.

## API docs

The HTTP API is described in [`docs/openapi.yaml`](docs/openapi.yaml) (OpenAPI 3.1): endpoints,
roles, request and response bodies, and the error format. Read it as HTML with:

```sh
npx @redocly/cli build-docs docs/openapi.yaml -o /tmp/api-docs.html && open /tmp/api-docs.html
```

or import the file into Postman or Insomnia.

Every error response is `{"code": "...", "message": "..."}`; `message` is in Indonesian and can be
shown to the user as is.

Update the spec whenever a route, request or response changes in `internal/delivery/http`.

## Development

```sh
go test ./...           # backend tests, repositories against a temporary SQLite database
TEST_POSTGRES_DSN=postgres://user:pass@localhost:5432/masjid_test?sslmode=disable go test ./internal/repository/...
                        # also run repository tests against PostgreSQL
go generate ./...       # regenerate mocks in mock/ (mockgen is a go tool)
cd web && npm run check # frontend typecheck, lint, format check and tests
```

Conventions:

- Dependencies point inward: `delivery` → `usecase` → `domain`; `repository` implements the
  interfaces in `internal/usecase/interface.go`.
- Errors that reach the client are `*apperror.AppError`; the HTTP layer maps its code to a status.
- Tests use gomock mocks from `mock/`, a temporary SQLite database and miniredis, never a real
  external service.
