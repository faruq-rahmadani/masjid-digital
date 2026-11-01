import { apiGet, apiPost, apiPut, type DataResponse, type MessageResponse } from '@/api/client'
import type { CreateUserRequest, UpdateUserRequest, User } from '@/types/user'

export async function listUsers(): Promise<User[]> {
  const resp = await apiGet<DataResponse<User[]>>('/v1/users')
  return resp.data
}

export async function createUser(payload: CreateUserRequest): Promise<MessageResponse> {
  return apiPost<MessageResponse>('/v1/users', payload)
}

export async function updateUser(id: number, payload: UpdateUserRequest): Promise<User> {
  const resp = await apiPut<DataResponse<User>>(`/v1/users/${id}`, payload)
  return resp.data
}

export async function resetUserPassword(id: number, newPassword: string): Promise<MessageResponse> {
  return apiPut<MessageResponse>(`/v1/users/${id}/password`, { new_password: newPassword })
}
