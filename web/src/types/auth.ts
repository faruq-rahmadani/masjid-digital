// Shapes of the Go API (internal/domain/auth.go, user.go).

export type Role = 'super_admin' | 'takmir' | 'bendahara'

export type LoginRequest = { identifier: string; password: string }
export type LoginResponse = { token: string; expire_in: number }

/** The logged-in user as known from the token. The API has no /me yet, so no name here. */
export type Session = {
  token: string
  userId: number
  role: Role
  /** Epoch milliseconds. */
  expiresAt: number
}
