// Time on the TV is the mosque's wall clock (its profile time zone), not the device's.

const WEEKDAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

export type WallClock = {
  /** Seconds since local midnight. */
  seconds: number
  /** 0 = Sunday … 5 = Friday. */
  weekday: number
}

const partsFormatters = new Map<string, Intl.DateTimeFormat>()

export function wallClock(date: Date, timeZone: string): WallClock {
  let format = partsFormatters.get(timeZone)
  if (!format) {
    format = new Intl.DateTimeFormat('en-US', {
      timeZone,
      hourCycle: 'h23',
      weekday: 'short',
      hour: 'numeric',
      minute: 'numeric',
      second: 'numeric',
    })
    partsFormatters.set(timeZone, format)
  }
  const parts = Object.fromEntries(format.formatToParts(date).map((p) => [p.type, p.value]))
  return {
    seconds: Number(parts.hour) * 3600 + Number(parts.minute) * 60 + Number(parts.second),
    weekday: WEEKDAYS.indexOf(parts.weekday ?? ''),
  }
}

/** "18:15" → seconds since midnight. */
export function parseTime(time: string) {
  const [h = 0, m = 0] = time.split(':').map(Number)
  return h * 3600 + m * 60
}

const pad = (n: number) => String(n).padStart(2, '0')

/** 62_820 → "17:27" */
export function formatClock(seconds: number) {
  return `${pad(Math.floor(seconds / 3600) % 24)}:${pad(Math.floor(seconds / 60) % 60)}`
}

/** 2865 → "00:47:45" */
export function formatHms(seconds: number) {
  const s = Math.max(0, Math.ceil(seconds))
  return `${pad(Math.floor(s / 3600))}:${pad(Math.floor(s / 60) % 60)}:${pad(s % 60)}`
}

/** 372 → "06:12" */
export function formatMs(seconds: number) {
  const s = Math.max(0, Math.ceil(seconds))
  return `${pad(Math.floor(s / 60))}:${pad(s % 60)}`
}

/** "Rabu, 28 Oktober 2026" */
export function formatLongDate(date: Date, timeZone: string) {
  return new Intl.DateTimeFormat('id-ID', {
    timeZone,
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(date)
}

/** "17 Jumadilawal 1448 H" (Umm al-Qura; the Hijri correction setting comes with E4-05). */
export function formatHijriDate(date: Date, timeZone: string) {
  return new Intl.DateTimeFormat('id-ID-u-ca-islamic-umalqura', {
    timeZone,
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(date)
}

/** "28 Okt 2026" for a "YYYY-MM-DD" date. */
export function formatShortDate(isoDate: string) {
  const [y = 0, m = 1, d = 1] = isoDate.split('-').map(Number)
  return new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
    .format(new Date(y, m - 1, d))
    .replace('.', '')
}
