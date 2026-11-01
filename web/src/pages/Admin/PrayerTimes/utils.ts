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
