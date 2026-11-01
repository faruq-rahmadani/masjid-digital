import { describe, expect, it } from 'vitest'
import { wallClock } from '@/pages/Tv/clock'
import { displayMode } from '@/pages/Tv/displayMode'
import { MOCK_SCHEDULE, MOCK_SETTINGS, scenarioGroups, TIME_ZONE } from './scenarios'

const modeAt = (date: Date, afterSeconds: number) => {
  const state = displayMode(
    wallClock(new Date(date.getTime() + afterSeconds * 1000), TIME_ZONE),
    MOCK_SCHEDULE,
    MOCK_SETTINGS,
  )
  return state.mode === 'jumat' ? `jumat:${state.phase}` : state.mode
}

describe('scenarioGroups', () => {
  // Each jump lands a few seconds before a switch: [mode at the jump, mode 15 s later]
  const expected: Record<string, [string, string]> = {
    normal: ['normal', 'normal'],
    'before-adzan': ['normal', 'adzan'],
    adzan: ['adzan', 'adzan'],
    'adzan-end': ['adzan', 'iqamah'],
    'iqamah-1m': ['iqamah', 'iqamah'],
    'iqamah-10s': ['iqamah', 'sholat'],
    'sholat-end': ['sholat', 'normal'],
    'jumat-before': ['normal', 'jumat:before'],
    'jumat-1m': ['jumat:before', 'jumat:before'],
    'jumat-khutbah': ['adzan', 'jumat:khutbah'],
    'jumat-sholat': ['jumat:khutbah', 'sholat'],
    'jumat-end': ['sholat', 'normal'],
  }

  it('jumps to just before each mode switch', () => {
    const scenarios = scenarioGroups(MOCK_SETTINGS).flatMap((group) => group.scenarios)
    expect(
      Object.fromEntries(scenarios.map((s) => [s.id, [modeAt(s.start, 0), modeAt(s.start, 15)]])),
    ).toEqual(expected)
  })
})
