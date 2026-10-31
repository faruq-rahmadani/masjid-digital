import type { PrayerTime } from '@/types/home'

// Not wired yet. Per the PRD, times are computed on the device from the mosque's coordinates
// (Kemenag method, E4-01) plus per-waktu corrections from the API (E4-02). Mocked until then.

export async function getTodayPrayerTimes(): Promise<PrayerTime[]> {
  return MOCK_PRAYER_TIMES
}

// --- Mock ----------------------------------------------------------------------------------
// Jakarta, from the mockup.

const MOCK_PRAYER_TIMES: PrayerTime[] = [
  { name: 'Subuh', time: '04:42' },
  { name: 'Dzuhur', time: '12:09' },
  { name: 'Ashar', time: '15:24' },
  { name: 'Maghrib', time: '18:15' },
  { name: 'Isya', time: '19:25' },
]
