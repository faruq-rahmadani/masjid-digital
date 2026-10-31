import type { Role, Session } from '@/types/auth'
import { ROLES } from '@/utils/auth/roles'

const KEY = 'masjid-digital:session'

// Reads the claims the Go API puts in its JWT (pkg/token/jwt.go): sub, role, exp.
// Not a verification; the server checks the signature on every request.
export function sessionFromToken(token: string): Session | null {
  try {
    const payload = token.split('.')[1]
    if (!payload) return null
    const json = atob(payload.replace(/-/g, '+').replace(/_/g, '/'))
    const claims = JSON.parse(json) as { sub?: string; role?: string; exp?: number }
    const userId = Number(claims.sub)
    const role = claims.role as Role
    if (!Number.isInteger(userId) || !ROLES.includes(role) || typeof claims.exp !== 'number') {
      return null
    }
    return { token, userId, role, expiresAt: claims.exp * 1000 }
  } catch {
    return null
  }
}

function isExpired(session: Session, now = Date.now()) {
  return session.expiresAt <= now
}

// "Ingat saya" keeps the token across browser restarts; otherwise it ends with the tab.
export function saveToken(token: string, remember: boolean) {
  clearToken()
  try {
    ;(remember ? localStorage : sessionStorage).setItem(KEY, token)
  } catch {
    // Private mode or blocked storage: the session lasts until reload.
  }
}

export function loadSession(): Session | null {
  try {
    const token = localStorage.getItem(KEY) ?? sessionStorage.getItem(KEY)
    const session = token ? sessionFromToken(token) : null
    if (!session || isExpired(session)) {
      clearToken()
      return null
    }
    return session
  } catch {
    return null
  }
}

export function clearToken() {
  try {
    localStorage.removeItem(KEY)
    sessionStorage.removeItem(KEY)
  } catch {
    // Nothing stored.
  }
}
