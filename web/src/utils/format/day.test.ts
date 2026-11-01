import { describe, expect, it } from 'vitest'
import { formatLongDay, formatShortDay } from './day'

describe('day', () => {
  it('formats the mosque date whatever the device zone', () => {
    expect(formatShortDay('2026-11-05')).toBe('Kam 5')
    expect(formatLongDay('2026-11-06')).toBe('Jumat, 6 Nov')
  })
})
