import { usePublicMosqueLogo, usePublicMosqueProfile } from '@/hooks/useMosqueProfile'
import { useServerNow } from '@/hooks/useServerClock'
import { useIqamahSettings, useTodaySchedule } from '@/hooks/useTvData'
import { timezoneLabel } from '@/utils/timezone'
import { wallClock } from './clock'
import { TvStage } from './components/TvStage'
import { displayMode } from './displayMode'
import { AdzanMode } from './modes/AdzanMode'
import { IqamahMode } from './modes/IqamahMode'
import { JumatMode } from './modes/JumatMode'
import { NormalMode } from './modes/NormalMode'
import { SholatMode } from './modes/SholatMode'
import type { TvContext } from './types'
import { useWakeLock } from './useWakeLock'

// TV display (E8), /tv. Public and read-only; switches between its modes on its own from the prayer times.
export function TvPage() {
  useWakeLock()
  // The server's clock, not the TV's: a TV is often unset or drifting.
  const now = useServerNow()
  const profile = usePublicMosqueProfile()
  const logo = usePublicMosqueLogo()
  const schedule = useTodaySchedule()
  const settings = useIqamahSettings()

  const timeZone = profile.data?.timezone ?? 'Asia/Jakarta'
  const tv: TvContext = {
    now,
    clock: wallClock(now, timeZone),
    timeZone,
    zone: timezoneLabel(timeZone),
    mosque: {
      name: profile.data?.name ?? 'Masjid Digital',
      address: profile.data?.address ?? '',
      logo: logo.data ?? null,
    },
  }

  if (!schedule.data || !settings.data) return <TvStage className="bg-brand">{null}</TvStage>
  const state = displayMode(tv.clock, schedule.data, settings.data)

  return (
    <TvStage>
      {state.mode === 'normal' && (
        <NormalMode tv={tv} state={state} schedule={schedule.data} settings={settings.data} />
      )}
      {state.mode === 'adzan' && <AdzanMode tv={tv} state={state} settings={settings.data} />}
      {state.mode === 'iqamah' && <IqamahMode tv={tv} state={state} />}
      {state.mode === 'sholat' && <SholatMode tv={tv} state={state} />}
      {state.mode === 'jumat' && <JumatMode tv={tv} state={state} />}
    </TvStage>
  )
}
