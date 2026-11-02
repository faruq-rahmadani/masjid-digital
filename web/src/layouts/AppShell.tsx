import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { Outlet, ScrollRestoration } from 'react-router'
import { useDocumentBranding } from '@/hooks/useDocumentBranding'
import { useSessionStore } from '@/stores/sessionStore'

/*
 * Admin shell. The mockup is a single 390px column: Home is the hub and every other screen
 * has a back button, so there is no tab bar. Wider screens get the same column, centered.
 */
export function AppShell() {
  useSessionLifecycle()
  useDocumentBranding()

  return (
    <div className="px-safe bg-canvas md:border-line-soft max-w-app mx-auto flex min-h-dvh flex-col md:border-x">
      <Outlet />
      <ScrollRestoration />
    </div>
  )
}

/** Ends the session when its token expires, and drops cached data when anyone logs out. */
function useSessionLifecycle() {
  const queryClient = useQueryClient()
  const expiresAt = useSessionStore((s) => s.session?.expiresAt)

  useEffect(
    () =>
      useSessionStore.subscribe((state, previous) => {
        if (previous.session && !state.session) queryClient.clear()
      }),
    [queryClient],
  )

  useEffect(() => {
    if (expiresAt === undefined) return
    const timer = setTimeout(
      () => useSessionStore.getState().endSession(),
      Math.max(0, expiresAt - Date.now()),
    )
    return () => clearTimeout(timer)
  }, [expiresAt])
}
