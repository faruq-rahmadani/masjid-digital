import { describe, expect, it } from 'vitest'
import type { PrayerSettings, PrayerTimeOverride } from '@/types/prayer'
import { DEFAULT_PRAYER_SETTINGS, type PrayerLocation, prayerSchedule } from './calculate'
import { minutesOf } from './calendar'

const jakarta: PrayerLocation = { latitude: -6.1754, longitude: 106.8272, timezone: 'Asia/Jakarta' }
const makassar: PrayerLocation = {
  latitude: -5.1333,
  longitude: 119.4167,
  timezone: 'Asia/Makassar',
}
const jayapura: PrayerLocation = {
  latitude: -2.5333,
  longitude: 140.7167,
  timezone: 'Asia/Jayapura',
}

const times = (schedule: ReturnType<typeof prayerSchedule>) =>
  Object.fromEntries(schedule.map((e) => [e.key, e.time]))

describe('prayerSchedule', () => {
  // Bimas Islam Kemenag RI, November 2026 (via liputan6.com/jadwal-sholat), one city per zone.
  // Kemenag computes each city from its own reference point and elevation, so a time on a rounding
  // edge can land a minute apart; the mosque's per-prayer correction aligns it with the local table.
  it.each([
    [
      'Jakarta',
      '2026-11-01',
      jakarta,
      {
        imsak: '03:59',
        subuh: '04:09',
        dzuhur: '11:40',
        ashar: '14:55',
        maghrib: '17:49',
        isya: '19:00',
      },
    ],
    [
      'Jakarta',
      '2026-11-02',
      jakarta,
      {
        imsak: '03:58',
        subuh: '04:08',
        dzuhur: '11:40',
        ashar: '14:55',
        maghrib: '17:49',
        isya: '19:01',
      },
    ],
    [
      'Jakarta',
      '2026-11-06',
      jakarta,
      {
        imsak: '03:57',
        subuh: '04:07',
        dzuhur: '11:40',
        ashar: '14:57',
        maghrib: '17:50',
        isya: '19:02',
      },
    ],
    [
      'Makassar',
      '2026-11-01',
      makassar,
      {
        imsak: '04:10',
        subuh: '04:20',
        dzuhur: '11:49',
        ashar: '15:05',
        maghrib: '17:58',
        isya: '19:09',
      },
    ],
    [
      'Makassar',
      '2026-11-06',
      makassar,
      {
        imsak: '04:08',
        subuh: '04:18',
        dzuhur: '11:49',
        ashar: '15:07',
        maghrib: '17:59',
        isya: '19:10',
      },
    ],
    [
      'Jayapura',
      '2026-11-01',
      jayapura,
      {
        imsak: '03:48',
        subuh: '03:58',
        dzuhur: '11:24',
        ashar: '14:42',
        maghrib: '17:30',
        isya: '18:41',
      },
    ],
    [
      'Jayapura',
      '2026-11-06',
      jayapura,
      {
        imsak: '03:47',
        subuh: '03:57',
        dzuhur: '11:24',
        ashar: '14:44',
        maghrib: '17:30',
        isya: '18:42',
      },
    ],
  ] as const)('is within a minute of Kemenag for %s on %s', (_, date, location, want) => {
    const got = times(prayerSchedule(date, location, DEFAULT_PRAYER_SETTINGS))
    for (const [key, time] of Object.entries(want)) {
      const diff = Math.abs(minutesOf(got[key] ?? '') - minutesOf(time))
      expect(diff, `${key}: got ${got[key]}, Kemenag ${time}`).toBeLessThanOrEqual(1)
    }
  })

  it('matches Kemenag exactly for Jakarta', () => {
    expect(times(prayerSchedule('2026-11-01', jakarta, DEFAULT_PRAYER_SETTINGS))).toMatchObject({
      imsak: '03:59',
      subuh: '04:09',
      dzuhur: '11:40',
      ashar: '14:55',
      maghrib: '17:49',
      isya: '19:00',
    })
  })

  it('applies corrections, then the weekly rule, then the date override', () => {
    const settings: PrayerSettings = {
      ...DEFAULT_PRAYER_SETTINGS,
      corrections: { ...DEFAULT_PRAYER_SETTINGS.corrections, maghrib: 2, dzuhur: 1 },
      imsak_offset_min: 15,
      weekly_overrides: [
        { weekday: 'friday', prayer: 'dzuhur', time: '12:00', remarks: 'Sholat Jumat' },
        { weekday: 'friday', prayer: 'isya', time: '19:15', remarks: '' },
      ],
    }
    const overrides: PrayerTimeOverride[] = [
      {
        id: 1,
        date: '2026-11-06',
        prayer: 'isya',
        time: '19:30',
        remarks: 'Pengajian akbar',
        created_by: 1,
        created_at: '',
        updated_at: '',
      },
      {
        id: 2,
        date: '2026-11-07',
        prayer: 'subuh',
        time: '04:30',
        remarks: 'other day',
        created_by: 1,
        created_at: '',
        updated_at: '',
      },
    ]

    // 2026-11-06 is a Friday
    const day = prayerSchedule('2026-11-06', jakarta, settings, overrides)
    const by = Object.fromEntries(day.map((e) => [e.key, e]))

    expect(by.imsak).toMatchObject({ time: '03:52', source: 'calculated' })
    expect(by.subuh).toMatchObject({ time: '04:07', source: 'calculated' })
    expect(by.maghrib).toMatchObject({ time: '17:52', calculated: '17:52', source: 'calculated' })
    expect(by.dzuhur).toMatchObject({
      time: '12:00',
      calculated: '11:41',
      source: 'weekly',
      remarks: 'Sholat Jumat',
    })
    expect(by.isya).toMatchObject({ time: '19:30', source: 'date', remarks: 'Pengajian akbar' })

    // Thursday: no weekly rule
    expect(times(prayerSchedule('2026-11-05', jakarta, settings, overrides)).dzuhur).toBe('11:41')
  })
})
