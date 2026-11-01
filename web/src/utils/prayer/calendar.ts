import type { MosqueTimezone } from '@/types/mosque'
import type { Weekday } from '@/types/prayer'

// Dates are "YYYY-MM-DD" in the mosque's time zone and times "HH:MM", as the API stores them.
// Date math runs on UTC midnights so the device's own zone never shifts a day.

const WEEKDAYS: readonly Weekday[] = [
  'sunday',
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday',
]

const toUtc = (date: string) => new Date(`${date}T00:00:00Z`)

/** The mosque's date at `at`, e.g. 2026-11-01T20:00Z in Asia/Jakarta → "2026-11-02". */
export function localDate(at: Date, timezone: MosqueTimezone) {
  // en-CA writes dates as YYYY-MM-DD
  return new Intl.DateTimeFormat('en-CA', { timeZone: timezone }).format(at)
}

/** The mosque's clock at `at`, "HH:MM". */
export function localTime(at: Date, timezone: MosqueTimezone) {
  return new Intl.DateTimeFormat('en-GB', {
    timeZone: timezone,
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  }).format(at)
}

/** Seconds since the mosque's midnight at `at`, e.g. 17:28:30 WITA → 62910. */
export function secondsOfDay(at: Date, timezone: MosqueTimezone) {
  const parts = new Intl.DateTimeFormat('en-GB', {
    timeZone: timezone,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hourCycle: 'h23',
  }).formatToParts(at)
  const part = (type: string) => Number(parts.find((p) => p.type === type)?.value ?? 0)
  return part('hour') * 3600 + part('minute') * 60 + part('second')
}

/** "2026-11-30" + 2 → "2026-12-02" */
export function addDays(date: string, days: number) {
  const d = toUtc(date)
  d.setUTCDate(d.getUTCDate() + days)
  return d.toISOString().slice(0, 10)
}

export function weekdayOf(date: string): Weekday {
  return WEEKDAYS[toUtc(date).getUTCDay()] as Weekday
}

/** "04:05" → 245 */
export function minutesOf(time: string) {
  const [h = 0, m = 0] = time.split(':').map(Number)
  return h * 60 + m
}

/** 245 → "04:05", wrapping around midnight. */
export function clockOf(minutes: number) {
  const m = ((minutes % 1440) + 1440) % 1440
  return `${String(Math.floor(m / 60)).padStart(2, '0')}:${String(m % 60).padStart(2, '0')}`
}
