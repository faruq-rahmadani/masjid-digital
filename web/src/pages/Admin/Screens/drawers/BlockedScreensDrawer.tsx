import { Button, List, ListRow, Sheet } from '@/components'
import type { Screen } from '@/types/screen'
import { formatTimeAgo } from '@/utils/format/date'
import { deviceLabel } from '../utils'

type BlockedScreensDrawerProps = {
  screens: Screen[]
  now: Date
  onUnblock: (screen: Screen) => void
  onClose: () => void
}

// Blocked screens stay out of the list; this is where a wrong block is undone.
export function BlockedScreensDrawer({
  screens,
  now,
  onUnblock,
  onClose,
}: BlockedScreensDrawerProps) {
  return (
    <Sheet open onClose={onClose} title={`Layar diblokir (${screens.length})`}>
      <p className="text-ink-muted mb-3 text-sm">
        Layar yang diblokir tidak bisa terhubung. Buka blokir jika salah blokir; layar akan
        tersambung lagi dalam 5 menit.
      </p>
      <List>
        {screens.map((screen) => (
          <ListRow
            key={screen.id}
            title={screen.name || 'Layar baru'}
            description={[
              deviceLabel(screen.user_agent),
              screen.blocked_at && `Diblokir ${formatTimeAgo(screen.blocked_at, now)}`,
            ]
              .filter(Boolean)
              .join(' · ')}
            trailing={
              <Button size="sm" variant="secondary" onClick={() => onUnblock(screen)}>
                Buka blokir
              </Button>
            }
          />
        ))}
      </List>
    </Sheet>
  )
}
