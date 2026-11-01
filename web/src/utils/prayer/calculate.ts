import { CalculationParameters, Coordinates, Madhab, PrayerTimes, Rounding } from 'adhan'
import type { MosqueTimezone } from '@/types/mosque'
import type {
  PrayerKey,
  PrayerSettings,
  PrayerTimeEntry,
  PrayerTimeOverride,
  ScheduleKey,
} from '@/types/prayer'
import { clockOf, localTime, minutesOf, weekdayOf } from './calendar'

export type PrayerLocation = { latitude: number; longitude: number; timezone: MosqueTimezone }

export const SCHEDULE_NAMES: Record<ScheduleKey, string> = {
  imsak: 'Imsak',
  subuh: 'Subuh',
  terbit: 'Terbit',
  dzuhur: 'Dzuhur',
  ashar: 'Ashar',
  maghrib: 'Maghrib',
  isya: 'Isya',
}

export const PRAYER_KEYS: readonly PrayerKey[] = ['subuh', 'dzuhur', 'ashar', 'maghrib', 'isya']

// Kemenag RI: Subuh at the sun 20° below the horizon, Isya at 18°, Ashar when a shadow equals its
// object (Syafi'i), then 2 minutes of ihtiyat (safety) on every time, Terbit taken 2 minutes
// earlier. Dzuhur gets one more minute, for the sun's disk to clear the meridian. Rounded to the
// safe side: prayers up to the next minute, Terbit down.
const IHTIYAT_MIN = 2
const DZUHUR_EXTRA_MIN = 1
// Kemenag's sunset also counts the horizon dip of the site's elevation, so it falls a little after
// adhan's (sun's center 0.833° down). 30 s is fitted to Bimas Islam's Jakarta tables (see tests):
// published Maghrib sits 4–51 s after adhan's sunset before the ihtiyat.
const SUNSET_DIP_MS = 30_000

function kemenagParameters() {
  const params = new CalculationParameters('Other', 20, 18)
  params.madhab = Madhab.Shafi
  params.rounding = Rounding.None
  return params
}

/** Minutes after the mosque's midnight, rounded to the safe side. */
function toLocalMinutes(at: Date, timezone: MosqueTimezone, round: 'up' | 'down') {
  const shifted = new Date(at.getTime() + (round === 'up' ? 59_999 : 0))
  return minutesOf(localTime(shifted, timezone))
}

/** Kemenag times of `date` ("YYYY-MM-DD", the mosque's), in minutes after local midnight. */
export function computeKemenag(date: string, location: PrayerLocation) {
  const [y = 0, m = 1, d = 1] = date.split('-').map(Number)
  // adhan reads the day from the Date's local fields, whatever the device's zone
  const times = new PrayerTimes(
    new Coordinates(location.latitude, location.longitude),
    new Date(y, m - 1, d),
    kemenagParameters(),
  )
  const tz = location.timezone
  return {
    subuh: toLocalMinutes(times.fajr, tz, 'up') + IHTIYAT_MIN,
    terbit: toLocalMinutes(times.sunrise, tz, 'down') - IHTIYAT_MIN,
    dzuhur: toLocalMinutes(times.dhuhr, tz, 'up') + IHTIYAT_MIN + DZUHUR_EXTRA_MIN,
    ashar: toLocalMinutes(times.asr, tz, 'up') + IHTIYAT_MIN,
    maghrib:
      toLocalMinutes(new Date(times.maghrib.getTime() + SUNSET_DIP_MS), tz, 'up') + IHTIYAT_MIN,
    isya: toLocalMinutes(times.isha, tz, 'up') + IHTIYAT_MIN,
  }
}

/**
 * The day's schedule, Imsak to Isya. Layers from the most general, the most specific winning:
 * computed (Kemenag) + correction → weekly override → date override.
 * `overrides` may hold other dates; only `date`'s apply.
 */
export function prayerSchedule(
  date: string,
  location: PrayerLocation,
  settings: PrayerSettings,
  overrides: readonly PrayerTimeOverride[] = [],
): PrayerTimeEntry[] {
  const computed = computeKemenag(date, location)
  const corrected = (key: keyof typeof computed) => computed[key] + settings.corrections[key]
  const weekday = weekdayOf(date)

  const entry = (key: ScheduleKey, minutes: number): PrayerTimeEntry => ({
    key,
    name: SCHEDULE_NAMES[key],
    time: clockOf(minutes),
    calculated: clockOf(minutes),
    source: 'calculated',
    remarks: '',
  })

  const prayer = (key: PrayerKey): PrayerTimeEntry => {
    const base = entry(key, corrected(key))
    const onDate = overrides.find((o) => o.date === date && o.prayer === key)
    if (onDate) return { ...base, time: onDate.time, source: 'date', remarks: onDate.remarks }
    const weekly = settings.weekly_overrides.find((o) => o.weekday === weekday && o.prayer === key)
    if (weekly) return { ...base, time: weekly.time, source: 'weekly', remarks: weekly.remarks }
    return base
  }

  return [
    // Imsak follows the computed Subuh, not an override of it: it marks the fast, not the adzan.
    entry('imsak', corrected('subuh') - settings.imsak_offset_min),
    prayer('subuh'),
    entry('terbit', corrected('terbit')),
    prayer('dzuhur'),
    prayer('ashar'),
    prayer('maghrib'),
    prayer('isya'),
  ]
}

export const DEFAULT_PRAYER_SETTINGS: PrayerSettings = {
  corrections: { subuh: 0, terbit: 0, dzuhur: 0, ashar: 0, maghrib: 0, isya: 0 },
  imsak_offset_min: 10,
  weekly_overrides: [],
  hijri_offset_days: 0,
  hijri_offset_from: '',
}
