import type { Role } from '@/types/auth'

// Shapes of the Go API (internal/delivery/http/controller/user.go).

export type UserStatus = 'active' | 'disabled'

export type User = {
  id: number
  name: string
  /** E.164, e.g. +6281234561001 */
  phone: string
  /** Empty when not set. */
  email: string
  role: Role
  status: UserStatus
  last_login_at: string | null
  created_by: number
  created_at: string
  updated_at: string
}

export type CreateUserRequest = Pick<User, 'name' | 'phone' | 'email' | 'role'> & {
  password: string
}

/** PUT replaces every editable field; an empty email clears it. */
export type UpdateUserRequest = Pick<User, 'name' | 'phone' | 'email' | 'role' | 'status'>
