import { describe, expect, it } from 'vitest'
import type { Screen } from '@/types/screen'
import { currentMode, deviceLabel, groupScreens, seenLine } from './utils'

const now = new Date('2026-11-02T10:00:00Z')

const screen = (patch: Partial<Screen>): Screen => ({
  id: 1,
  device_id: 'device-0001',
  name: 'TV Teras',
  name_hint: '',
  user_agent: '',
  forced_mode: 'auto',
  forced_until: null,
  blocked_at: null,
  first_seen_at: '2026-11-01T08:00:00Z',
  updated_at: '2026-11-01T08:00:00Z',
  status: 'offline',
  last_seen_at: null,
  app_version: '',
  ...patch,
})

describe('screens utils', () => {
  it('names the device from its user agent', () => {
    expect(
      deviceLabel(
        'Mozilla/5.0 (Linux; Android 11; TV Box) AppleWebKit/537.36 Chrome/120.0 Safari/537.36',
      ),
    ).toBe('Chrome Android')
    expect(deviceLabel('Mozilla/5.0 (SMART-TV; Linux; Tizen 6.0) SamsungBrowser/4.0')).toBe(
      'Samsung Internet Tizen TV',
    )
    expect(deviceLabel('')).toBe('Browser')
  })

  it('ends a forced mode at forced_until', () => {
    expect(
      currentMode(screen({ forced_mode: 'dark', forced_until: '2026-11-02T17:00:00Z' }), now),
    ).toBe('dark')
    expect(
      currentMode(screen({ forced_mode: 'dark', forced_until: '2026-11-02T09:00:00Z' }), now),
    ).toBe('auto')
  })

  it('hides blocked screens and sets new ones apart', () => {
    const groups = groupScreens([
      screen({ id: 1 }),
      screen({ id: 2, name: '' }),
      screen({ id: 3, blocked_at: '2026-11-02T09:00:00Z' }),
    ])
    expect([groups.named, groups.fresh, groups.blocked].map((g) => g.map((s) => s.id))).toEqual([
      [1],
      [2],
      [3],
    ])
  })

  it('says when the screen was last seen', () => {
    expect(seenLine(screen({ status: 'online', app_version: '1.4.0' }), now)).toBe(
      'Aktif sekarang  ·  versi 1.4.0',
    )
    expect(seenLine(screen({ last_seen_at: '2026-11-02T08:00:00Z' }), now)).toBe('Aktif 2 jam lalu')
  })
})
