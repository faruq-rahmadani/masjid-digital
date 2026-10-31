# web

Mobile-first React client for masjid-digital: the pengurus admin panel now, the TV display next.
Product and design specs live in [../docs](../docs/README.md); read them before building a screen.

**Stack:** Vite · React 19 · TypeScript (strict) · React Router (lazy routes) · TanStack Query ·
axios · Zustand · react-hook-form + zod · Tailwind CSS v4 · lucide-react · Biome · Vitest +
Testing Library

## Getting started

```sh
nvm use            # Node 24
npm install
npm run dev        # http://localhost:5173
```

Environment variables are documented with sample values in `.env.example`. Defaults live in the
committed `.env`, `.env.development` and `.env.production`; put your own overrides in
`.env.development.local` (git-ignored), e.g. `VITE_API_PROXY_TARGET=http://localhost:8000` when
your Go API runs on 8000. All `VITE_*` values reach the browser, so never put secrets in them.

| Script   | What it does                                      |
| -------- | ------------------------------------------------- |
| `dev`    | Vite dev server; `/kit` is the live design system |
| `build`  | Type-check, then production build to `dist/`      |
| `test`   | Vitest in watch mode (`test:run` once)            |
| `lint`   | Biome: lint + format + import order check         |
| `format` | Biome: apply safe fixes and formatting            |
| `check`  | typecheck + lint + tests (run before pushing)     |

## Structure

```
web/
├── index.html, vite.config.ts, tsconfig*.json   # Vite + TS
├── biome.json                                   # lint + format
├── .env, .env.development, .env.production      # public VITE_* defaults (*.local = your overrides)
├── public/                                      # static files
└── src/
    ├── main.tsx, App.tsx      # entry; providers + router
    ├── route.ts               # ROUTES: every app path, use instead of string literals
    ├── routes/
    │   ├── index.tsx          # route table (lazy pages, guards)
    │   ├── menu.ts            # Beranda "Kelola" menu: label, icon, path, permission
    │   ├── guards.tsx         # RequireAuth, RequireAccess
    │   └── types.ts
    ├── layouts/               # AppShell (390px column, session lifecycle)
    ├── pages/
    │   ├── LoginPage/  Kit/   # public / dev-only
    │   ├── NotFoundPage, ComingSoonPage, RouteError
    │   └── Admin/             # internal users; access per role via the ACL, not the folder
    │       ├── Home/          # hub every role lands on
    │       └── Users/         # UsersPage + its sheets, row, schema, utils, tests
    ├── components/            # design system, import from '@/components' (see its README)
    ├── hooks/                 # TanStack Query hooks per resource (useUsers) and shared hooks
    ├── api/client.ts          # API client: apiGet / apiPost / apiPut (axios + interceptors)
    ├── services/              # <resource>Service.ts: one exported function per endpoint
    ├── stores/                # Zustand: sessionStore (login session only)
    ├── types/                 # API request/response types, global.d.ts
    ├── utils/
    │   ├── auth/              # permissions (ACL), roles, token storage
    │   ├── format/            # money, phone, text
    │   └── cn.ts, env.ts
    ├── styles/                # Tailwind entry + design tokens
    └── test/                  # Vitest setup and render helpers
```

### Where things go

- **A new screen:** a folder `pages/Admin/<Name>/` for internal users (public screens such as the
  TV display go beside `Admin/`). Add it to `routes/index.tsx` as a lazy route, its path to
  `route.ts`, and (if it is a menu item) set `ready: true` in `routes/menu.ts`. Inside the folder,
  group what only that page uses, and import it relatively (`./drawers/...`):

  ```
  pages/Admin/Users/
  ├── UsersPage.tsx (+ .test.tsx)   # the screen
  ├── components/                   # pieces of this page (UserRow)
  ├── drawers/                      # bottom sheets built on <Sheet> (AddUserDrawer)
  ├── modals/                       # confirmations built on <ConfirmDialog> (DeactivateUserModal)
  ├── schema.ts                     # zod schemas for its forms
  └── utils.ts (+ .test.ts)         # sorting, filtering, rules
  ```

  Create a subfolder only when it has a file. Anything a second page needs moves to
  `src/components/` (generic) or `src/hooks/` (data).
