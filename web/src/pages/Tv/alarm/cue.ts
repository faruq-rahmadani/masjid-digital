import type { AlarmSettings } from '@/types/iqamah'
import type { DisplayState } from '../displayMode'

/** The last seconds of the iqamah countdown that sound the "before iqamah" alarm. */
export const BEFORE_IQAMAH_SECONDS = 10

const isLastSeconds = (state: DisplayState | null) =>
  state?.mode === 'iqamah' && state.secondsLeft <= BEFORE_IQAMAH_SECONDS

/**
 * Whether the TV sounds the alarm as it goes from `prev` to `next`: on entering the adzan screen,
 * and on entering the last 10 seconds before iqamah. Nothing on the first state (a TV that
 * reloads mid-adzan stays quiet).
 */
export function shouldSound(prev: DisplayState | null, next: DisplayState, alarm: AlarmSettings) {
  if (!alarm.enabled || prev === null) return false
  const enteredAdzan = alarm.at_adzan && next.mode === 'adzan' && prev.mode !== 'adzan'
  const enteredLastSeconds = alarm.before_iqamah && isLastSeconds(next) && !isLastSeconds(prev)
  return enteredAdzan || enteredLastSeconds
}
