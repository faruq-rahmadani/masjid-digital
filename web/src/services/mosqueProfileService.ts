import { apiDelete, apiGet, apiPut, type DataResponse, type MessageResponse } from '@/api/client'
import type { MosqueLogo, MosqueProfile, UpdateMosqueProfileRequest } from '@/types/mosque'

export async function getMosqueProfile(): Promise<MosqueProfile> {
  const resp = await apiGet<DataResponse<MosqueProfile>>('/v1/mosque-profile')
  return resp.data
}

export async function updateMosqueProfile(
  payload: UpdateMosqueProfileRequest,
): Promise<MosqueProfile> {
  const resp = await apiPut<DataResponse<MosqueProfile>>('/v1/mosque-profile', payload)
  return resp.data
}

export async function getMosqueLogo(): Promise<MosqueLogo> {
  const resp = await apiGet<DataResponse<MosqueLogo>>('/v1/mosque-profile/logo')
  return resp.data
}

export async function uploadMosqueLogo(file: File): Promise<MessageResponse> {
  const body = new FormData()
  body.append('file', file)
  return apiPut<MessageResponse>('/v1/mosque-profile/logo', body)
}

export async function deleteMosqueLogo(): Promise<MessageResponse> {
  return apiDelete<MessageResponse>('/v1/mosque-profile/logo')
}

// Public twins of the reads above (no login), for the TV display.

export async function getPublicMosqueProfile(): Promise<MosqueProfile> {
  const resp = await apiGet<DataResponse<MosqueProfile>>('/public/v1/mosque-profile')
  return resp.data
}

export async function getPublicMosqueLogo(): Promise<MosqueLogo> {
  const resp = await apiGet<DataResponse<MosqueLogo>>('/public/v1/mosque-profile/logo')
  return resp.data
}
