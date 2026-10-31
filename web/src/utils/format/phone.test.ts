import { describe, expect, it } from 'vitest'
import { formatPhone, normalizePhone } from './phone'

describe('phone', () => {
  it('normalizes Indonesian numbers to E.164 like the API', () => {
    expect(normalizePhone('0812-3456-1001')).toBe('+6281234561001')
    expect(normalizePhone('62 812 3456 1001')).toBe('+6281234561001')
    expect(normalizePhone('+6281234561001')).toBe('+6281234561001')
  })

  it('rejects what the API rejects', () => {
    expect(normalizePhone('812345')).toBeNull() // no prefix
    expect(normalizePhone('021-5551234')).toBeNull() // landline, not mobile
    expect(normalizePhone('08123')).toBeNull() // too short
  })

  it('formats E.164 for display', () => {
    expect(formatPhone('+6281234561001')).toBe('0812-3456-1001')
    expect(formatPhone('+14155550123')).toBe('+14155550123')
  })
})
