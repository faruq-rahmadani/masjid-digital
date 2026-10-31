type PrayerTime = { name: string; time: string }

type NextPrayer = { name: string; time: string; minutesLeft: number }

/** The next prayer after `now`; after Isya it is tomorrow's Subuh. Times are "HH:MM" local. */
export function nextPrayer(now: Date, times: readonly PrayerTime[]): NextPrayer {
  const minutesNow = now.getHours() * 60 + now.getMinutes() + now.getSeconds() / 60
  const toMinutes = (time: string) => {
    const [h = 0, m = 0] = time.split(':').map(Number)
    return h * 60 + m
  }

  const upcoming = times.find((t) => toMinutes(t.time) > minutesNow)
  const next = upcoming ?? times[0]
  if (!next) throw new Error('nextPrayer needs at least one time')
  const target = toMinutes(next.time) + (upcoming ? 0 : 24 * 60)
  return { name: next.name, time: next.time, minutesLeft: Math.ceil(target - minutesNow) }
}

/** 47 → "00:47", 125 → "02:05" (hours:minutes, as on the mockup). */
export function formatCountdown(minutes: number) {
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`
}
