import { apiDelete, apiGet, apiPut, type DataResponse, type MessageResponse } from '@/api/client'
import type { FridayDuty, SaveFridayDutyRequest } from '@/types/fridayDuty'

/** Petugas Jumat from `from` to `to` inclusive, "YYYY-MM-DD". */
export async function listFridayDuties(from: string, to: string): Promise<FridayDuty[]> {
  const resp = await apiGet<DataResponse<FridayDuty[]>>('/v1/friday-duties', {
    params: { from, to },
  })
  return resp.data
}

/** Sets the petugas of the Friday `date`, replacing what was there. */
export async function saveFridayDuty(
  date: string,
  payload: SaveFridayDutyRequest,
): Promise<FridayDuty> {
  const resp = await apiPut<DataResponse<FridayDuty>>(`/v1/friday-duties/${date}`, payload)
  return resp.data
}

export async function deleteFridayDuty(date: string): Promise<MessageResponse> {
  return apiDelete<MessageResponse>(`/v1/friday-duties/${date}`)
}

/** Public (no login), for the TV: the petugas of the coming 30 days. */
export async function listUpcomingFridayDuties(): Promise<FridayDuty[]> {
  const resp = await apiGet<DataResponse<FridayDuty[]>>('/public/v1/friday-duties')
  return resp.data
}
