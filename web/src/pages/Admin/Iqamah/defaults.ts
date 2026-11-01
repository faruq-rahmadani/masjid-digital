import type { IqamahSettings } from '@/types/iqamah'

/** Mirrors domain.DefaultIqamahSettings in Go; "Atur ulang" puts the table back to these. */
export const DEFAULT_IQAMAH_SETTINGS: IqamahSettings = {
  iqamah_min: { subuh: 15, dzuhur: 10, ashar: 10, maghrib: 7, isya: 10 },
  sholat_min: { subuh: 15, dzuhur: 12, ashar: 12, maghrib: 12, isya: 15, jumat: 40 },
  adzan_display_min: 4,
  sholat_display: 'dim',
  jumat_before_min: 60,
  khutbah_min: 30,
  alarm: { enabled: true, at_adzan: true, before_iqamah: true, sound: 'beep', volume: 70 },
}
