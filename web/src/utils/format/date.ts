const shortDate = new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short' })

/** "18:05" (id-ID itself would write "18.05"). */
export function formatTime(date: Date) {
  return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

const startOfDay = (date: Date) => new Date(date.getFullYear(), date.getMonth(), date.getDate())

/** Activity feed style: today "16:02", yesterday "Kemarin", earlier "25 Okt". */
export function formatActivityTime(iso: string, now = new Date()) {
  const date = new Date(iso)
  const days = Math.round((startOfDay(now).getTime() - startOfDay(date).getTime()) / 86_400_000)
  if (days <= 0) return formatTime(date)
  if (days === 1) return 'Kemarin'
  return shortDate.format(date)
}

/** "baru saja", "5 menit lalu", "2 jam lalu", "3 hari lalu". */
export function formatTimeAgo(iso: string, now = new Date()) {
  const minutes = Math.floor((now.getTime() - new Date(iso).getTime()) / 60_000)
  if (minutes < 1) return 'baru saja'
  if (minutes < 60) return `${minutes} menit lalu`
  if (minutes < 24 * 60) return `${Math.floor(minutes / 60)} jam lalu`
  return `${Math.floor(minutes / (24 * 60))} hari lalu`
}
