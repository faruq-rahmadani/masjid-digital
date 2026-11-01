import { Plus, X } from 'lucide-react'
import { Button, Card, IconButton, SectionHeader } from '@/components'
import type { WeeklyPrayerOverride } from '@/types/prayer'
import { SCHEDULE_NAMES } from '@/utils/prayer/calculate'
import { WEEKDAY_LABELS } from '../utils'

type WeeklyRulesCardProps = {
  rules: WeeklyPrayerOverride[]
  onAdd: () => void
  onRemove: (index: number) => void
}

// Rules that repeat every week, e.g. Jumat Dzuhur 12:00. Saved with the rest of the settings.
export function WeeklyRulesCard({ rules, onAdd, onRemove }: WeeklyRulesCardProps) {
  return (
    <Card className="flex flex-col gap-2.5">
      <SectionHeader
        as="h3"
        title="Jadwal mingguan"
        description="Jam yang sama setiap minggu, disimpan dengan tombol Simpan"
        action={
          <Button size="sm" variant="secondary" onClick={onAdd}>
            <Plus aria-hidden />
            Tambah
          </Button>
        }
      />
      {rules.length === 0 ? (
        <p className="text-ink-muted text-sm">Belum ada. Contoh: Dzuhur setiap Jumat jam 12:00.</p>
      ) : (
        <ul className="flex flex-col gap-2">
          {rules.map((rule, index) => (
            <li
              key={`${rule.weekday}-${rule.prayer}`}
              className="bg-canvas flex items-center justify-between gap-2 rounded-md py-2 pr-1 pl-3"
            >
              <div className="min-w-0 text-xs">
                <p className="text-sm font-bold">
                  Setiap {WEEKDAY_LABELS[rule.weekday]}: {SCHEDULE_NAMES[rule.prayer]} {rule.time}
                </p>
                {rule.remarks && <p className="text-ink-muted mt-0.5 truncate">{rule.remarks}</p>}
              </div>
              <IconButton
                variant="ghost"
                label={`Hapus jadwal ${SCHEDULE_NAMES[rule.prayer]} setiap ${WEEKDAY_LABELS[rule.weekday]}`}
                icon={<X className="text-expense" />}
                onClick={() => onRemove(index)}
              />
            </li>
          ))}
        </ul>
      )}
    </Card>
  )
}
