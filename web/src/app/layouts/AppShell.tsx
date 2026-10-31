import { Outlet, ScrollRestoration } from 'react-router'

/*
 * Admin shell. The mockup is a single 390px column: Beranda is the hub and every other screen
 * has a back button, so there is no tab bar. Wider screens get the same column, centered.
 */
export function AppShell() {
  return (
    <div className="mx-auto flex min-h-dvh max-w-app flex-col bg-canvas px-safe md:border-x md:border-line-soft">
      <Outlet />
      <ScrollRestoration />
    </div>
  )
}
