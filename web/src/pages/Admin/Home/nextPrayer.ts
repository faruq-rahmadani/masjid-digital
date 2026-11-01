import type { MosqueTimezone } from '@/types/mosque'
import { minutesOf, secondsOfDay } from '@/utils/prayer/calendar'

type PrayerTime = { name: string; time: string }

type NextPrayer = { name: string; time: string; minutesLeft: number }

/**
 * The next prayer after `now` on the mosque's clock; after Isya it is tomorrow's Subuh. Times are
 * "HH:MM" in `timezone`, so a phone set to another zone still counts down right.
 */
export function nextPrayer(
  now: Date,
  times: readonly PrayerTime[],
  timezone: MosqueTimezone,
): NextPrayer {
  const minutesNow = secondsOfDay(now, timezone) / 60
  const upcoming = times.find((t) => minutesOf(t.time) > minutesNow)
  const next = upcoming ?? times[0]
  if (!next) throw new Error('nextPrayer needs at least one time')
  const target = minutesOf(next.time) + (upcoming ? 0 : 24 * 60)
  return { name: next.name, time: next.time, minutesLeft: Math.ceil(target - minutesNow) }
}

/** 47 → "00:47", 125 → "02:05" (hours:minutes, as on the mockup). */
export function formatCountdown(minutes: number) {
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`
}
