import { Plus, X } from 'lucide-react'
import { Badge, Button, Callout, Card, IconButton, SectionHeader, Skeleton } from '@/components'
import type { PrayerTimeOverride } from '@/types/prayer'
import { SCHEDULE_NAMES } from '@/utils/prayer/calculate'
import { formatLongDay } from '../utils'

type DateOverridesCardProps = {
  overrides: PrayerTimeOverride[] | undefined
  error: Error | null
  onAdd: () => void
  onDelete: (override: PrayerTimeOverride) => void
}

// One prayer on one date. Unlike the settings, each is saved (and deleted) on its own right away,
// after a confirmation; it sits last in its group so the draft settings stay together.
export function DateOverridesCard({ overrides, error, onAdd, onDelete }: DateOverridesCardProps) {
  return (
    <Card className="flex flex-col gap-2.5">
      <SectionHeader
        as="h3"
        title="Jadwal khusus tanggal tertentu"
        description={
          <span className="flex flex-wrap items-center gap-1.5">
            <Badge tone="notice">Langsung tersimpan</Badge>
            tanpa tombol Simpan · 60 hari ke depan
          </span>
        }
        action={
          <Button size="sm" variant="secondary" onClick={onAdd}>
            <Plus aria-hidden />
            Tambah
          </Button>
        }
      />
      <OverrideList overrides={overrides} error={error} onDelete={onDelete} />
    </Card>
  )
}

type OverrideListProps = Pick<DateOverridesCardProps, 'overrides' | 'error' | 'onDelete'>

function OverrideList({ overrides, error, onDelete }: OverrideListProps) {
  if (error) return <Callout tone="danger">{error.message}</Callout>
  if (!overrides) return <Skeleton className="h-12" />
  if (overrides.length === 0) {
    return <p className="text-ink-muted text-sm">Belum ada jadwal khusus.</p>
  }

  return (
    <ul className="flex flex-col gap-2">
      {overrides.map((o) => (
        <li
          key={o.id}
          className="bg-canvas flex items-center justify-between gap-2 rounded-md py-2 pr-1 pl-3"
        >
          <div className="min-w-0 text-xs">
            <p className="text-sm font-bold">
              {formatLongDay(o.date)}: {SCHEDULE_NAMES[o.prayer]} {o.time}
            </p>
            {o.remarks && <p className="text-ink-muted mt-0.5 truncate">{o.remarks}</p>}
          </div>
          <IconButton
            variant="ghost"
            label={`Hapus jadwal khusus ${SCHEDULE_NAMES[o.prayer]} ${formatLongDay(o.date)}`}
            icon={<X className="text-expense" />}
            onClick={() => onDelete(o)}
          />
        </li>
      ))}
    </ul>
  )
}