- **A new API resource:** types in `types/<resource>.ts`, calls in
  `services/<resource>Service.ts` using `apiGet` / `apiPost` / `apiPut` from `@/api/client`
  (paths from `internal/delivery/http/router.go`), query hooks in
  `hooks/use<Resource>.ts`. Mock it inside the function while the endpoint isn't ready (below).
- **State:** server data lives in TanStack Query. Zustand is only for client state that many
  screens share (today: the session). Filters and form state stay local to the page.
- **Permissions (ACL):** `utils/auth/permissions.ts` is the PRD roles table. Every `Admin/` route sits
  behind `RequireAccess module="..."`, and menus show only what `canOpen(role, module)` allows
  (hidden, not disabled). Changing who sees what is a change to that table, not to folders.
- Tests sit next to the code they cover (`*.test.ts(x)`).

## Mobile-first conventions

- Base classes for the smallest screen; `sm:`/`md:` only to enhance.
- `min-h-dvh`, not `h-screen`. Safe areas via `pt-safe-<n>` / `pb-safe-<n>` (spacing n, or the
  notch inset if larger) and `px-safe`. Never pair them with `py-*`/`pt-*`; one class per side.
- Interactive elements at least `min-h-touch` (44px). Form controls at least 16px on iOS.
- Colors, type and radii only from tokens in `src/styles/index.css`
  ([design-system.md](../docs/design/design-system.md)).
- One 390px column (`max-w-app`); Beranda is the hub, other screens have a back button.
- Only UI copy is Bahasa Indonesia; file names, identifiers and URL paths are English.

## API

Every request goes through `src/api/client.ts`:

```ts
import { apiGet, apiPost, apiPut, type DataResponse } from '@/api/client'

const users = (await apiGet<DataResponse<User[]>>('/v1/users')).data
await apiPost<MessageResponse>('/v1/users', body)
```

- Each helper resolves to the response body (no axios wrapper); type it as the API returns it.
- Paths are relative to `VITE_API_BASE_URL` (default `/api`).
- Interceptors send the session token and turn every failure into an `ApiError` with the API's
  `code` and its user-facing `message`; a 401 on a logged-in request ends the session.
- The axios instance is private, so all requests share this behavior. Screens call services or
  query hooks, not the client directly.

## Mock API

Pages never hold data: anything a screen shows comes from a service, even before its endpoint
exists. Until then the service function returns mock data, and swapping in the real call later
doesn't touch the page.

| Service                | Status                                   |
| ---------------------- | ---------------------------------------- |
| `authService`          | Live                                     |
| `userService`          | Live                                     |
| `mosqueProfileService` | Live                                     |
| `prayerTimeService`    | Mock (computed on device per PRD, E4-01) |
| `screenService`        | Mock (`GET /screens`, E9-02)             |
| `activityService`      | Mock (activity log, E2-05)               |

Mock in the shape the API will return (timestamps, ids), not display text; pages format it.
Pattern:

```ts
export async function listSlides(): Promise<Slide[]> {
  // const resp = await apiGet<DataResponse<Slide[]>>('/v1/slides')
  // return resp.data
  return MOCK_SLIDES // in a `--- Mock ---` section at the bottom of the file
}
```

- Biome only warns (and keeps) the unused `apiGet` / `apiPost` imports while a file is mocked;
  TypeScript's own unused checks are off for this reason.
- Tests never use these mocks: `src/test/setup.ts` swaps every service for the fakes in
  `src/test/fakeApi.ts`. Add a fake there for each new service.
- Mocks ship in the bundle: make sure nothing is mocked before a production release.

## Running against the Go API

1. From the repo root: `go run ./cmd/http` (reads `config.yaml`; check `curl localhost:<port>/health`).
2. If the API doesn't listen on 3000, set `VITE_API_PROXY_TARGET=http://localhost:<port>` in
   `web/.env.development.local` and restart `npm run dev`.
3. Log in with an account from that database. Contracts: `docs/openapi.yaml`.
