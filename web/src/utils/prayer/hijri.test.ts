import { describe, expect, it } from 'vitest'
import { hijriDate } from './hijri'

describe('hijriDate', () => {
  it('formats the Umm al-Qura date in Indonesian', () => {
    expect(hijriDate('2026-11-01', { hijri_offset_days: 0, hijri_offset_from: '' })).toBe(
      '21 Jumadilawal 1448 H',
    )
  })

  it('shifts by the offset from its start date only', () => {
    const settings = { hijri_offset_days: -1, hijri_offset_from: '2027-02-08' }
    expect(hijriDate('2027-02-07', settings)).toBe('30 Syakban 1448 H')
    expect(hijriDate('2027-02-08', settings)).toBe('30 Syakban 1448 H')
    expect(hijriDate('2027-02-09', settings)).toBe('1 Ramadan 1448 H')
  })
})
