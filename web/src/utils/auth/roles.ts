import type { Role } from '@/types/auth'

export const ROLES = ['super_admin', 'takmir', 'bendahara'] as const satisfies readonly Role[]

export const roleLabels: Record<Role, string> = {
  super_admin: 'Super admin',
  takmir: 'Takmir',
  bendahara: 'Bendahara',
}

/** Avatar and badge tone per role, as in the Pengguna mockup. */
export const roleTones = {
  super_admin: 'brand',
  takmir: 'success',
  bendahara: 'gold',
} as const satisfies Record<Role, string>
