import type { RouteObject } from 'react-router'
import { AppShell } from '@/layouts/AppShell'
import { ComingSoonPage } from '@/pages/ComingSoonPage'
import { RouteError } from '@/pages/RouteError'
import { ROUTES } from '@/route'
import { RequireAccess, RequireAuth } from './guards'
import { MENU } from './menu'

// Route table. Pages are lazy-loaded so each screen is its own chunk.
export const routes: RouteObject[] = [
  {
    element: <AppShell />,
    errorElement: <RouteError />,
    children: [
      {
        path: ROUTES.login,
        lazy: async () => ({ Component: (await import('@/pages/LoginPage/LoginPage')).LoginPage }),
      },
      {
        element: <RequireAuth />,
        children: [
          // Every internal screen sits behind its ACL entry (utils/permissions.ts), Home included.
          {
            element: <RequireAccess module="home" />,
            children: [
              {
                index: true,
                lazy: async () => ({
                  Component: (await import('@/pages/Admin/Home/HomePage')).HomePage,
                }),
              },
            ],
          },
          {
            path: ROUTES.users,
            element: <RequireAccess module="users" />,
            children: [
              {
                index: true,
                lazy: async () => ({
                  Component: (await import('@/pages/Admin/Users/UsersPage')).UsersPage,
                }),
              },
            ],
          },
          // Menu items without a screen yet, each still guarded by its role permission.
          ...MENU.filter((item) => !item.ready).map((item) => ({
            path: item.path,
            element: <RequireAccess module={item.module} />,
            children: [{ index: true, element: <ComingSoonPage title={item.label} /> }],
          })),
        ],
      },
      // Component gallery for building screens against the design system; not shipped.
      ...(import.meta.env.DEV
        ? [
            {
              path: ROUTES.kit,
              lazy: async () => ({ Component: (await import('@/pages/Kit/KitPage')).KitPage }),
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
