import type { Role } from '@/types/auth'
import type { User } from '@/types/user'
import { ROLES } from '@/utils/auth/roles'

// You first, then active users by role (as in the mockup), then disabled ones; by name within.
export function sortUsers(users: readonly User[], selfId: number) {
  const rank = (u: User) =>
    u.id === selfId ? -1 : u.status === 'disabled' ? ROLES.length : ROLES.indexOf(u.role)
  return [...users].sort((a, b) => rank(a) - rank(b) || a.name.localeCompare(b.name, 'id'))
}

/** The API only blocks changing your own role; keep at least one active super admin (PRD). */
export function isLastActiveSuperAdmin(user: User, users: readonly User[]) {
  if (user.role !== 'super_admin' || user.status !== 'active') return false
  return users.filter((u) => u.role === 'super_admin' && u.status === 'active').length === 1
}

export type UserFilter = { query: string; role: Role | 'all' }

/** Case-insensitive name match plus an optional role. */
export function filterUsers(users: readonly User[], { query, role }: UserFilter) {
  const needle = query.trim().toLocaleLowerCase('id')
  return users.filter(
    (u) =>
      (role === 'all' || u.role === role) &&
      (needle === '' || u.name.toLocaleLowerCase('id').includes(needle)),
  )
}
