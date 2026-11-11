import { apiGet, type DataResponse } from '@/api/client'
import type { FinanceDisplay } from '@/types/finance'
import type { Quote } from '@/types/tv'

/*
 * Content for the TV display that the other services don't cover: the kas header (live) and the
 * quotes (mocked until they get an endpoint).
 * The mosque name, address, time zone and logo are live (mosqueProfileService), and so are
 * the prayer times (computed on the TV from prayerTimeService's config), the running text
 * (runningTextService), the iqamah settings (iqamahService), the petugas Jumat
 * (fridayDutyService) and the slides (slideService).
 */

/** E7-04: the kas marked "tampil di layar TV" over the period pengurus chose; public. */
export async function getCashDisplay(): Promise<FinanceDisplay> {
  const resp = await apiGet<DataResponse<FinanceDisplay>>('/public/v1/finance/display')
  return resp.data
}

/** Quotes rotating under the clock. */
export async function listQuotes(): Promise<Quote[]> {
  return MOCK_QUOTES
}

// --- Mock ----------------------------------------------------------------------------------
// From the mockups (Jakarta).

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
