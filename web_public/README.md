# web_public

The public site for Masjid Digital: what the product does, its documentation, and links to the
live demo. Separate from `web/` (the admin panel and TV display) so it can be hosted on its own,
e.g. on a static host in front of the app's domain.

- `/` (`index.html`): the product page. The hero's TV switches between the real screenshots of
  each mode; the admin panel is shown as its phone screens.
- `/docs/` (`docs/index.html`): the user guide, from installing a TV to every admin menu.

Plain HTML pages, Tailwind CSS v4 and a small script (`src/main.ts`), built with Vite. No React:
the content is in the HTML, so search engines read it without running JavaScript.

## Commands

```sh
cp .env.example .env   # demo links point at the local admin app, http://localhost:5173
npm install
npm run dev            # http://localhost:5174
npm run check          # typecheck + Biome + tests
npm run build          # static files in dist/
```

## Demo links

Every "Coba demo" and "Lihat layar TV demo" button reads its URL from `.env`:

| Variable           | What it opens        | When empty                  |
| ------------------ | -------------------- | --------------------------- |
| `VITE_DEMO_URL`    | the demo admin panel | the demo buttons are hidden |
| `VITE_DEMO_TV_URL` | the demo TV display  | `VITE_DEMO_URL` + `/tv`     |

Locally, run the admin app too (`cd web && npm run dev`, port 5173) so the demo works; both ports
are fixed (`strictPort`), so neither shifts onto the other. The values are baked in at build
time, so rebuild after changing them. Without JavaScript the
buttons open the documentation instead.

The "Coba demo" section runs the demo inside the page: the admin panel in a 390 px phone frame
(on a computer; on a phone it shows a button to open the panel instead) and the TV in a 16:9
frame. Each frame only loads once its option is picked and scrolled into view. For this to work
the demo host must allow being framed by this site: send no `X-Frame-Options`, or a
`Content-Security-Policy: frame-ancestors` that lists the public site's domain.

Every visitor who opens the TV demo registers as a new screen on the demo server (BARU, at most
50). Reset the demo's data now and then, e.g. nightly.

## Screenshots

The images are the ones the repo README shows, in `../docs/screenshots`. They are not copied
here: `vite.config.ts` serves them at `/screenshots/` in dev and writes them to
`dist/screenshots/` in the build. Retake a screenshot there and both the README and the site
update.

## Copy

All visible text is Bahasa Indonesia, like the app. Describe only what the app does today; a
feature that is not built yet is labelled "segera hadir". Keep the docs in step with
`docs/product/prd.md` and the root README when a feature changes.
