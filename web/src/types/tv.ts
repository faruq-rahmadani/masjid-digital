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

/** Iqamah & alarm settings (E5-04), in minutes. Iqamah is counted from the adzan time. */
export type IqamahSettings = {
  iqamah_minutes: Record<PrayerKey, number>
  /** How long the "Telah masuk waktu" screen stays up after the adzan time. */
  adzan_minutes: number
  /** How long the screen stays dim after iqamah, while the prayer is on. */
  prayer_minutes: number
  /** Friday: the Jumat screen shows this long before Dzuhur, and the khutbah lasts this long. */
  jumat_before_minutes: number
  khutbah_minutes: number
}

/** Petugas (E5-05): today's imam and muadzin, and on Friday the khatib and theme. */
export type DutyRoster = {
  imam: string
  muadzin: string
  khatib: string
  khutbah_theme: string
}
