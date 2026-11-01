import { describe, expect, it } from 'vitest'
import { formatCountdown, nextPrayer } from './nextPrayer'

const times = [
  { name: 'Subuh', time: '04:42' },
  { name: 'Maghrib', time: '18:15' },
  { name: 'Isya', time: '19:25' },
]

describe('nextPrayer', () => {
  it('finds the next prayer today', () => {
    // 17:28 WIB
    expect(nextPrayer(new Date('2026-10-28T10:28:00Z'), times, 'Asia/Jakarta')).toEqual({
      name: 'Maghrib',
      time: '18:15',
      minutesLeft: 47,
    })
  })

  it('rolls over to tomorrow’s Subuh after Isya', () => {
    // 22:00 WIB
    expect(nextPrayer(new Date('2026-10-28T15:00:00Z'), times, 'Asia/Jakarta')).toMatchObject({
      name: 'Subuh',
      minutesLeft: 6 * 60 + 42,
    })
  })

  it("counts on the mosque's clock, not the device's zone", () => {
    // the same instant is 17:28 WIB but 18:28 WITA: Maghrib has passed in Makassar
    const at = new Date('2026-10-28T10:28:00Z')
    expect(nextPrayer(at, times, 'Asia/Makassar')).toMatchObject({ name: 'Isya', minutesLeft: 57 })
  })

  it('formats the countdown as hours:minutes', () => {
    expect(formatCountdown(47)).toBe('00:47')
    expect(formatCountdown(402)).toBe('06:42')
  })
})
