import type { SelectOption } from '@/components'
import type { PrayerKey, Weekday } from '@/types/prayer'
import { PRAYER_KEYS, SCHEDULE_NAMES } from '@/utils/prayer/calculate'

export const PRAYER_OPTIONS: SelectOption<PrayerKey>[] = PRAYER_KEYS.map((key) => ({
  value: key,
  label: SCHEDULE_NAMES[key],
}))

export const WEEKDAY_LABELS: Record<Weekday, string> = {
  monday: 'Senin',
  tuesday: 'Selasa',
  wednesday: 'Rabu',
  thursday: 'Kamis',
  friday: 'Jumat',
  saturday: 'Sabtu',
  sunday: 'Minggu',
}

export const WEEKDAY_OPTIONS: SelectOption<Weekday>[] = (
  Object.keys(WEEKDAY_LABELS) as Weekday[]
).map((value) => ({ value, label: WEEKDAY_LABELS[value] }))

// Dates are the mosque's "YYYY-MM-DD"; formatting them at UTC midnight keeps the device's zone out.
const atUtc = (date: string) => new Date(`${date}T00:00:00Z`)
const shortDay = new Intl.DateTimeFormat('id-ID', {
  weekday: 'short',
  day: 'numeric',
  timeZone: 'UTC',
})
const longDay = new Intl.DateTimeFormat('id-ID', {
  weekday: 'long',
  day: 'numeric',
  month: 'short',
  timeZone: 'UTC',
})

/** "2026-11-05" → "Kam 5" */
export function formatShortDay(date: string) {
  return shortDay.format(atUtc(date)).replace(',', '')
}

/** "2026-11-06" → "Jumat, 6 Nov" */
export function formatLongDay(date: string) {
  return longDay.format(atUtc(date))
}
