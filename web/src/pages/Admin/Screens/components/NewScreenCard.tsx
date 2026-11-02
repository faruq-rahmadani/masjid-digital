import { Badge, Button, Card } from '@/components'
import type { Screen } from '@/types/screen'
import { formatTimeAgo } from '@/utils/format/date'
import { deviceLabel } from '../utils'

type NewScreenCardProps = {
  screen: Screen
  now: Date
  onName: () => void
  onBlock: () => void
}

// A device that opened /tv but nobody named yet (BARU): name it, or block it if it isn't ours.
export function NewScreenCard({ screen, now, onName, onBlock }: NewScreenCardProps) {
  const title = screen.name_hint ? `Layar baru · ${screen.name_hint}` : 'Layar baru'

  return (
    <Card padding="lg" className="border-gold flex flex-col gap-2.5 border-2 border-dashed">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="text-md truncate font-bold">{title}</h3>
          <p className="text-ink-muted mt-0.5 text-xs">
            Pertama kali dibuka {formatTimeAgo(screen.first_seen_at, now)} ·{' '}
            {deviceLabel(screen.user_agent)}
          </p>
        </div>
        <Badge tone="gold" caps>
          BARU
        </Badge>
      </div>
      <div className="grid grid-cols-2 gap-2">
        <Button size="md" onClick={onName}>
          Beri nama
        </Button>
        <Button size="md" variant="secondary" onClick={onBlock}>
          Bukan layar kita
        </Button>
      </div>
    </Card>
  )
}
