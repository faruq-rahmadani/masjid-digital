export type PrayerKey = 'subuh' | 'dzuhur' | 'ashar' | 'maghrib' | 'isya'
export type ScheduleKey = 'imsak' | 'subuh' | 'terbit' | 'dzuhur' | 'ashar' | 'maghrib' | 'isya'

export type Weekday =
  | 'monday'
  | 'tuesday'
  | 'wednesday'
  | 'thursday'
  | 'friday'
  | 'saturday'
  | 'sunday'

export type PrayerCorrections = Record<Exclude<ScheduleKey, 'imsak'>, number>

export type WeeklyPrayerOverride = {
  weekday: Weekday
  prayer: PrayerKey
  time: string /** "HH:MM" */
  remarks: string
}

export type PrayerSettings = {
  corrections: PrayerCorrections
  imsak_offset_min: number
  weekly_overrides: WeeklyPrayerOverride[]
  /** Days added to the computed Hijri date (-2..2), from hijri_offset_from or always when empty. */
  hijri_offset_days: number
  hijri_offset_from: string
}

export type PrayerTimeOverride = {
  id: number
  date: string /** "YYYY-MM-DD" */
  prayer: PrayerKey
  time: string /** "HH:MM" */
  remarks: string
  created_by: number
  created_at: string
  updated_at: string
}

export type CreatePrayerOverrideRequest = Pick<
  PrayerTimeOverride,
  'date' | 'prayer' | 'time' | 'remarks'
>

export type PrayerConfig = {
  settings: PrayerSettings
  overrides: PrayerTimeOverride[]
}

export type PrayerTimeSource = 'calculated' | 'weekly' | 'date'

export type PrayerTimeEntry = {
  key: ScheduleKey
  name: string
  time: string /** "HH:MM" in the mosque's time zone. */
  calculated: string
  source: PrayerTimeSource
  remarks: string
}
