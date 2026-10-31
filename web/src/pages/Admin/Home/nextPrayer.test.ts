import { describe, expect, it } from 'vitest'
import { formatCountdown, nextPrayer } from './nextPrayer'

const times = [
  { name: 'Subuh', time: '04:42' },
  { name: 'Maghrib', time: '18:15' },
  { name: 'Isya', time: '19:25' },
]

describe('nextPrayer', () => {
  it('finds the next prayer today', () => {
    expect(nextPrayer(new Date(2026, 9, 28, 17, 28), times)).toEqual({
      name: 'Maghrib',
      time: '18:15',
      minutesLeft: 47,
    })
  })

  it('rolls over to tomorrow’s Subuh after Isya', () => {
    expect(nextPrayer(new Date(2026, 9, 28, 22, 0), times)).toMatchObject({
      name: 'Subuh',
      minutesLeft: 6 * 60 + 42,
    })
  })

  it('formats the countdown as hours:minutes', () => {
    expect(formatCountdown(47)).toBe('00:47')
    expect(formatCountdown(402)).toBe('06:42')
  })
})
