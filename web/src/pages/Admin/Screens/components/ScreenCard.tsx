import { Badge, Button, Callout, Card, Select } from '@/components'
import type { Screen, ScreenMode } from '@/types/screen'
import { formatTime } from '@/utils/format/date'
import { currentMode, MODE_OPTIONS, seenLine } from '../utils'

type ScreenCardProps = {
  screen: Screen
  now: Date
  onModeChange: (mode: ScreenMode) => void
  onReload: () => void
  onRename: () => void
  onBlock: () => void
}

// A named screen. Online: its forced mode and actions. Offline: why it may be stale.
export function ScreenCard({
  screen,
  now,
  onModeChange,
  onReload,
  onRename,
  onBlock,
}: ScreenCardProps) {
  const online = screen.status === 'online'
  const mode = currentMode(screen, now)
  const selectId = `screen-mode-${screen.id}`

  return (
    <Card padding="lg" className="flex flex-col gap-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="text-md truncate font-bold">{screen.name}</h3>
          <p className="text-ink-muted mt-0.5 text-xs">{seenLine(screen, now)}</p>
        </div>
        <Badge tone={online ? 'success' : 'danger'} size="md">
          {online ? 'Online' : 'Offline'}
        </Badge>
      </div>

      {online && (
        <div className="flex flex-col gap-1.5">
          <label htmlFor={selectId} className="text-ink-soft text-xs font-semibold">
            Paksa mode tampilan
          </label>
          <Select id={selectId} options={MODE_OPTIONS} value={mode} onChange={onModeChange} />
          {mode !== 'auto' && screen.forced_until && (
            <p className="text-gold-ink text-xs">
              Sampai {formatTime(new Date(screen.forced_until))}, lalu kembali ikut jadwal.
            </p>
          )}
        </div>
      )}

      {!online && (
        <Callout tone="danger">
          Layar tetap menampilkan jadwal dari cache, tapi konten baru belum masuk. Cek internet atau
          nyalakan ulang TV.
        </Callout>
      )}

      <div className="grid auto-cols-fr grid-flow-col gap-2">
        {online && (
          <Button size="md" variant="secondary" onClick={onReload}>
            Reload
          </Button>
        )}
        <Button size="md" variant="secondary" onClick={onRename}>
          Ganti nama
        </Button>
        <Button size="md" variant="danger" onClick={onBlock}>
          Blokir
        </Button>
      </div>
    </Card>
  )
}
