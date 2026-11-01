import type { WallClock } from './clock'

/** What every mode needs to know about the mosque and the time. */
export type TvContext = {
  now: Date
  clock: WallClock
  timeZone: string
  /** "WIB" */
  zone: string
  mosque: { name: string; address: string; logo: string | null }
}
