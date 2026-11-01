import { usePublicMosqueLogo, usePublicMosqueProfile } from '@/hooks/useMosqueProfile'
import { usePublicDaySchedule } from '@/hooks/usePrayerTimes'
import { useServerNow } from '@/hooks/useServerClock'
import { useIqamahSettings } from '@/hooks/useTvData'
import { localDate } from '@/utils/prayer/calendar'
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
  const settings = useIqamahSettings()

  const timeZone = profile.data?.timezone ?? 'Asia/Jakarta'
  // Computed on the TV for the mosque's date, so it rolls over at the mosque's midnight
  const schedule = usePublicDaySchedule(localDate(now, timeZone))
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

  if (schedule.data === null) {
    return (
      <TvStage className="bg-brand">
        <p className="text-on-brand-muted text-tv-36 m-auto max-w-[1100px] text-center font-semibold">
          Jadwal sholat tampil setelah lokasi masjid diisi di menu Profil masjid.
        </p>
      </TvStage>
    )
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
