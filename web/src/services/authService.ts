import { apiPost, type MessageResponse } from '@/api/client'
import type { LoginRequest, LoginResponse, Role } from '@/types/auth'
import { normalizePhone } from '@/utils/format/phone'

// POST /api/v1/auth/* (internal/delivery/http/router.go)
// Mocked while the app isn't wired to the API: to go live, uncomment the real lines and delete
// the mock line below them (and the mock section at the bottom once nothing uses it).

export async function login(payload: LoginRequest): Promise<LoginResponse> {
  // return apiPost<LoginResponse>('/v1/auth/login', payload)
  return mockLogin(payload)
}

export async function logout(): Promise<MessageResponse> {
  // return apiPost<MessageResponse>('/v1/auth/logout')
  return { message: 'logout berhasil' }
}

// --- Mock ----------------------------------------------------------------------------------
// Any of these logs in with MOCK_PASSWORD; ids match MOCK_USERS in userService.ts.

export const MOCK_PASSWORD = 'masjid123'

const MOCK_ACCOUNTS: { id: number; login: string; role: Role }[] = [
  { id: 1, login: 'ketua@masjid.id', role: 'super_admin' },
  { id: 2, login: '+6281234561001', role: 'takmir' },
  { id: 3, login: 'bendahara@masjid.id', role: 'bendahara' },
]

function mockLogin({ identifier, password }: LoginRequest): LoginResponse {
  const id = identifier.includes('@') ? identifier.trim().toLowerCase() : normalizePhone(identifier)
  const account = MOCK_ACCOUNTS.find((a) => a.login === id)
  if (!account || password !== MOCK_PASSWORD) throw new Error('Nomor HP/email atau password salah')
  return mockToken(account.id, account.role)
}

/** Unsigned JWT with the Go API's claims (sub, role, exp), so the app decodes it the same way. */
export function mockToken(userId: number, role: Role): LoginResponse {
  const encode = (value: object) => btoa(JSON.stringify(value)).replace(/=+$/, '')
  const exp = Math.floor(Date.now() / 1000) + 24 * 60 * 60
  return {
    token: `${encode({ alg: 'none' })}.${encode({ sub: String(userId), role, exp })}.mock`,
    expire_in: 24 * 60 * 60,
  }
}
