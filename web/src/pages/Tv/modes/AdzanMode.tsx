import '@fontsource/amiri/arabic-700.css'
import type { IqamahSettings } from '@/types/iqamah'
import { formatClock, formatLongDate } from '../clock'
import { TopBar } from '../components/TopBar'
import type { DisplayState } from '../displayMode'
import type { TvContext } from '../types'

type AdzanModeProps = {
  tv: TvContext
  state: Extract<DisplayState, { mode: 'adzan' }>
  settings: IqamahSettings
}

// Adzan (E8-02, tv-adzan.html): "Telah masuk waktu {waktu}".
export function AdzanMode({ tv, state, settings }: AdzanModeProps) {
  const reminders = [
    'Dengarkan dan jawab adzan',
    state.jumat
      ? 'Khutbah segera dimulai'
      : `Iqamah dalam ${settings.iqamah_min[state.prayer.key]} menit`,
    'Matikan nada dering HP',
  ]

  return (
    <div className="bg-brand text-on-brand flex min-h-0 flex-1 flex-col px-16 py-12">
      <TopBar
        name={tv.mosque.name}
        right={`${formatLongDate(tv.now, tv.timeZone)}  ·  ${formatClock(tv.clock.seconds)}`}
      />
      <div className="flex flex-1 flex-col items-center justify-center gap-5 text-center">
        <div
          lang="ar"
          dir="rtl"
          className="font-arabic text-tv-120 text-gold leading-snug font-bold"
        >
          اللّٰهُ أَكْبَرُ
        </div>
        <div className="text-tv-36 text-on-brand-muted font-semibold tracking-widest">
          TELAH MASUK WAKTU
        </div>
        <div className="text-tv-200 leading-none font-extrabold tracking-tight uppercase">
          {state.jumat ? 'Jumat' : state.prayer.name}
        </div>
        <div className="text-tv-44 font-semibold">
          {state.prayer.time} {tv.zone}
        </div>
      </div>
      <div className="flex justify-center gap-6">
        {reminders.map((text) => (
          <div key={text} className="bg-brand-2 text-tv-26 rounded-2xl px-8 py-5 font-semibold">
            {text}
          </div>
        ))}
      </div>
    </div>
  )
}
