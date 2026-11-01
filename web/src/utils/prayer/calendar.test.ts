import { describe, expect, it } from 'vitest'
import { addDays, clockOf, localDate, weekdayOf } from './calendar'

describe('calendar', () => {
  it('reads the date in the mosque time zone', () => {
    const at = new Date('2026-11-01T20:00:00Z')
    expect(localDate(at, 'Asia/Jakarta')).toBe('2026-11-02')
    expect(localDate(new Date('2026-11-01T16:30:00Z'), 'Asia/Jakarta')).toBe('2026-11-01')
  })

  it('does date math across months', () => {
    expect(addDays('2026-11-30', 2)).toBe('2026-12-02')
    expect(addDays('2027-03-01', -1)).toBe('2027-02-28')
    expect(weekdayOf('2026-11-06')).toBe('friday')
  })

  it('wraps clocks around midnight', () => {
    expect(clockOf(245)).toBe('04:05')
    expect(clockOf(-5)).toBe('23:55')
  })
})
