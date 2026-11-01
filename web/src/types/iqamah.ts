// Iqamah, mode sholat and alarm (E5): what the TV does around each prayer.

import type { PrayerKey } from './prayer'

export type SholatDisplay = 'dim' | 'black'

export type AlarmSound = 'beep' | 'bell'

export type AlarmSettings = {
  enabled: boolean
  at_adzan: boolean
  /** The last 10 seconds before iqamah. */
  before_iqamah: boolean
  sound: AlarmSound
  volume: number /** 0..100 */
}

export type IqamahSettings = {
  /** Minutes from the adzan time to iqamah, 1..30. */
  iqamah_min: Record<PrayerKey, number>
  /** How long mode sholat lasts, 5..30 (Jumat 5..60). */
  sholat_min: Record<PrayerKey | 'jumat', number>
  /** How long the "Telah masuk waktu" screen stays up, 1..10. */
  adzan_display_min: number
  sholat_display: SholatDisplay
  /** The Jumat screen shows this long before Dzuhur on Friday, 0..120 (0 = off). */
  jumat_before_min: number
  khutbah_min: number
  alarm: AlarmSettings
}
