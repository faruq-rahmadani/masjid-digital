import { Pencil } from 'lucide-react'
import { Button, Card } from '@/components'
import type { FridayDuty } from '@/types/fridayDuty'
import { formatLongDay } from '@/utils/format/day'

type NextFridayCardProps = {
  date: string
  duty: FridayDuty | undefined
  onEdit: () => void
}

// Brand like the other screens' "now" card: the coming Friday is the one that matters this week.
export function NextFridayCard({ date, duty, onEdit }: NextFridayCardProps) {
  const officers = [
    { label: 'Khatib', name: duty?.khatib },
    { label: 'Imam', name: duty?.imam },
    { label: 'Muadzin', name: duty?.muadzin },
  ]

  return (
    <Card tone="brand" className="flex flex-col gap-3">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-on-brand-muted text-xs font-semibold">Jumat terdekat</p>
          <p className="text-lg font-extrabold">{formatLongDay(date)}</p>
        </div>
        {/* gold asks to act on what is missing (design-system.md#button); filled is just editable */}
        <Button size="sm" variant={duty ? 'secondary' : 'gold'} onClick={onEdit}>
          <Pencil aria-hidden />
          {duty ? 'Ubah' : 'Isi petugas'}
        </Button>
      </div>
      <dl className="grid grid-cols-3 gap-2">
        {officers.map((officer) => (
          <div key={officer.label} className="bg-brand-2 min-w-0 rounded-md px-3 py-2">
            <dt className="text-on-brand-muted text-2xs font-semibold">{officer.label}</dt>
            <dd className="m-0 truncate text-sm font-bold">{officer.name || '—'}</dd>
          </div>
        ))}
      </dl>
      {duty?.theme && <p className="text-on-brand-muted text-sm italic">"{duty.theme}"</p>}
    </Card>
  )
}
