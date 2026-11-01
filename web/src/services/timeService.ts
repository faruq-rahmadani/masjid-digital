import { apiGet, type DataResponse } from '@/api/client'

/** The server's clock as Unix milliseconds (public, no login). */
export async function getServerTime(): Promise<number> {
  const resp = await apiGet<DataResponse<{ now: number }>>('/public/v1/time')
  return resp.data.now
}
