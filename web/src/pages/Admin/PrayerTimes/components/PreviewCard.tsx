import { Card, SectionHeader, Table, TBody, Td, THead, Th, Tr } from '@/components'
import type { PrayerTimeEntry } from '@/types/prayer'
import { cn } from '@/utils/cn'
import { formatShortDay } from '@/utils/format/day'
import { PRAYER_KEYS } from '@/utils/prayer/calculate'

const heads = { subuh: 'Sbh', dzuhur: 'Dzh', ashar: 'Asr', maghrib: 'Mgr', isya: 'Isy' } as const

type PreviewCardProps = {
  /** The mosque's today, highlighted. */
  today: string
  days: { date: string; schedule: PrayerTimeEntry[] }[]
}

// The screen's answer first: the coming week, today on top. It follows the unsaved settings below,
// so a correction or weekly rule is checked before saving.
export function PreviewCard({ today, days }: PreviewCardProps) {
  return (
    <Card>
      <SectionHeader as="h2" title="Pratinjau 7 hari" />
      <Table density="compact" caption="Pratinjau waktu sholat 7 hari" className="mt-2">
        <THead>
          <Tr>
            <Th>Tgl</Th>
            {PRAYER_KEYS.map((key) => (
              <Th key={key}>{heads[key]}</Th>
            ))}
          </Tr>
        </THead>
        <TBody>
          {days.map(({ date, schedule }) => (
            <Tr
              key={date}
              aria-current={date === today ? 'date' : undefined}
              // brand like the context card: the one row to find at a glance
              className={cn(
                date === today && 'bg-brand text-on-brand font-bold [&>*:last-child]:pr-2',
              )}
            >
              <Th scope="row" className={cn(date === today && 'text-on-brand pl-2')}>
                {formatShortDay(date)}
                {date === today && (
                  <span className="text-2xs text-on-brand-muted block font-semibold">Hari ini</span>
                )}
              </Th>
              {PRAYER_KEYS.map((key) => {
                const entry = schedule.find((e) => e.key === key)
                const replaced = entry && entry.source !== 'calculated'
                return (
                  <Td
                    key={key}
                    // gold-ink reads on white, gold on the brand row
                    className={cn(
                      replaced && 'font-bold',
                      replaced && (date === today ? 'text-gold' : 'text-gold-ink'),
                    )}
                  >
                    <span>{entry?.time}</span>
                    {replaced && (
                      <span className="sr-only"> (diganti, hitungan {entry.calculated})</span>
                    )}
                  </Td>
                )
              })}
            </Tr>
          ))}
        </TBody>
      </Table>
      <p className="text-ink-muted mt-2 flex items-center gap-1.5 text-xs">
        <span aria-hidden className="bg-gold inline-block size-1.5 rounded-full" />
        Emas: diganti jadwal mingguan atau jadwal khusus
      </p>
    </Card>
  )
}
