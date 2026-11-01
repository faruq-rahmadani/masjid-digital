import type { MosqueTimezone } from '@/types/mosque'

// The mosque's time zone (one of Indonesia's three), as shown in the admin and on the TV.
export const TIMEZONES: readonly { value: MosqueTimezone; label: string; offset: string }[] = [
  { value: 'Asia/Jakarta', label: 'WIB', offset: 'UTC+7' },
  { value: 'Asia/Makassar', label: 'WITA', offset: 'UTC+8' },
  { value: 'Asia/Jayapura', label: 'WIT', offset: 'UTC+9' },
]

/** "Asia/Makassar" → "WITA" */
export function timezoneLabel(timezone: MosqueTimezone) {
  return TIMEZONES.find((tz) => tz.value === timezone)?.label ?? timezone
}
