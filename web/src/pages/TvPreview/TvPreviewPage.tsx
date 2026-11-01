import { useState } from 'react'
import { wallClock } from '@/pages/Tv/clock'
import { TvScreen } from '@/pages/Tv/components/TvScreen'
import { TvStage } from '@/pages/Tv/components/TvStage'
import { displayMode } from '@/pages/Tv/displayMode'
import type { TvContext } from '@/pages/Tv/types'
import type { SholatDisplay } from '@/types/iqamah'
import { timezoneLabel } from '@/utils/timezone'
import { ControlPanel } from './components/ControlPanel'
import {
  MOCK_DUTY,
  MOCK_MOSQUE,
  MOCK_SCHEDULE,
  MOCK_SETTINGS,
  scenarioGroups,
  TIME_ZONE,
} from './scenarios'
import { useSimulatedClock } from './useSimulatedClock'

const clockFormat = new Intl.DateTimeFormat('id-ID', {
  timeZone: TIME_ZONE,
  weekday: 'long',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
  hourCycle: 'h23',
})

/*
 * Dev-only preview of the TV display (route /tv-preview): the real TV screen on mock data and a
 * simulated clock, with jumps to just before each mode switch and a faster clock to watch them.
 */
export function TvPreviewPage() {
  const [sholatDisplay, setSholatDisplay] = useState<SholatDisplay>(MOCK_SETTINGS.sholat_display)
  const [hasDuty, setHasDuty] = useState(true)
  const settings = { ...MOCK_SETTINGS, sholat_display: sholatDisplay }
  const groups = scenarioGroups(settings)
  const clock = useSimulatedClock(groups[0]?.scenarios[0]?.start ?? new Date())

  const tv: TvContext = {
    now: clock.now,
    clock: wallClock(clock.now, TIME_ZONE),
    timeZone: TIME_ZONE,
    zone: timezoneLabel(TIME_ZONE),
    mosque: MOCK_MOSQUE,
  }
  const duty = hasDuty ? MOCK_DUTY : undefined

  return (
    <>
      <TvStage>
        <TvScreen tv={tv} schedule={MOCK_SCHEDULE} settings={settings} duty={duty} />
      </TvStage>
      <ControlPanel
        state={displayMode(tv.clock, MOCK_SCHEDULE, settings)}
        clockLabel={`${clockFormat.format(clock.now)} ${tv.zone}`}
        groups={groups}
        speed={clock.speed}
        paused={clock.paused}
        sholatDisplay={sholatDisplay}
        hasDuty={hasDuty}
        onScenario={(scenario) => clock.jump(scenario.start)}
        onSpeed={clock.changeSpeed}
        onTogglePaused={clock.togglePaused}
        onSholatDisplay={setSholatDisplay}
        onHasDuty={setHasDuty}
      />
    </>
  )
}
