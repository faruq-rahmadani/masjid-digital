import {
  apiDelete,
  apiGet,
  apiPost,
  apiPut,
  type DataResponse,
  type MessageResponse,
} from '@/api/client'
import type { CreateSlideRequest, DisplaySlide, Slide, UpdateSlideRequest } from '@/types/slide'
import { env } from '@/utils/env'

/**
 * A slide's image, public so an <img> needs no token. It never changes (a new picture is a new
 * slide), so browsers cache it for good.
 */
export function slideImageUrl(id: number) {
  return `${env.apiBaseUrl}/public/v1/slides/${id}/image`
}

export async function listSlides(): Promise<Slide[]> {
  const resp = await apiGet<DataResponse<Slide[]>>('/v1/slides')
  return resp.data
}

/** One image per request, uploaded with its details as multipart form fields. */
export async function createSlide(file: File, payload: CreateSlideRequest): Promise<Slide> {
  const body = new FormData()
  body.append('file', file)
  body.append('title', payload.title)
  body.append('caption', payload.caption)
  body.append('duration_seconds', String(payload.duration_seconds))
  body.append('show_caption', String(payload.show_caption))
  const resp = await apiPost<DataResponse<Slide>>('/v1/slides', body)
  return resp.data
}

export async function updateSlide(id: number, payload: UpdateSlideRequest): Promise<Slide> {
  const resp = await apiPut<DataResponse<Slide>>(`/v1/slides/${id}`, payload)
  return resp.data
}

export async function deleteSlide(id: number): Promise<MessageResponse> {
  return apiDelete<MessageResponse>(`/v1/slides/${id}`)
}

/** `ids` must list every slide once, in the new order. */
export async function reorderSlides(ids: number[]): Promise<Slide[]> {
  const resp = await apiPut<DataResponse<Slide[]>>('/v1/slides/order', { ids })
  return resp.data
}

export async function getPublicSlides(): Promise<DisplaySlide[]> {
  const resp = await apiGet<DataResponse<DisplaySlide[]>>('/public/v1/slides')
  return resp.data
}
