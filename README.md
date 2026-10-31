# masjid-digital

Admin panel and backend for running a mosque: users and roles, the mosque profile, and next
prayer times, TV display slides, announcements and kas (treasury) records.

Each deployment serves **one mosque** (single tenant): there is no mosque table or `mosque_id`;
the mosque's profile and logo are stored once.

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
docs/            # API spec (openapi.yaml)
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
