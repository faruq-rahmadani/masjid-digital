import type { IqamahSettings } from '@/types/iqamah'
import type { PrayerKey } from '@/types/prayer'
import { clockOf, minutesOf } from '@/utils/prayer/calendar'

export type FlowStep = { label: string; from: string; to: string }

export type PrayerFlow = { name: string; time: string; jumat: boolean; steps: FlowStep[] }

type Prayer = { key: PrayerKey; name: string; time: string }

/**
 * What the TV will show around one prayer, as clock ranges. Mirrors the TV's displayMode: adzan
 * screen → iqamah countdown → mode sholat; on Friday Dzuhur, Jumat screen → adzan → khutbah →
 * mode sholat. A step that ends up empty (adzan screen longer than the iqamah wait) is left out.
 */
export function prayerFlow(prayer: Prayer, settings: IqamahSettings, friday: boolean): PrayerFlow {
  const start = minutesOf(prayer.time)
  const adzanEnd = start + settings.adzan_display_min
  const step = (label: string, from: number, to: number) => ({
    label,
    from: clockOf(from),
    to: clockOf(to),
  })

  if (friday && prayer.key === 'dzuhur') {
    const khutbahEnd = adzanEnd + settings.khutbah_min
    const before =
      settings.jumat_before_min > 0
        ? [step('Layar Jumat', start - settings.jumat_before_min, start)]
        : []
    return {
      name: 'Jumat',
      time: prayer.time,
      jumat: true,
      steps: [
        ...before,
        step('Layar adzan', start, adzanEnd),
        step('Khutbah', adzanEnd, khutbahEnd),
        step('Mode sholat', khutbahEnd, khutbahEnd + settings.sholat_min.jumat),
      ],
    }
  }

  const iqamahAt = start + settings.iqamah_min[prayer.key]
  const countdown = adzanEnd < iqamahAt ? [step('Hitung mundur iqamah', adzanEnd, iqamahAt)] : []
  const sholatStart = Math.max(adzanEnd, iqamahAt)
  return {
    name: prayer.name,
    time: prayer.time,
    jumat: false,
    steps: [
      step('Layar adzan', start, adzanEnd),
      ...countdown,
      step('Mode sholat', sholatStart, iqamahAt + settings.sholat_min[prayer.key]),
    ],
  }
}
