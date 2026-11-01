import { useUpcomingFridayDuties } from '@/hooks/useFridayDuties'
import { usePublicIqamahSettings } from '@/hooks/useIqamah'
import { usePublicMosqueLogo, usePublicMosqueProfile } from '@/hooks/useMosqueProfile'
import { usePublicDaySchedule } from '@/hooks/usePrayerTimes'
import { useServerNow } from '@/hooks/useServerClock'
import { localDate } from '@/utils/prayer/calendar'
import { timezoneLabel } from '@/utils/timezone'
import { wallClock } from './clock'
import { TvScreen } from './components/TvScreen'
import { TvStage } from './components/TvStage'
import type { TvContext } from './types'
import { useWakeLock } from './useWakeLock'

// TV display (E8), /tv. Public and read-only; switches between its modes on its own from the prayer times.
export function TvPage() {
  useWakeLock()
  // The server's clock, not the TV's: a TV is often unset or drifting.
  const now = useServerNow()
  const profile = usePublicMosqueProfile()
  const logo = usePublicMosqueLogo()
  const iqamah = usePublicIqamahSettings()
  const fridayDuties = useUpcomingFridayDuties()

  const timeZone = profile.data?.timezone ?? 'Asia/Jakarta'
  // Computed on the TV for the mosque's date, so it rolls over at the mosque's midnight
  const today = localDate(now, timeZone)
  const schedule = usePublicDaySchedule(today)
  const settings = iqamah.data
  const duty = fridayDuties.data?.find((d) => d.date === today)
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
  if (!schedule.data || !settings) return <TvStage className="bg-brand">{null}</TvStage>

  return (
    <TvStage>
      <TvScreen tv={tv} schedule={schedule.data} settings={settings} duty={duty} />
    </TvStage>
  )
}
