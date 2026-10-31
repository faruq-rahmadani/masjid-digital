# web

Mobile-first React client for masjid-digital: the pengurus admin panel now, the TV display
next. Product and design specs live in [../docs](../docs/README.md); read them before building a
screen.

**Stack:** Vite · React 19 · TypeScript (strict) · React Router (data router, lazy routes) ·
TanStack Query · Tailwind CSS v4 · lucide-react icons · Plus Jakarta Sans (self-hosted) · Vitest + Testing Library · oxlint · Prettier

## Getting started

```sh
nvm use            # Node 24
cp .env.example .env.local
npm install
npm run dev        # http://localhost:5173, /api is proxied to the Go server
```

| Script   | What it does                                 |
| -------- | -------------------------------------------- |
| `dev`    | Vite dev server with HMR                     |
| `build`  | Type-check, then production build to `dist/` |
| `test`   | Vitest in watch mode (`test:run` for CI)     |
| `lint`   | oxlint                                       |
| `format` | Prettier (sorts Tailwind classes too)        |
| `check`  | typecheck + lint + format check + tests      |

## Structure

```
src/
  app/            # composition root: providers, route tree, shell layout, nav
  pages/          # app-level pages not owned by a feature (home, 404, route error)
  features/       # domain areas; see features/README.md
  components/     # the design system; import from '@/components' (see components/README.md)
    actions/  forms/  data-display/  feedback/  overlays/  navigation/  layout/
  lib/            # framework-agnostic helpers: api client, query client, cn(), Rupiah formatting
  config/         # typed env access
  styles/         # Tailwind entry + design tokens
  test/           # test setup and render helpers
```

Dependency direction: `app → pages/features → components → lib/config`. Never the reverse.

## Mobile-first conventions

- Write base classes for the smallest screen; add `sm:`/`md:`/`lg:` only to enhance.
- Use `min-h-dvh`, not `h-screen`: `100vh` is wrong on mobile browsers with dynamic toolbars.
- Respect notches with `pt-safe` / `pb-safe` / `px-safe` (the viewport uses `viewport-fit=cover`).
- Interactive elements are at least `min-h-touch` (44px).
- Form controls stay at least 16px so iOS doesn't zoom on focus.
- Colors, type sizes and radii come from tokens in `src/styles/index.css`; Tailwind's default
  palette is removed. See [design-system.md](../docs/design/design-system.md).
- One 390px column (`max-w-app`), centered on wider screens. Beranda is the hub; other screens
  use `PageHeader` with a back button. No tab bar.
- UI copy is Bahasa Indonesia.

## API

All requests go through `@/lib/api-client`, which prefixes `VITE_API_BASE_URL` (default `/api`)
and throws `ApiError` on non-2xx. Backend routes are expected under `/api`.
