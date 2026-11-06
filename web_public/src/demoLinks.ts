export type DemoEnv = { VITE_DEMO_URL?: string; VITE_DEMO_TV_URL?: string }

/** Where each demo button goes; null hides that button. */
export type DemoUrls = { panel: string | null; tv: string | null }

/** The panel demo is VITE_DEMO_URL; its TV is VITE_DEMO_TV_URL, or the panel demo's /tv. */
export function demoUrls(env: DemoEnv): DemoUrls {
  const panel = env.VITE_DEMO_URL?.trim() || null
  const tv = env.VITE_DEMO_TV_URL?.trim() || (panel && `${panel.replace(/\/+$/, '')}/tv`)
  return { panel, tv: tv || null }
}

/**
 * Points every `[data-demo="panel" | "tv"]` link and `iframe[data-demo-frame]` at its demo, and
 * hides what has no URL: a site deployed without a demo shows no dead buttons, and every
 * `[data-demo-only]` element (the demo section and the buttons leading to it) disappears.
 */
export function applyDemoLinks(root: ParentNode, urls: DemoUrls) {
  for (const link of root.querySelectorAll<HTMLAnchorElement>('a[data-demo]')) {
    const url = urls[link.dataset.demo as keyof DemoUrls]
    link.hidden = !url
    if (url) link.href = url
  }
  // tabs.ts sets src from data-src once the frame's panel shows, so the demo loads on demand
  for (const frame of root.querySelectorAll<HTMLIFrameElement>('iframe[data-demo-frame]')) {
    const url = urls[frame.dataset.demoFrame as keyof DemoUrls]
    if (url) frame.dataset.src = url
  }
  for (const element of root.querySelectorAll<HTMLElement>('[data-demo-only]')) {
    element.hidden = !urls.panel
  }
}
