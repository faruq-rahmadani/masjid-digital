import { describe, expect, it } from 'vitest'
import type { AlarmSettings } from '@/types/iqamah'
import type { DisplayState, Prayer } from '../displayMode'
import { shouldSound } from './cue'

const maghrib: Prayer = { key: 'maghrib', name: 'Maghrib', time: '17:53' }
const alarm: AlarmSettings = {
  enabled: true,
  at_adzan: true,
  before_iqamah: true,
  sound: 'beep',
  volume: 70,
}

const normal: DisplayState = { mode: 'normal', next: maghrib, secondsLeft: 5, jumat: false }
const adzan: DisplayState = { mode: 'adzan', prayer: maghrib, jumat: false }
const iqamah = (secondsLeft: number): DisplayState => ({
  mode: 'iqamah',
  prayer: maghrib,
  secondsLeft,
  totalMinutes: 7,
})

describe('shouldSound', () => {
  it('sounds once on entering the adzan screen', () => {
    expect(shouldSound(normal, adzan, alarm)).toBe(true)
    expect(shouldSound(adzan, adzan, alarm)).toBe(false)
  })

  it('sounds once on entering the last 10 seconds before iqamah', () => {
    expect(shouldSound(iqamah(11), iqamah(10), alarm)).toBe(true)
    expect(shouldSound(iqamah(10), iqamah(9), alarm)).toBe(false)
    expect(shouldSound(iqamah(60), iqamah(59), alarm)).toBe(false)
  })

  it('stays quiet when off, per cue, and on the first state', () => {
    expect(shouldSound(normal, adzan, { ...alarm, enabled: false })).toBe(false)
    expect(shouldSound(normal, adzan, { ...alarm, at_adzan: false })).toBe(false)
    expect(shouldSound(iqamah(11), iqamah(10), { ...alarm, before_iqamah: false })).toBe(false)
    expect(shouldSound(null, adzan, alarm)).toBe(false)
  })
})
