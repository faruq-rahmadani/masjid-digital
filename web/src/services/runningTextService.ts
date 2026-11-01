import {
  apiDelete,
  apiGet,
  apiPost,
  apiPut,
  type DataResponse,
  type MessageResponse,
} from '@/api/client'
import type {
  RunningText,
  RunningTextDisplay,
  RunningTextSetting,
  UpdateRunningTextRequest,
} from '@/types/runningText'

export async function listRunningTexts(): Promise<RunningText[]> {
  const resp = await apiGet<DataResponse<RunningText[]>>('/v1/running-texts')
  return resp.data
}

export async function createRunningText(text: string): Promise<RunningText> {
  const resp = await apiPost<DataResponse<RunningText>>('/v1/running-texts', { text })
  return resp.data
}

export async function updateRunningText(
  id: number,
  payload: UpdateRunningTextRequest,
): Promise<RunningText> {
  const resp = await apiPut<DataResponse<RunningText>>(`/v1/running-texts/${id}`, payload)
  return resp.data
}

export async function deleteRunningText(id: number): Promise<MessageResponse> {
  return apiDelete<MessageResponse>(`/v1/running-texts/${id}`)
}

/** `ids` must list every running text once, in the new order. */
export async function reorderRunningTexts(ids: number[]): Promise<RunningText[]> {
  const resp = await apiPut<DataResponse<RunningText[]>>('/v1/running-texts/order', { ids })
  return resp.data
}

export async function getRunningTextSetting(): Promise<RunningTextSetting> {
  const resp = await apiGet<DataResponse<RunningTextSetting>>('/v1/running-text/setting')
  return resp.data
}

export async function updateRunningTextSetting(
  payload: RunningTextSetting,
): Promise<RunningTextSetting> {
  const resp = await apiPut<DataResponse<RunningTextSetting>>('/v1/running-text/setting', payload)
  return resp.data
}

export async function getPublicRunningTexts(): Promise<RunningTextDisplay> {
  const resp = await apiGet<DataResponse<RunningTextDisplay>>('/public/v1/running-texts')
  return resp.data
}
