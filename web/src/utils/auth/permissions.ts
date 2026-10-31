import type { Role } from '@/types/auth'

// docs/product/prd.md#roles-and-permissions. Missing role = menu hidden.

export type ModuleKey =
  | 'home'
  | 'profile'
  | 'users'
  | 'screens'
  | 'prayerTimes'
  | 'iqamah'
  | 'content'
  | 'fridayDuty'
  | 'finance'
  | 'categories'

type Access = 'edit' | 'view'

const MATRIX: Record<ModuleKey, Partial<Record<Role, Access>>> = {
  home: { super_admin: 'edit', takmir: 'view', bendahara: 'view' },
  profile: { super_admin: 'edit', takmir: 'view' },
  users: { super_admin: 'edit' },
  screens: { super_admin: 'edit', takmir: 'edit' },
  prayerTimes: { super_admin: 'edit', takmir: 'edit' },
  iqamah: { super_admin: 'edit', takmir: 'edit' },
  content: { super_admin: 'edit', takmir: 'edit' },
  fridayDuty: { super_admin: 'edit', takmir: 'edit' },
  finance: { super_admin: 'edit', takmir: 'view', bendahara: 'edit' },
  categories: { super_admin: 'edit', bendahara: 'edit' },
}

export function accessTo(role: Role, module: ModuleKey): Access | null {
  return MATRIX[module][role] ?? null
}

export function canOpen(role: Role, module: ModuleKey) {
  return accessTo(role, module) !== null
}
