import {
  apiDelete,
  apiGet,
  apiPost,
  apiPut,
  type DataResponse,
  type MessageResponse,
} from '@/api/client'
import type { Screen, ScreenCommand, UpdateScreenRequest } from '@/types/screen'
import { env } from '@/utils/env'

/** Every screen, blocked ones too (the page shows them apart). */
export async function listScreens(): Promise<Screen[]> {
  const resp = await apiGet<DataResponse<Screen[]>>('/v1/screens')
  return resp.data
}

/** Names the screen and sets its mode; a forced mode lasts until the mosque's midnight. */
export async function updateScreen(id: number, payload: UpdateScreenRequest): Promise<Screen> {
  const resp = await apiPut<DataResponse<Screen>>(`/v1/screens/${id}`, payload)
  return resp.data
}

export async function blockScreen(id: number): Promise<Screen> {
  const resp = await apiPost<DataResponse<Screen>>(`/v1/screens/${id}/block`)
  return resp.data
}

export async function unblockScreen(id: number): Promise<Screen> {
  const resp = await apiDelete<DataResponse<Screen>>(`/v1/screens/${id}/block`)
  return resp.data
}

/** To one screen, or to every connected one when `id` is left out. */
export async function sendScreenCommand(command: ScreenCommand, id?: number) {
  const url = id === undefined ? '/v1/screens/commands' : `/v1/screens/${id}/commands`
  return apiPost<MessageResponse>(url, { command })
}

/** The TV's live connection (server-sent events); public, the TV has no login. */
export function screenEventsUrl(deviceId: string, nameHint: string, appVersion: string) {
  const query = new URLSearchParams({ name: nameHint, v: appVersion })
  return `${env.apiBaseUrl}/public/v1/screens/${encodeURIComponent(deviceId)}/events?${query}`
}
