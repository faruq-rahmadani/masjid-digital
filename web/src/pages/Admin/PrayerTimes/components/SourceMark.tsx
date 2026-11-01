import type { PrayerTimeEntry } from '@/types/prayer'

const labels = { weekly: 'jadwal mingguan', date: 'jadwal khusus' } as const

/** A gold dot on a time an override replaced; screen readers also hear the computed time. */
export function SourceMark({ entry }: { entry: PrayerTimeEntry }) {
  if (entry.source === 'calculated') return null
  const label = `${labels[entry.source]}${entry.remarks && `: ${entry.remarks}`}, hitungan ${entry.calculated}`
  return (
    <span className="bg-gold inline-block size-1.5 rounded-full">
      <span className="sr-only">({label})</span>
    </span>
  )
}
