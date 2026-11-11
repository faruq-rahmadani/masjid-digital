import { useRef } from 'react'
import { useDocumentBranding } from '@/hooks/useDocumentBranding'
import { useUpcomingFridayDuties } from '@/hooks/useFridayDuties'
import { usePublicIqamahSettings } from '@/hooks/useIqamah'
import { usePublicMosqueLogo, usePublicMosqueProfile } from '@/hooks/useMosqueProfile'
import { usePublicDaySchedule } from '@/hooks/usePrayerTimes'
import { useServerNow } from '@/hooks/useServerClock'
import { localDate } from '@/utils/prayer/calendar'
import { timezoneLabel } from '@/utils/timezone'
import { wallClock } from './clock'
import { BlockedScreen, OfflineMark, SoundPrompt } from './components/ScreenNotices'
import { TvScreen } from './components/TvScreen'
import { TvStage } from './components/TvStage'
import { displayMode } from './displayMode'
import type { TvContext } from './types'
import { useAlarm } from './useAlarm'
import { useAlignedRefresh } from './useAlignedRefresh'
import { activeMode, useScreenConnection } from './useScreenConnection'
import { useWakeLock } from './useWakeLock'

// TV display (E8), /tv. Public and read-only; switches between its modes on its own from the prayer times.
export function TvPage() {
  useWakeLock()
  useDocumentBranding()
  // The server's clock, not the TV's: a TV is often unset or drifting.
  const now = useServerNow()
  useAlignedRefresh(now)
  // the alarm needs the connection's forced mode, the connection needs the alarm's test sound
  const testSound = useRef(() => {})
  const connection = useScreenConnection(() => testSound.current())
  const profile = usePublicMosqueProfile()
  const logo = usePublicMosqueLogo()
  const iqamah = usePublicIqamahSettings()
  const fridayDuties = useUpcomingFridayDuties()

  const timeZone = profile.data?.timezone || 'Asia/Jakarta'
  // Computed on the TV for the mosque's date, so it rolls over at the mosque's midnight
  const today = localDate(now, timeZone)
  const schedule = usePublicDaySchedule(today)
  const settings = iqamah.data
  const duty = fridayDuties.data?.find((d) => d.date === today)
  const forcedMode = activeMode(connection.state, now)
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

  // a forced mode overrides the schedule, and its alarm with it
  const scheduled =
    schedule.data && settings && forcedMode === 'auto'
      ? displayMode(tv.clock, schedule.data, settings)
      : null
  const alarm = useAlarm(settings, scheduled)
  testSound.current = alarm.testSound

  if (connection.blocked) return <BlockedScreen />

  if (schedule.data === null) {
    return (
      <TvStage className="bg-brand">
        <p className="text-on-brand-muted text-tv-36 m-auto max-w-[1100px] text-center font-semibold">
          Jadwal sholat tampil setelah lokasi masjid diisi di menu Profil masjid.
        </p>
      </TvStage>
    )
  }
  if (!schedule.data || !settings) return <TvStage className="bg-brand">{null}</TvStage>

  return (
    <TvStage className="relative">
      <TvScreen
        tv={tv}
        schedule={schedule.data}
        settings={settings}
        duty={duty}
        forcedMode={forcedMode}
      />
      {alarm.needsTap && <SoundPrompt />}
      {isLongOffline(connection.disconnectedAt) && <OfflineMark />}
    </TvStage>
  )
}

/** The PRD's offline rule: more than 2 minutes without the live connection. */
const OFFLINE_AFTER_MS = 2 * 60_000

function isLongOffline(disconnectedAt: number | null) {
  return disconnectedAt !== null && Date.now() - disconnectedAt >= OFFLINE_AFTER_MS
}
