import { useState } from 'react'
import {
  Button,
  Callout,
  Card,
  EmptyState,
  PageBody,
  PageHeader,
  SectionHeader,
  Skeleton,
  useToast,
} from '@/components'
import { useNow } from '@/hooks/useNow'
import {
  useBlockScreen,
  useScreens,
  useSendScreenCommand,
  useUnblockScreen,
  useUpdateScreen,
} from '@/hooks/useScreens'
import { ROUTES } from '@/route'
import type { Screen } from '@/types/screen'
import { InstallCard } from './components/InstallCard'
import { NewScreenCard } from './components/NewScreenCard'
import { ScreenCard } from './components/ScreenCard'
import { BlockedScreensDrawer } from './drawers/BlockedScreensDrawer'
import { NameScreenDrawer } from './drawers/NameScreenDrawer'
import { type ScreenAction, ScreenActionModal } from './modals/ScreenActionModal'
import { groupScreens } from './utils'

// Layar TV (E9-06, admin-layar.html). Super admin and takmir; the route is guarded by
// RequireAccess. Blocked screens are hidden here and listed apart so a wrong block can be undone.
export function ScreensPage() {
  return (
    <>
      <PageHeader title="Layar TV" />
      <ScreensContent />
    </>
  )
}

function ScreensContent() {
  const now = useNow()
  const screens = useScreens()
  const [naming, setNaming] = useState<Screen | null>(null)
  const [showBlocked, setShowBlocked] = useState(false)
  const [action, setAction] = useState<ScreenAction | null>(null)
  const run = useScreenActions()
  const tvUrl = `${window.location.origin}${ROUTES.tv}`

  if (screens.error) {
    return (
      <PageBody>
        <InstallCard url={tvUrl} />
        <Callout
          tone="danger"
          title="Daftar layar gagal dimuat"
          action={
            <Button size="md" variant="secondary" onClick={() => screens.refetch()}>
              Coba lagi
            </Button>
          }
        >
          {screens.error.message}
        </Callout>
      </PageBody>
    )
  }

  if (!screens.data) {
    return (
      <PageBody>
        <InstallCard url={tvUrl} />
        <Card aria-busy aria-label="Memuat daftar layar" className="flex flex-col gap-3">
          <Skeleton className="h-8" />
          <Skeleton className="h-8" />
        </Card>
      </PageBody>
    )
  }

  const { named, fresh, blocked } = groupScreens(screens.data)
  const online = named.some((s) => s.status === 'online')

  return (
    <>
      <PageBody>
        <InstallCard url={tvUrl} />

        <SectionHeader
          title={`Layar terdaftar (${named.length + fresh.length})`}
          action={
            online && (
              <Button size="sm" variant="ghost" onClick={() => setAction({ kind: 'reload-all' })}>
                Muat ulang semua
              </Button>
            )
          }
        />

        {named.length + fresh.length === 0 && (
          <EmptyState
            title="Belum ada layar"
            description="Buka alamat di atas di browser TV, layar langsung muncul di sini."
          />
        )}

        {named.map((screen) => (
          <ScreenCard
            key={screen.id}
            screen={screen}
            now={now}
            onModeChange={(mode) => setAction({ kind: 'mode', screen, mode })}
            onReload={() => setAction({ kind: 'reload', screen })}
            onRename={() => setNaming(screen)}
            onBlock={() => setAction({ kind: 'block', screen })}
          />
        ))}

        {fresh.map((screen) => (
          <NewScreenCard
            key={screen.id}
            screen={screen}
            now={now}
            onName={() => setNaming(screen)}
            onBlock={() => setAction({ kind: 'block', screen })}
          />
        ))}

        {blocked.length > 0 && (
          <Button variant="ghost" size="md" onClick={() => setShowBlocked(true)}>
            Lihat layar diblokir ({blocked.length})
          </Button>
        )}
      </PageBody>

      {naming && <NameScreenDrawer screen={naming} now={now} onClose={() => setNaming(null)} />}
      {showBlocked && (
        <BlockedScreensDrawer
          screens={blocked}
          now={now}
          onUnblock={(screen) => setAction({ kind: 'unblock', screen })}
          onClose={() => setShowBlocked(false)}
        />
      )}
      <ScreenActionModal action={action} onClose={() => setAction(null)} onConfirm={run} />
    </>
  )
}

/** Runs a confirmed action and reports it with a toast. */
function useScreenActions() {
  const update = useUpdateScreen()
  const block = useBlockScreen()
  const unblock = useUnblockScreen()
  const command = useSendScreenCommand()
  const toast = useToast()
  const report = (message: string) => ({
    onSuccess: () => toast({ message }),
    onError: (error: Error) => toast({ tone: 'danger', message: error.message }),
  })

  const handlers: { [K in ScreenAction['kind']]: (a: Extract<ScreenAction, { kind: K }>) => void } =
    {
      mode: ({ screen, mode }) =>
        update.mutate(
          { id: screen.id, body: { name: screen.name, forced_mode: mode } },
          report('Mode layar diubah'),
        ),
      reload: ({ screen }) =>
        command.mutate({ command: 'reload', id: screen.id }, report('Layar dimuat ulang')),
      'reload-all': () => command.mutate({ command: 'reload' }, report('Semua layar dimuat ulang')),
      block: ({ screen }) => block.mutate(screen.id, report('Layar diblokir')),
      unblock: ({ screen }) => unblock.mutate(screen.id, report('Blokir layar dibuka')),
    }

  return (action: ScreenAction) => (handlers[action.kind] as (a: ScreenAction) => void)(action)
}
