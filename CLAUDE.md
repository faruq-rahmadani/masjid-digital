# masjid-digital

Masjid platform MVP: a mobile-first admin panel where mosque pengurus record kas and manage a TV
display, plus the TV display itself (`/tv`, 1920×1080, offline-first). Single-tenant: one
installation serves one mosque.

## Read first

The knowledge base is in [docs/](./docs/README.md). Treat it as the source of truth:

- [docs/product/prd.md](./docs/product/prd.md): scope, roles and permissions, settled decisions,
  per-module rules, data model.
- [docs/product/backlog.md](./docs/product/backlog.md): task IDs (`E7-06`) to cite in commits and PRs.
- [docs/design/design-system.md](./docs/design/design-system.md): tokens, components, UI rules.
- [docs/design/screens.md](./docs/design/screens.md) and `docs/design/mockups/*.html`: the
  mockup is the spec; build to its layout and copy.

If code and docs disagree, ask which is right, then update the other in the same change.

## Layout

- `cmd/`, `internal/`, `pkg/`: Go API (Echo). Layering is described in `internal/*/README.md`.
- `web/`: Vite + React + TypeScript client. See [web/README.md](./web/README.md).
- `web_public/`: the public product site and user docs (static HTML + Tailwind). See
  [web_public/README.md](./web_public/README.md). Its copy describes only what the app does today.

## Non-negotiables

- **Only UI copy is Bahasa Indonesia** (labels, messages, toasts). Everything else is English:
  file and folder names, components, functions, variables, URL paths (`ROUTES` in
  `web/src/route.ts`), DOM ids, comments, tests and docs. Screens are named in English in code
  (Home, not Beranda) even when their UI title is Bahasa.
- **Money is integer Rupiah**; balances are computed, never stored. Soft delete anything involving money.
- **Single-tenant: no `mosque_id`.** One installation serves one mosque.
- **Repositories only read and write.** Calculations (balances, totals) and rules live in the
  usecase; queries filter, order and page, never sum.
- **Menus follow the role** (super admin / takmir / bendahara): hidden, not disabled.
- **UI uses design-system tokens and `@/components`**; no raw hex values or ad-hoc styles.
- **Prayer times are computed on the device** (Kemenag: Subuh 20°, Isya 18°); no external time API.

## Commands

```sh
# web/
npm run dev      # http://localhost:5173 (/kit is the live design system)
npm run check    # typecheck + Biome + tests

# repo root
go run ./cmd/http
```
