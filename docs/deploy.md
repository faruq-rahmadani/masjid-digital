# Deploy

Three parts, each on its own host:

| Part | What it is | Where |
| --- | --- | --- |
| API | Go server (`cmd/http`), long-running: the TV keeps a server-sent events stream open | A container host: Render, Fly.io or Railway (`Dockerfile`) |
| Admin panel + TV | `web/`, static SPA | Vercel (`web/vercel.json`) |
| Public site + docs | `web_public/`, static pages | Vercel (`web_public/vercel.json`) |
| Database | Postgres | Supabase |

The API does not run as a Vercel Function: TV updates are pushed from memory of one running
process over a stream that a function would cut.

## 1. Database (Supabase)

Use the **session pooler** connection string (port 5432), not the transaction pooler (6543):
the API keeps its connections open and uses prepared statements.

```
postgres://postgres.<project-ref>:<password>@aws-<n>-<region>.pooler.supabase.com:5432/postgres?sslmode=require
```

The API applies pending migrations when it starts (`auto_migrate`, on by default).

## 2. API (container)

Configuration comes from environment variables: every key of `config.yaml.example` is its
path in capitals with `_`, e.g. `app.database.dsn` is `APP_DATABASE_DSN`. `PORT`, which the
host sets, is used as the listen port.

| Variable | Value |
| --- | --- |
| `APP_DATABASE_DRIVER` | `postgres` |
| `APP_DATABASE_DSN` | the Supabase connection string above |
| `APP_AUTH_JWT_SECRET` | at least 32 bytes: `openssl rand -base64 32` |
| `APP_HTTP_SERVER_ALLOWED_ORIGINS` | the admin panel's origin(s), comma separated, e.g. `https://masjid-admin.vercel.app` |
| `APP_CACHE_DRIVER`, `APP_CACHE_DSN` | optional: `redis` and a Redis URL, so logins survive a restart |

Run **one instance**. TV events and the memory cache live in that process.

Pick a region near the database, and a plan that does not sleep: a sleeping service drops
the TV's live connection.

**Render:** New > Blueprint, pick the repo; `render.yaml` creates the service, then fill in
the secret variables. Health check: `/health`.

**Fly.io:** `fly launch --no-deploy` (it finds the `Dockerfile`), `fly secrets set ...`,
`fly scale count 1`, `fly deploy`.

**Railway:** New project from the repo; it builds the `Dockerfile`. Add the variables and
generate a domain.

## 3. Admin panel and TV (Vercel)

New project from the repo, **Root Directory `web`**. Environment variable:

| Variable | Value |
| --- | --- |
| `VITE_API_BASE_URL` | the API's URL with `/api`, e.g. `https://masjid-digital-api.onrender.com/api` |

Then put this project's domain in the API's `APP_HTTP_SERVER_ALLOWED_ORIGINS` and redeploy
the API. The TV is the same site at `/tv`.

## 4. Public site (Vercel)

Another project from the same repo, **Root Directory `web_public`**. Optional variables:
`VITE_DEMO_URL` (the admin panel's URL) and `VITE_DEMO_TV_URL`.

## Check

- `https://<api>/health` answers 200.
- The admin panel logs in, and the network tab shows calls to the API's domain without CORS errors.
- `/tv` on the admin panel's domain pairs, and a change saved in the panel shows on the TV
  within a few seconds.
