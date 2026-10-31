import { apiGet, apiPost, apiPut, type DataResponse, type MessageResponse } from '@/api/client'
import type { CreateUserRequest, UpdateUserRequest, User } from '@/types/user'
import { normalizePhone } from '@/utils/format/phone'

// /api/v1/users, super admin only (internal/delivery/http/router.go)
// Mocked while the app isn't wired to the API: to go live, uncomment the real lines and delete
// the mock lines below them (and the mock section at the bottom once nothing uses it).

export async function listUsers(): Promise<User[]> {
  // const resp = await apiGet<DataResponse<User[]>>('/v1/users')
  // return resp.data
  return MOCK_USERS.map((u) => ({ ...u }))
}

export async function getUser(id: number): Promise<User> {
  // const resp = await apiGet<DataResponse<User>>(`/v1/users/${id}`)
  // return resp.data
  const user = MOCK_USERS.find((u) => u.id === id)
  if (!user) throw new Error('User tidak ditemukan')
  return { ...user }
}

export async function createUser(payload: CreateUserRequest): Promise<MessageResponse> {
  // return apiPost<MessageResponse>('/v1/users', payload)
  const { password: _password, ...user } = payload
  MOCK_USERS.push({
    ...mockUser({ ...user, status: 'active' }),
    id: Math.max(...MOCK_USERS.map((u) => u.id)) + 1,
    phone: normalizePhone(user.phone) ?? user.phone,
  })
  return { message: 'user berhasil dibuat' }
}

export async function updateUser(id: number, payload: UpdateUserRequest): Promise<User> {
  // const resp = await apiPut<DataResponse<User>>(`/v1/users/${id}`, payload)
  // return resp.data
  const index = MOCK_USERS.findIndex((u) => u.id === id)
  if (index < 0) throw new Error('User tidak ditemukan')
  MOCK_USERS[index] = { ...MOCK_USERS[index], ...payload } as User
  return { ...MOCK_USERS[index] } as User
}

export async function resetUserPassword(id: number, newPassword: string): Promise<MessageResponse> {
  // return apiPut<MessageResponse>(`/v1/users/${id}/password`, { new_password: newPassword })
  return { message: 'password berhasil direset' }
}

// --- Mock ----------------------------------------------------------------------------------
// In memory, resets on reload. Functions return copies, like a real response would. Login accounts for these users are in authService.ts.

function mockUser(user: Pick<User, 'name' | 'phone' | 'email' | 'role' | 'status'>): User {
  const now = '2026-09-01T08:00:00.000Z'
  return { id: 0, ...user, last_login_at: null, created_by: 1, created_at: now, updated_at: now }
}

const MOCK_USERS: User[] = [
  {
    ...mockUser({
      name: 'Ahmad Fauzi',
      phone: '+6281234561000',
      email: 'ketua@masjid.id',
      role: 'super_admin',
      status: 'active',
    }),
    id: 1,
    created_by: 0,
  },
  {
    ...mockUser({
      name: 'Hasan Basri',
      phone: '+6281234561001',
      email: '',
      role: 'takmir',
      status: 'active',
    }),
    id: 2,
  },
  {
    ...mockUser({
      name: 'Siti Aminah',
      phone: '+6281234561002',
      email: 'bendahara@masjid.id',
      role: 'bendahara',
      status: 'active',
    }),
    id: 3,
  },
  {
    ...mockUser({
      name: 'Umar Hadi',
      phone: '+6281234561003',
      email: '',
      role: 'takmir',
      status: 'disabled',
    }),
    id: 4,
  },
]
