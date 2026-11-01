import { describe, expect, it } from 'vitest'
import { accessTo, canOpen } from './permissions'

describe('permissions (PRD roles table)', () => {
  it('limits user management to super admins', () => {
    expect(canOpen('super_admin', 'users')).toBe(true)
    expect(canOpen('takmir', 'users')).toBe(false)
    expect(canOpen('bendahara', 'users')).toBe(false)
  })

  it('lets takmir edit the mosque profile, like the API', () => {
    expect(accessTo('takmir', 'profile')).toBe('edit')
    expect(canOpen('bendahara', 'profile')).toBe(false)
  })

  it('gives takmir a read-only view of finance and hides categories', () => {
    expect(accessTo('takmir', 'finance')).toBe('view')
    expect(canOpen('takmir', 'categories')).toBe(false)
  })

  it('hides TV and prayer settings from the bendahara', () => {
    expect(canOpen('bendahara', 'screens')).toBe(false)
    expect(canOpen('bendahara', 'prayerTimes')).toBe(false)
    expect(accessTo('bendahara', 'finance')).toBe('edit')
  })
})
