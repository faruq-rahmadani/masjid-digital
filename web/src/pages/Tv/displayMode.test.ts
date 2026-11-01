import { describe, expect, it } from 'vitest'
import type { IqamahSettings, ScheduleItem } from '@/types/tv'
import { formatHijriDate, formatHms, formatLongDate, formatMs, parseTime, wallClock } from './clock'
import { displayMode } from './displayMode'

const SCHEDULE: ScheduleItem[] = [
  { key: 'imsak', name: 'Imsak', time: '04:32' },
  { key: 'subuh', name: 'Subuh', time: '04:42' },
  { key: 'terbit', name: 'Terbit', time: '05:56' },
  { key: 'dzuhur', name: 'Dzuhur', time: '12:09' },
  { key: 'ashar', name: 'Ashar', time: '15:24' },
  { key: 'maghrib', name: 'Maghrib', time: '18:15' },
  { key: 'isya', name: 'Isya', time: '19:25' },
]

const SETTINGS: IqamahSettings = {
  iqamah_minutes: { subuh: 10, dzuhur: 7, ashar: 7, maghrib: 7, isya: 7 },
  adzan_minutes: 1,
  prayer_minutes: 10,
  jumat_before_minutes: 60,
  khutbah_minutes: 30,
}

const WEDNESDAY = 3
const FRIDAY = 5
const at = (time: string, weekday = WEDNESDAY, extraSeconds = 0) =>
  displayMode({ seconds: parseTime(time) + extraSeconds, weekday }, SCHEDULE, SETTINGS)

describe('displayMode', () => {
  it('counts down to the next prayer, skipping Imsak and Terbit', () => {
    expect(at('17:27', WEDNESDAY, 15)).toMatchObject({
      mode: 'normal',
      next: { name: 'Maghrib' },
      secondsLeft: 47 * 60 + 45,
    })
    expect(at('05:30')).toMatchObject({ mode: 'normal', next: { name: 'Dzuhur' } })
  })

  it('goes from adzan to iqamah to sholat and back to normal', () => {
    expect(at('18:15')).toMatchObject({ mode: 'adzan', prayer: { name: 'Maghrib' } })
    expect(at('18:16', WEDNESDAY, 48)).toMatchObject({
      mode: 'iqamah',
      secondsLeft: 5 * 60 + 12,
      totalMinutes: 7,
    })
    expect(at('18:22')).toMatchObject({ mode: 'sholat', prayer: { name: 'Maghrib' } })
    expect(at('18:32')).toMatchObject({ mode: 'normal', next: { name: 'Isya' } })
  })

  it('uses each prayer its own iqamah time', () => {
    expect(at('04:50')).toMatchObject({
      mode: 'iqamah',
      prayer: { name: 'Subuh' },
      totalMinutes: 10,
    })
  })

  it('wraps to tomorrow after Isya, and flags a Friday Dzuhur as Jumat', () => {
    expect(at('21:00')).toMatchObject({
      mode: 'normal',
      next: { name: 'Subuh' },
      secondsLeft: parseTime('04:42') + 24 * 3600 - parseTime('21:00'),
    })
    expect(at('09:00', FRIDAY)).toMatchObject({ mode: 'normal', jumat: true })
  })

  it('runs Jumat instead of Dzuhur on Friday', () => {
    expect(at('11:51', FRIDAY, 48)).toMatchObject({
      mode: 'jumat',
      phase: 'before',
      secondsLeft: 17 * 60 + 12,
    })
    expect(at('12:09', FRIDAY)).toMatchObject({ mode: 'adzan', jumat: true })
    expect(at('12:20', FRIDAY)).toMatchObject({ mode: 'jumat', phase: 'khutbah' })
    expect(at('12:45', FRIDAY)).toMatchObject({ mode: 'sholat', jumat: true })
    expect(at('13:00', FRIDAY)).toMatchObject({ mode: 'normal', next: { name: 'Ashar' } })
  })
})

describe('clock', () => {
  const date = new Date('2026-10-28T10:27:05Z') // 17:27:05 WIB, 18:27:05 WITA

  it("reads the mosque's wall clock, not the device's", () => {
    expect(wallClock(date, 'Asia/Jakarta')).toEqual({ seconds: parseTime('17:27') + 5, weekday: 3 })
    expect(wallClock(date, 'Asia/Makassar').seconds).toBe(parseTime('18:27') + 5)
  })

  it('formats dates and countdowns like the mockup', () => {
    expect(formatLongDate(date, 'Asia/Jakarta')).toBe('Rabu, 28 Oktober 2026')
    expect(formatHijriDate(date, 'Asia/Jakarta')).toBe('17 Jumadilawal 1448 H')
    expect(formatHms(2865)).toBe('00:47:45')
    expect(formatMs(372)).toBe('06:12')
  })
})
