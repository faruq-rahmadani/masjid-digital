import { apiGet, apiPut, type DataResponse } from '@/api/client'
import type { IqamahSettings } from '@/types/iqamah'

/** The defaults until first saved. */
export async function getIqamahSettings(): Promise<IqamahSettings> {
  const resp = await apiGet<DataResponse<IqamahSettings>>('/v1/iqamah/config')
  return resp.data
}

/** Replaces every setting. */
export async function updateIqamahSettings(payload: IqamahSettings): Promise<IqamahSettings> {
  const resp = await apiPut<DataResponse<IqamahSettings>>('/v1/iqamah/config', payload)
  return resp.data
}

/** Public (no login), for the TV. */
export async function getPublicIqamahSettings(): Promise<IqamahSettings> {
  const resp = await apiGet<DataResponse<IqamahSettings>>('/public/v1/iqamah/config')
  return resp.data
}
