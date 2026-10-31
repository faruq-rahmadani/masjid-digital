import type { RouteObject } from 'react-router'
import { RouteError } from '@/pages/RouteError'
import { AppShell } from './layouts/AppShell'

// Each page is lazy-loaded so the initial bundle only carries the shell.
export const routes: RouteObject[] = [
  {
    element: <AppShell />,
    errorElement: <RouteError />,
    children: [
      {
        index: true,
        lazy: async () => ({ Component: (await import('@/pages/HomePage')).HomePage }),
      },
      // Component gallery for building screens against the design system; not shipped.
      ...(import.meta.env.DEV
        ? [
            {
              path: 'kit',
              lazy: async () => ({ Component: (await import('@/pages/kit/KitPage')).KitPage }),
            },
          ]
        : []),
      {
        path: '*',
        lazy: async () => ({ Component: (await import('@/pages/NotFoundPage')).NotFoundPage }),
      },
    ],
  },
]
