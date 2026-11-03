import type { FridayDuty } from '@/types/fridayDuty'
import type { IqamahSettings } from '@/types/iqamah'
import type { ScreenMode } from '@/types/screen'
import type { ScheduleItem } from '@/types/tv'
import { displayMode, forcedDisplayMode } from '../displayMode'
import { AdzanMode } from '../modes/AdzanMode'
import { IqamahMode } from '../modes/IqamahMode'
import { JumatMode } from '../modes/JumatMode'
import { NormalMode } from '../modes/NormalMode'
import { SholatMode } from '../modes/SholatMode'
import type { TvContext } from '../types'

type TvScreenProps = {
  tv: TvContext
  schedule: readonly ScheduleItem[]
  settings: IqamahSettings
  duty: FridayDuty | undefined
  /** A mode forced from the Layar TV page; `auto` follows the schedule. */
  forcedMode?: ScreenMode
}

// The mode the TV is in right now, from the clock, the day's schedule and the iqamah settings,
// unless an admin forced one.
// Shared by the TV (/tv) and its dev preview (/tv-preview).
export function TvScreen({ tv, schedule, settings, duty, forcedMode = 'auto' }: TvScreenProps) {
  if (forcedMode === 'dark') {
    return <div role="img" aria-label="Layar gelap" className="flex-1 bg-black" />
  }
  const state =
    forcedDisplayMode(forcedMode, tv.clock, schedule) ?? displayMode(tv.clock, schedule, settings)

  return (
    <>
      {state.mode === 'normal' && (
        <NormalMode tv={tv} state={state} schedule={schedule} settings={settings} />
      )}
      {state.mode === 'adzan' && <AdzanMode tv={tv} state={state} settings={settings} />}
      {state.mode === 'iqamah' && <IqamahMode tv={tv} state={state} />}
      {state.mode === 'sholat' && (
        <SholatMode tv={tv} state={state} display={settings.sholat_display} />
      )}
      {state.mode === 'jumat' && <JumatMode tv={tv} state={state} duty={duty} />}
    </>
  )
}
