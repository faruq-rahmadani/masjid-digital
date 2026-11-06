/// <reference types="vitest/config" />

import { createReadStream, readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, type Plugin } from 'vite'

const local = (path: string) => fileURLToPath(new URL(path, import.meta.url))

/** The README's screenshots, the single source for both: served and published at /screenshots/. */
const SCREENSHOTS = local('../docs/screenshots')

function screenshots(): Plugin {
  return {
    name: 'repo-screenshots',
    configureServer(server) {
      server.middlewares.use('/screenshots', (req, res, next) => {
        const name = decodeURIComponent((req.url ?? '').split('?')[0] ?? '').replace(/^\/+/, '')
        if (!readdirSync(SCREENSHOTS).includes(name)) return next()
        res.setHeader('Content-Type', 'image/png')
        createReadStream(join(SCREENSHOTS, name)).pipe(res)
      })
    },
    generateBundle() {
      for (const name of readdirSync(SCREENSHOTS)) {
        this.emitFile({
          type: 'asset',
          fileName: `screenshots/${name}`,
          source: readFileSync(join(SCREENSHOTS, name)),
        })
      }
    },
  }
}

// A static multi-page site: each page is plain HTML (readable by search engines without
// JavaScript) with a small shared script.
export default defineConfig({
  plugins: [tailwindcss(), screenshots()],
  build: {
    rollupOptions: {
      input: {
        home: local('./index.html'),
        docs: local('./docs/index.html'),
      },
    },
  },
  // fixed, next to the admin app's 5173 that the local demo links point at
  server: { port: 5174, strictPort: true },
  test: {
    environment: 'jsdom',
  },
})
