import { describe, expect, it } from 'vitest'
import { formatActivityTime, formatTime, formatTimeAgo } from './date'

const now = new Date(2026, 9, 28, 17, 30)
const at = (d: number, h: number, m: number) => new Date(2026, 9, d, h, m).toISOString()

describe('date', () => {
  it('writes times 24h with a colon', () => {
    expect(formatTime(new Date(2026, 9, 28, 18, 5))).toBe('18:05')
  })

  it('formats activity times by day', () => {
    expect(formatActivityTime(at(28, 16, 2), now)).toBe('16:02')
    expect(formatActivityTime(at(27, 23, 0), now)).toBe('Kemarin')
    expect(formatActivityTime(at(25, 9, 0), now)).toBe('25 Okt')
  })

  it('formats how long ago', () => {
    expect(formatTimeAgo(at(28, 17, 30), now)).toBe('baru saja')
    expect(formatTimeAgo(at(28, 17, 25), now)).toBe('5 menit lalu')
    expect(formatTimeAgo(at(28, 15, 10), now)).toBe('2 jam lalu')
    expect(formatTimeAgo(at(25, 17, 30), now)).toBe('3 hari lalu')
  })
})
