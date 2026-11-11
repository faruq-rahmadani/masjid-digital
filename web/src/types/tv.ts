// What the TV display (/tv) shows besides the mosque profile; see tvService.

import type { PrayerKey } from './prayer'

export type { PrayerKey }

/** One tile of the day's schedule. Imsak and Terbit are shown but have no adzan or iqamah. */
export type ScheduleItem = {
  key: PrayerKey | 'imsak' | 'terbit'
  name: string
  time: string /** "HH:MM" in the mosque's time zone. */
}

export type Quote = { text: string; source: string }
