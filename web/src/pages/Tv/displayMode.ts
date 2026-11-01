import type { IqamahSettings, PrayerKey, ScheduleItem } from '@/types/tv'
import { parseTime, type WallClock } from './clock'

/*
 * Which screen the TV shows (E8-06). Per waktu: normal → adzan → iqamah → sholat → normal.
 * On Friday, Dzuhur is Jumat instead: jumat (countdown) → adzan → jumat (khutbah) → sholat.
 */

export type Prayer = { key: PrayerKey; name: string; time: string }

export type DisplayState =
  | { mode: 'normal'; next: Prayer; secondsLeft: number; jumat: boolean }
  | { mode: 'adzan'; prayer: Prayer; jumat: boolean }
  | { mode: 'iqamah'; prayer: Prayer; secondsLeft: number; totalMinutes: number }
  | { mode: 'sholat'; prayer: Prayer; jumat: boolean }
  | { mode: 'jumat'; prayer: Prayer; phase: 'before' | 'khutbah'; secondsLeft: number }

const DAY = 24 * 3600
const FRIDAY = 5

const isPrayer = (item: ScheduleItem): item is Prayer =>
  item.key !== 'imsak' && item.key !== 'terbit'

export function displayMode(
  clock: WallClock,
  schedule: readonly ScheduleItem[],
  settings: IqamahSettings,
): DisplayState {
  const prayers = schedule.filter(isPrayer)
  const now = clock.seconds
  const min = 60

  for (const prayer of prayers) {
    const start = parseTime(prayer.time)
    const jumat = clock.weekday === FRIDAY && prayer.key === 'dzuhur'
    const adzanEnd = start + settings.adzan_minutes * min

    if (jumat) {
      const khutbahEnd = adzanEnd + settings.khutbah_minutes * min
      if (now >= start - settings.jumat_before_minutes * min && now < start) {
        return { mode: 'jumat', prayer, phase: 'before', secondsLeft: start - now }
      }
      if (now >= start && now < adzanEnd) return { mode: 'adzan', prayer, jumat }
      if (now >= adzanEnd && now < khutbahEnd) {
        return { mode: 'jumat', prayer, phase: 'khutbah', secondsLeft: 0 }
      }
      if (now >= khutbahEnd && now < khutbahEnd + settings.prayer_minutes * min) {
        return { mode: 'sholat', prayer, jumat }
      }
      continue
    }

    const iqamahMinutes = settings.iqamah_minutes[prayer.key]
    const iqamahAt = start + iqamahMinutes * min
    if (now >= start && now < adzanEnd) return { mode: 'adzan', prayer, jumat }
    if (now >= adzanEnd && now < iqamahAt) {
      return { mode: 'iqamah', prayer, secondsLeft: iqamahAt - now, totalMinutes: iqamahMinutes }
    }
    if (now >= Math.max(adzanEnd, iqamahAt) && now < iqamahAt + settings.prayer_minutes * min) {
      return { mode: 'sholat', prayer, jumat }
    }
  }

  // Normal: count down to the next adzan; after Isya that is tomorrow's Subuh.
  const upcoming = prayers.find((p) => parseTime(p.time) > now)
  const next = upcoming ?? prayers[0]
  if (!next) throw new Error('displayMode needs at least one prayer time')
  const target = parseTime(next.time) + (upcoming ? 0 : DAY)
  const weekday = upcoming ? clock.weekday : (clock.weekday + 1) % 7
  return {
    mode: 'normal',
    next,
    secondsLeft: target - now,
    jumat: weekday === FRIDAY && next.key === 'dzuhur',
  }
}
