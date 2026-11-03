import { apiGet, type DataResponse } from '@/api/client'
import type { Activity } from '@/types/home'

/** How many changes Home shows under "Aktivitas terakhir". */
const RECENT_LIMIT = 10

/** The latest changes, newest first (E2-05). */
export async function listRecentActivity(): Promise<Activity[]> {
  const resp = await apiGet<DataResponse<Activity[]>>(`/v1/activity-logs?limit=${RECENT_LIMIT}`)
  return resp.data
}
