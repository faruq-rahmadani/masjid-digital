import type { CashSummary, Quote } from '@/types/tv'

/*
 * Content for the TV display that has no endpoint yet. Each function returns mock data in the
 * shape the API will return; swap in the real call when its task lands (noted per function).
 * The mosque name, address, time zone and logo are live (mosqueProfileService), and so are
 * the prayer times (computed on the TV from prayerTimeService's config), the running text
 * (runningTextService), the iqamah settings (iqamahService), the petugas Jumat
 * (fridayDutyService) and the slides (slideService).
 */

/** E7: the latest period's cash report. */
export async function getCashSummary(): Promise<CashSummary> {
  return MOCK_CASH
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
