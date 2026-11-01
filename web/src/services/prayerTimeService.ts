import {
  apiDelete,
  apiGet,
  apiPost,
  apiPut,
  type DataResponse,
  type MessageResponse,
} from '@/api/client'
import type {
  CreatePrayerOverrideRequest,
  PrayerConfig,
  PrayerSettings,
  PrayerTimeOverride,
} from '@/types/prayer'

export async function getPrayerSettings(): Promise<PrayerSettings> {
  const resp = await apiGet<DataResponse<PrayerSettings>>('/v1/prayer-time/config')
  return resp.data
}

export async function updatePrayerSettings(payload: PrayerSettings): Promise<PrayerSettings> {
  const resp = await apiPut<DataResponse<PrayerSettings>>('/v1/prayer-time/config', payload)
  return resp.data
}

export async function listPrayerOverrides(from: string, to: string): Promise<PrayerTimeOverride[]> {
  const resp = await apiGet<DataResponse<PrayerTimeOverride[]>>('/v1/prayer-time/overrides', {
    params: { from, to },
  })
  return resp.data
}

export async function createPrayerOverride(
  payload: CreatePrayerOverrideRequest,
): Promise<PrayerTimeOverride> {
  const resp = await apiPost<DataResponse<PrayerTimeOverride>>('/v1/prayer-time/overrides', payload)
  return resp.data
}

export async function deletePrayerOverride(id: number): Promise<MessageResponse> {
  return apiDelete<MessageResponse>(`/v1/prayer-time/overrides/${id}`)
}

export async function getPublicPrayerConfig(): Promise<PrayerConfig> {
  const resp = await apiGet<DataResponse<PrayerConfig>>('/public/v1/prayer-time/config')
  return resp.data
}
