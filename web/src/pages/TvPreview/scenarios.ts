import { DEFAULT_IQAMAH_SETTINGS } from '@/pages/Admin/Iqamah/defaults'
import { parseTime } from '@/pages/Tv/clock'
import type { FridayDuty } from '@/types/fridayDuty'
import type { IqamahSettings } from '@/types/iqamah'
import type { ScheduleItem } from '@/types/tv'

// Mock data for the TV preview: one fixed day of prayer times in WIB, the default iqamah
// settings, and a Wednesday and a Friday to run them on.

export const TIME_ZONE = 'Asia/Jakarta'
const UTC_OFFSET = '+07:00'

const WEDNESDAY = '2026-11-04'
const FRIDAY = '2026-11-06'

export const MOCK_SCHEDULE: ScheduleItem[] = [
  { key: 'imsak', name: 'Imsak', time: '04:01' },
  { key: 'subuh', name: 'Subuh', time: '04:11' },
  { key: 'terbit', name: 'Terbit', time: '05:28' },
  { key: 'dzuhur', name: 'Dzuhur', time: '11:41' },
  { key: 'ashar', name: 'Ashar', time: '15:01' },
  { key: 'maghrib', name: 'Maghrib', time: '17:53' },
  { key: 'isya', name: 'Isya', time: '19:05' },
]

export const MOCK_SETTINGS: IqamahSettings = DEFAULT_IQAMAH_SETTINGS

export const MOCK_DUTY: FridayDuty = {
  date: FRIDAY,
  khatib: 'Ustadz Abdullah',
  imam: 'Ustadz Hasan',
  muadzin: 'Bilal Ahmad',
  theme: 'Menjaga amanah',
  updated_by: 1,
  created_at: `${FRIDAY}T00:00:00Z`,
  updated_at: `${FRIDAY}T00:00:00Z`,
}

export const MOCK_MOSQUE = {
  name: 'Masjid Al-Ikhlas',
  address: 'Jl. Melati No. 12, Jakarta Selatan',
  logo: null,
}

/** The moment `seconds` after the mosque's midnight on `date`. */
function at(date: string, seconds: number) {
  return new Date(new Date(`${date}T00:00:00${UTC_OFFSET}`).getTime() + seconds * 1000)
}

const timeOf = (key: ScheduleItem['key']) =>
  parseTime(MOCK_SCHEDULE.find((item) => item.key === key)?.time ?? '00:00')

export type Scenario = { id: string; label: string; start: Date }
export type ScenarioGroup = { label: string; scenarios: Scenario[] }

/** Jump points a few seconds before each switch, so every transition can be watched. */
export function scenarioGroups(settings: IqamahSettings): ScenarioGroup[] {
  const min = 60
  const maghrib = timeOf('maghrib')
  const maghribAdzanEnd = maghrib + settings.adzan_display_min * min
  const maghribIqamah = maghrib + settings.iqamah_min.maghrib * min
  const maghribDone = maghribIqamah + settings.sholat_min.maghrib * min

  const dzuhur = timeOf('dzuhur')
  const jumatStart = dzuhur - settings.jumat_before_min * min
  const jumatAdzanEnd = dzuhur + settings.adzan_display_min * min
  const khutbahEnd = jumatAdzanEnd + settings.khutbah_min * min
  const jumatDone = khutbahEnd + settings.sholat_min.jumat * min

  return [
    {
      label: 'Hari biasa · Maghrib',
      scenarios: [
        { id: 'normal', label: 'Normal', start: at(WEDNESDAY, maghrib - 10 * min) },
        { id: 'before-adzan', label: '10 dtk sebelum adzan', start: at(WEDNESDAY, maghrib - 10) },
        { id: 'adzan', label: 'Adzan', start: at(WEDNESDAY, maghrib) },
        { id: 'adzan-end', label: 'Adzan → iqamah', start: at(WEDNESDAY, maghribAdzanEnd - 5) },
        {
          id: 'iqamah-1m',
          label: '1 mnt sebelum iqamah',
          start: at(WEDNESDAY, maghribIqamah - min),
        },
        {
          id: 'iqamah-10s',
          label: '10 dtk sebelum iqamah',
          start: at(WEDNESDAY, maghribIqamah - 10),
        },
        { id: 'sholat-end', label: 'Sholat → normal', start: at(WEDNESDAY, maghribDone - 10) },
      ],
    },
    {
      label: 'Jumat',
      scenarios: [
        { id: 'jumat-before', label: 'Normal → layar Jumat', start: at(FRIDAY, jumatStart - 10) },
        { id: 'jumat-1m', label: '1 mnt sebelum adzan', start: at(FRIDAY, dzuhur - min) },
        { id: 'jumat-khutbah', label: 'Adzan → khutbah', start: at(FRIDAY, jumatAdzanEnd - 5) },
        { id: 'jumat-sholat', label: 'Khutbah → sholat', start: at(FRIDAY, khutbahEnd - 10) },
        { id: 'jumat-end', label: 'Sholat → normal', start: at(FRIDAY, jumatDone - 10) },
      ],
    },
  ]
}
