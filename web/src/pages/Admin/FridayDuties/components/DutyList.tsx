import { Pencil, X } from 'lucide-react'
import { Badge, IconButton } from '@/components'
import type { FridayDuty } from '@/types/fridayDuty'
import { formatLongDay } from '@/utils/format/day'

type DutyListProps = {
  /** The Fridays to list, "YYYY-MM-DD". */
  fridays: string[]
  duties: FridayDuty[]
  onEdit: (date: string) => void
  onDelete: (duty: FridayDuty) => void
}

export function DutyList({ fridays, duties, onEdit, onDelete }: DutyListProps) {
  return (
    <ul className="flex flex-col gap-2">
      {fridays.map((date) => {
        const duty = duties.find((d) => d.date === date)
        const day = formatLongDay(date)
        return (
          <li
            key={date}
            className="bg-card flex items-center justify-between gap-2 rounded-xl py-2.5 pr-1 pl-4"
          >
            <div className="min-w-0 text-xs">
              <p className="text-sm font-bold">{day}</p>
              <DutySummary duty={duty} />
            </div>
            <div className="flex shrink-0">
              <IconButton
                variant="ghost"
                label={`${duty ? 'Ubah' : 'Isi'} petugas ${day}`}
                icon={<Pencil />}
                onClick={() => onEdit(date)}
              />
              {duty && (
                <IconButton
                  variant="ghost"
                  label={`Hapus petugas ${day}`}
                  icon={<X className="text-expense" />}
                  onClick={() => onDelete(duty)}
                />
              )}
            </div>
          </li>
        )
      })}
    </ul>
  )
}

function DutySummary({ duty }: { duty: FridayDuty | undefined }) {
  if (!duty) {
    return (
      <Badge tone="notice" className="mt-1">
        Belum diisi
      </Badge>
    )
  }
  const names = [
    duty.khatib && `Khatib ${duty.khatib}`,
    duty.imam && `Imam ${duty.imam}`,
    duty.muadzin && `Muadzin ${duty.muadzin}`,
  ].filter(Boolean)
  return (
    <>
      <p className="text-ink-muted mt-0.5 truncate">{names.join(' · ')}</p>
      {duty.theme && <p className="text-ink-muted mt-0.5 truncate italic">"{duty.theme}"</p>}
    </>
  )
}
