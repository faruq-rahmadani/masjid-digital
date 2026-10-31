import { Navigate, Outlet, useLocation } from 'react-router'
import { EmptyState, PageBody, PageHeader, TextLink } from '@/components'
import { ROUTES } from '@/route'
import { useAuth, useSession } from '@/stores/sessionStore'
import { canOpen, type ModuleKey } from '@/utils/auth/permissions'

/** Layout route: renders its children only with a session, otherwise goes to login and back. */
export function RequireAuth() {
  const { session } = useAuth()
  const location = useLocation()

  if (!session) {
    const next = location.pathname + location.search
    const to =
      next === ROUTES.home ? ROUTES.login : `${ROUTES.login}?next=${encodeURIComponent(next)}`
    return <Navigate to={to} replace />
  }

  return <Outlet />
}

/** Layout route for a module. Menus already hide what a role can't open; this covers typed URLs. */
export function RequireAccess({ module }: { module: ModuleKey }) {
  const { role } = useSession()
  if (canOpen(role, module)) return <Outlet />

  return (
    <>
      <PageHeader title="Tidak ada akses" backTo={ROUTES.home} />
      <PageBody>
        <EmptyState
          title="Menu ini tidak tersedia untuk peran Anda"
          description="Hubungi super admin masjid jika Anda memerlukan akses."
          action={<TextLink to={ROUTES.home}>Kembali ke Beranda</TextLink>}
        />
      </PageBody>
    </>
  )
}
