import type { CashSummary, Quote, Slide } from '@/types/tv'

/*
 * Content for the TV display that has no endpoint yet. Each function returns mock data in the
 * shape the API will return; swap in the real call when its task lands (noted per function).
 * The mosque name, address, time zone and logo are live (mosqueProfileService), and so are
 * the prayer times (computed on the TV from prayerTimeService's config), the iqamah settings
 * (iqamahService) and the petugas Jumat (fridayDutyService).
 */

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

// --- Mock ----------------------------------------------------------------------------------
// From the mockups (Jakarta).

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
