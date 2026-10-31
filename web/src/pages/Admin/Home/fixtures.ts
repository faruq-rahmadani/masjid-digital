/*
 * Sample data for parts of the Home screen whose API doesn't exist yet. Replace each with its feature
 * when built: prayer engine (E4-01), screens (E9-02), activity log (E2-05).
 */

/** Today's times for Jakarta from the mockup, "HH:MM" local time. */
export const prayerTimes = [
  { name: 'Subuh', time: '04:42' },
  { name: 'Dzuhur', time: '12:09' },
  { name: 'Ashar', time: '15:24' },
  { name: 'Maghrib', time: '18:15' },
  { name: 'Isya', time: '19:25' },
] as const

export const screens = { online: 1, offline: 1, offlineSince: '2 jam lalu' }

export const recentActivity = [
  { id: 1, actor: 'Siti Aminah', action: 'menambah pemasukan Rp 750.000', time: '16:02' },
  { id: 2, actor: 'Hasan Basri', action: 'mengubah iqamah Maghrib jadi 7 mnt', time: '14:20' },
  { id: 3, actor: 'Hasan Basri', action: 'mengaktifkan slide Kajian Ahad', time: 'Kemarin' },
  { id: 4, actor: 'Ahmad Fauzi', action: 'menambah pengguna Bendahara', time: '25 Okt' },
]
