import { apiGet, type DataResponse } from '@/api/client'
import type { MosqueProfile } from '@/types/mosque'

export async function getMosqueProfile(): Promise<MosqueProfile> {
  const resp = await apiGet<DataResponse<MosqueProfile>>('/v1/mosque-profile')
  return resp.data
}
