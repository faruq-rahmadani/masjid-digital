import { create } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { configureApiAuth } from '@/api/client'
import * as authService from '@/services/authService'
import type { Session } from '@/types/auth'
import { clearToken, loadSession, saveToken, sessionFromToken } from '@/utils/auth/token'

type LoginInput = { identifier: string; password: string; remember: boolean }

type SessionState = {
  session: Session | null
  login: (input: LoginInput) => Promise<Session>
  logout: () => Promise<void>
  /** Drops the local session without calling the API (expired or revoked token). */
  endSession: () => void
}

export const useSessionStore = create<SessionState>()((set, get) => ({
  session: loadSession(),

  login: async ({ identifier, password, remember }) => {
    const { token } = await authService.login({ identifier: identifier.trim(), password })
    const session = sessionFromToken(token)
    if (!session) throw new Error('Token dari server tidak dikenali')
    saveToken(token, remember)
    set({ session })
    return session
  },

  logout: async () => {
    // Revoke on the server when possible; the local session ends either way.
    await authService.logout().catch(() => {})
    get().endSession()
  },

  endSession: () => {
    clearToken()
    set({ session: null })
  },
}))

// The HTTP client reads the token per request and reports a revoked or expired session.
configureApiAuth({
  getToken: () => useSessionStore.getState().session?.token ?? null,
  onUnauthenticated: () => useSessionStore.getState().endSession(),
})

export function useAuth() {
  return useSessionStore(
    useShallow((s) => ({ session: s.session, login: s.login, logout: s.logout })),
  )
}

/** For screens behind <RequireAuth>, where a session always exists. */
export function useSession() {
  const session = useSessionStore((s) => s.session)
  if (!session) throw new Error('useSession must be used behind <RequireAuth>')
  return session
}
