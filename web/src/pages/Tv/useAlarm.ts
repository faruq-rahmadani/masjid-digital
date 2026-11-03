import { useEffect, useMemo, useRef, useState } from 'react'
import type { IqamahSettings } from '@/types/iqamah'
import { shouldSound } from './alarm/cue'
import { createAlarmPlayer } from './alarm/sound'
import type { DisplayState } from './displayMode'

/**
 * The alarm (E8-06): sounds on the adzan screen and in the last 10 seconds before iqamah, per the
 * Iqamah & Alarm settings. Audio stays locked until the first tap (or remote OK press) after the
 * TV boots; `needsTap` asks for it while the alarm is on.
 */
export function useAlarm(settings: IqamahSettings | undefined, state: DisplayState | null) {
  const player = useMemo(createAlarmPlayer, [])
  const [locked, setLocked] = useState(() => player.locked)
  const previous = useRef<DisplayState | null>(null)

  useEffect(() => {
    const unlock = () => player.unlock()?.then(() => setLocked(player.locked))
    window.addEventListener('click', unlock)
    window.addEventListener('keydown', unlock)
    return () => {
      window.removeEventListener('click', unlock)
      window.removeEventListener('keydown', unlock)
    }
  }, [player])

  useEffect(() => {
    if (!settings || !state) return
    if (shouldSound(previous.current, state, settings.alarm)) {
      player.play(settings.alarm.sound, settings.alarm.volume)
    }
    previous.current = state
  }, [player, settings, state])

  return {
    needsTap: Boolean(settings?.alarm.enabled) && player.supported && locked,
    /** "Uji bunyi" from the admin: plays the configured sound once, alarm on or off. */
    testSound: () => settings && player.play(settings.alarm.sound, settings.alarm.volume),
  }
}
