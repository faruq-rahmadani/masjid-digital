import { apiGet, apiPost, type DataResponse, type MessageResponse } from '@/api/client'
import type { LoginRequest, LoginResponse } from '@/types/auth'
import type { User } from '@/types/user'

export async function login(payload: LoginRequest): Promise<LoginResponse> {
  return apiPost<LoginResponse>('/v1/auth/login', payload)
}

export async function logout(): Promise<MessageResponse> {
  return apiPost<MessageResponse>('/v1/auth/logout')
}

export async function getProfile(): Promise<User> {
  const resp = await apiGet<DataResponse<User>>('/v1/auth/profile')
  return resp.data
}
