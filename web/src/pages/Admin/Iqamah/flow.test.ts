import { describe, expect, it } from 'vitest'
import { DEFAULT_IQAMAH_SETTINGS as settings } from './defaults'
import { prayerFlow } from './flow'

const maghrib = { key: 'maghrib', name: 'Maghrib', time: '17:52' } as const
const dzuhur = { key: 'dzuhur', name: 'Dzuhur', time: '11:58' } as const

describe('prayerFlow', () => {
  it('runs adzan, the iqamah countdown, then mode sholat', () => {
    expect(prayerFlow(maghrib, settings, false).steps).toEqual([
      { label: 'Layar adzan', from: '17:52', to: '17:56' },
      { label: 'Hitung mundur iqamah', from: '17:56', to: '17:59' },
      { label: 'Mode sholat', from: '17:59', to: '18:11' },
    ])
  })

  it('skips the countdown when the adzan screen outlasts the iqamah wait', () => {
    const quick = { ...settings, iqamah_min: { ...settings.iqamah_min, maghrib: 3 } }
    expect(prayerFlow(maghrib, quick, false).steps.map((s) => s.label)).toEqual([
      'Layar adzan',
      'Mode sholat',
    ])
    expect(prayerFlow(maghrib, quick, false).steps[1]).toEqual({
      label: 'Mode sholat',
      from: '17:56',
      to: '18:07',
    })
  })

  it('turns Friday Dzuhur into Jumat', () => {
    expect(prayerFlow(dzuhur, settings, true)).toEqual({
      name: 'Jumat',
      time: '11:58',
      jumat: true,
      steps: [
        { label: 'Layar Jumat', from: '10:58', to: '11:58' },
        { label: 'Layar adzan', from: '11:58', to: '12:02' },
        { label: 'Khutbah', from: '12:02', to: '12:32' },
        { label: 'Mode sholat', from: '12:32', to: '13:12' },
      ],
    })
    const noScreen = { ...settings, jumat_before_min: 0 }
    expect(prayerFlow(dzuhur, noScreen, true).steps[0]?.label).toBe('Layar adzan')
    expect(prayerFlow(dzuhur, settings, false).name).toBe('Dzuhur')
  })
})
