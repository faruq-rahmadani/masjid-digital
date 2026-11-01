import type {
  CashSummary,
  DutyRoster,
  IqamahSettings,
  Quote,
  ScheduleItem,
  Slide,
} from '@/types/tv'

/*
 * Content for the TV display that has no endpoint yet. Each function returns mock data in the
 * shape the API will return; swap in the real call when its task lands (noted per function).
 * The mosque name, address, time zone and logo are live (mosqueProfileService).
 */

/** E4-01: computed on the device from the mosque's coordinates (Kemenag method). */
export async function getTodaySchedule(): Promise<ScheduleItem[]> {
  return MOCK_SCHEDULE
}

/** E7: the latest period's cash report. */
export async function getCashSummary(): Promise<CashSummary> {
  return MOCK_CASH
}

/** E6-04: active slides in display order. */
export async function listSlides(): Promise<Slide[]> {
  return MOCK_SLIDES
}

/** E6-05: active running texts in display order. */
export async function listRunningTexts(): Promise<string[]> {
  return MOCK_RUNNING_TEXTS
}

/** Quotes rotating under the clock. */
export async function listQuotes(): Promise<Quote[]> {
  return MOCK_QUOTES
}

/** E5-04 */
export async function getIqamahSettings(): Promise<IqamahSettings> {
  return MOCK_IQAMAH
}

/** E5-05 */
export async function getDutyRoster(): Promise<DutyRoster> {
  return MOCK_DUTY
}

// --- Mock ----------------------------------------------------------------------------------
// From the mockups (Jakarta).

const MOCK_SCHEDULE: ScheduleItem[] = [
  { key: 'imsak', name: 'Imsak', time: '04:32' },
  { key: 'subuh', name: 'Subuh', time: '04:42' },
  { key: 'terbit', name: 'Terbit', time: '05:56' },
  { key: 'dzuhur', name: 'Dzuhur', time: '12:09' },
  { key: 'ashar', name: 'Ashar', time: '15:24' },
  { key: 'maghrib', name: 'Maghrib', time: '18:15' },
  { key: 'isya', name: 'Isya', time: '19:25' },
]

const MOCK_CASH: CashSummary = {
  as_of: '2026-10-28',
  opening_balance: 91_934_781,
  income: 1_250_000,
  expense: 1_126_667,
  closing_balance: 92_058_114,
}

const MOCK_SLIDES: Slide[] = [
  {
    id: 1,
    title: 'Kajian Rutin Ahad Pagi',
    caption: 'Ahad, 08.00 WIB · Ustadz Abdullah',
    image_url: '',
  },
  { id: 2, title: 'Kajian Tafsir', caption: "Kamis ba'da Maghrib · Ruang utama", image_url: '' },
  { id: 3, title: 'Santunan Anak Yatim', caption: 'Sabtu, 15 November 2026', image_url: '' },
  {
    id: 4,
    title: 'Infaq Pembangunan',
    caption: 'QRIS tersedia di dekat pintu utama',
    image_url: '',
  },
]

const MOCK_RUNNING_TEXTS = [
  "Kajian Tafsir setiap Kamis ba'da Maghrib",
  'Infaq via QRIS tersedia di dekat pintu utama',
  'Mohon matikan nada dering HP selama di dalam masjid',
]

const MOCK_QUOTES: Quote[] = [
  {
    text: 'Sesungguhnya hanya orang-orang yang bersabarlah yang dicukupkan pahala mereka tanpa batas.',
    source: 'QS. Az-Zumar: 10',
  },
  {
    text: 'Maka sesungguhnya bersama kesulitan ada kemudahan.',
    source: 'QS. Al-Insyirah: 5',
  },
  {
    text: 'Sebaik-baik manusia adalah yang paling bermanfaat bagi manusia lainnya.',
    source: 'HR. Ahmad',
  },
]

const MOCK_IQAMAH: IqamahSettings = {
  iqamah_minutes: { subuh: 10, dzuhur: 7, ashar: 7, maghrib: 7, isya: 7 },
  adzan_minutes: 1,
  prayer_minutes: 10,
  jumat_before_minutes: 60,
  khutbah_minutes: 30,
}

const MOCK_DUTY: DutyRoster = {
  imam: 'Ustadz Hasan',
  muadzin: 'Bilal Ahmad',
  khatib: 'Ustadz Abdullah',
  khutbah_theme: 'Menjaga amanah di tengah keluarga',
}
