// What the TV display (/tv) shows besides the mosque profile. All mocked for now; see tvService.

import type { PrayerKey } from './prayer'

export type { PrayerKey }

/** One tile of the day's schedule. Imsak and Terbit are shown but have no adzan or iqamah. */
export type ScheduleItem = {
  key: PrayerKey | 'imsak' | 'terbit'
  name: string
  time: string /** "HH:MM" in the mosque's time zone. */
}

/** Laporan kas strip (E7): one period, integer Rupiah. */
export type CashSummary = {
  as_of: string /** "YYYY-MM-DD" */
  opening_balance: number
  income: number
  expense: number
  closing_balance: number
}

/** Slide (E6-04). image_url is empty until uploads exist; the TV shows a placeholder then. */
export type Slide = {
  id: number
  title: string
  caption: string
  image_url: string
}

export type Quote = { text: string; source: string }
