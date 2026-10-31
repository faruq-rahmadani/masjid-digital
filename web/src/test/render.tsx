import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { ToastProvider } from '@/components'
import { routes } from '@/routes'
import { useSessionStore } from '@/stores/sessionStore'
import { sessionFromToken } from '@/utils/auth/token'
import { fakeToken } from './fakeApi'

type Options = {
  /** Log in as this mock user id first (see src/test/fakeApi.ts): 1 super admin, 2 takmir, 3 bendahara. */
  as?: { id: number; role: 'super_admin' | 'takmir' | 'bendahara' }
}

// Render the real route tree at `path` with the app's providers and a fresh query client.
export function renderRoute(path = '/', { as }: Options = {}) {
  if (as) {
    const token = fakeToken(as.id, as.role)
    sessionStorage.setItem('masjid-digital:session', token)
    useSessionStore.setState({ session: sessionFromToken(token) })
  }

  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const router = createMemoryRouter(routes, { initialEntries: [path] })

  return {
    router,
    ...render(
      <QueryClientProvider client={queryClient}>
        <ToastProvider>
          <RouterProvider router={router} />
        </ToastProvider>
      </QueryClientProvider>,
    ),
  }
}

export const superAdmin = { id: 1, role: 'super_admin' } as const
export const takmir = { id: 2, role: 'takmir' } as const
export const bendahara = { id: 3, role: 'bendahara' } as const
