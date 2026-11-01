// Dates are the mosque's "YYYY-MM-DD"; formatting them at UTC midnight keeps the device's zone out.
const atUtc = (date: string) => new Date(`${date}T00:00:00Z`)
const shortDay = new Intl.DateTimeFormat('id-ID', {
  weekday: 'short',
  day: 'numeric',
  timeZone: 'UTC',
})
const longDay = new Intl.DateTimeFormat('id-ID', {
  weekday: 'long',
  day: 'numeric',
  month: 'short',
  timeZone: 'UTC',
})

/** "2026-11-05" → "Kam 5" */
export function formatShortDay(date: string) {
  return shortDay.format(atUtc(date)).replace(',', '')
}

/** "2026-11-06" → "Jumat, 6 Nov" */
export function formatLongDay(date: string) {
  return longDay.format(atUtc(date))
}
