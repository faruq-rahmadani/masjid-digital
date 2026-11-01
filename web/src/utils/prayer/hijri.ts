import type { PrayerSettings } from '@/types/prayer'
import { addDays } from './calendar'

const hijriFormat = new Intl.DateTimeFormat('id-ID-u-ca-islamic-umalqura', {
  day: 'numeric',
  month: 'long',
  year: 'numeric',
  timeZone: 'UTC',
})

/**
 * "2026-11-01" → "21 Jumadilawal 1448 H" (Umm al-Qura), shifted by the mosque's Hijri offset
 * once it applies, e.g. when the sidang isbat starts Ramadan a day later than computed.
 */
export function hijriDate(
  date: string,
  settings: Pick<PrayerSettings, 'hijri_offset_days' | 'hijri_offset_from'>,
) {
  const applies = !settings.hijri_offset_from || date >= settings.hijri_offset_from
  const shifted = applies ? addDays(date, settings.hijri_offset_days) : date
  return hijriFormat.format(new Date(`${shifted}T00:00:00Z`))
}
