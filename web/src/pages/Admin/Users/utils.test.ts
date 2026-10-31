import { describe, expect, it } from 'vitest'
import type { User } from '@/types/user'
import { filterUsers, isLastActiveSuperAdmin, sortUsers } from './utils'

const user = (id: number, role: User['role'], status: User['status'] = 'active'): User => ({
  id,
  name: `User ${id}`,
  phone: '',
  email: '',
  role,
  status,
  last_login_at: null,
  created_by: 0,
  created_at: '',
  updated_at: '',
})

describe('users', () => {
  it('lists you first, active by role, disabled last', () => {
    const users = [
      user(1, 'bendahara', 'disabled'),
      user(2, 'bendahara'),
      user(3, 'takmir'),
      user(4, 'super_admin'),
    ]
    expect(sortUsers(users, 3).map((u) => u.id)).toEqual([3, 4, 2, 1])
  })

  it('protects the last active super admin', () => {
    const only = user(1, 'super_admin')
    expect(isLastActiveSuperAdmin(only, [only, user(2, 'takmir')])).toBe(true)
    expect(isLastActiveSuperAdmin(only, [only, user(2, 'super_admin')])).toBe(false)
    expect(isLastActiveSuperAdmin(only, [only, user(2, 'super_admin', 'disabled')])).toBe(true)
  })

  it('filters by name and role', () => {
    const users = [
      { ...user(1, 'takmir'), name: 'Hasan Basri' },
      { ...user(2, 'bendahara'), name: 'Siti Aminah' },
      { ...user(3, 'takmir'), name: 'Umar Hadi' },
    ]
    expect(filterUsers(users, { query: 'HASAN', role: 'all' }).map((u) => u.id)).toEqual([1])
    expect(filterUsers(users, { query: '', role: 'takmir' }).map((u) => u.id)).toEqual([1, 3])
    expect(filterUsers(users, { query: 'siti', role: 'takmir' })).toEqual([])
  })
})
