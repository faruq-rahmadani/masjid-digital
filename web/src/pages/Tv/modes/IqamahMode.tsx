import { useDutyRoster } from '@/hooks/useTvData'
import { formatClock, formatMs } from '../clock'
import { TopBar } from '../components/TopBar'
import type { DisplayState } from '../displayMode'
import type { TvContext } from '../types'

type IqamahModeProps = { tv: TvContext; state: Extract<DisplayState, { mode: 'iqamah' }> }

// Iqamah (E8-03, tv-iqamah.html): countdown with progress, reminders, imam and muadzin.
export function IqamahMode({ tv, state }: IqamahModeProps) {
  const duty = useDutyRoster().data
  const total = state.totalMinutes * 60
  const percent = Math.round((Math.min(state.secondsLeft, total) / total) * 100)
  const cards = [
    { title: 'Luruskan dan rapatkan shaf', note: 'Isi shaf terdepan lebih dulu' },
    { title: 'Matikan nada dering HP', note: 'Agar sholat tetap khusyuk' },
    ...(duty ? [{ title: `Imam: ${duty.imam}`, note: `Muadzin: ${duty.muadzin}` }] : []),
  ]

  return (
    <div className="bg-brand text-on-brand flex min-h-0 flex-1 flex-col px-16 py-12">
      <TopBar
        name={tv.mosque.name}
        right={`${state.prayer.name} ${state.prayer.time}  ·  Sekarang ${formatClock(tv.clock.seconds)}`}
      />
      <div className="flex flex-1 flex-col items-center justify-center gap-6 text-center">
        <div className="text-tv-40 text-on-brand-muted font-semibold tracking-widest uppercase">
          Menuju iqamah {state.prayer.name}
        </div>
        <div className="text-tv-340 text-gold leading-none font-extrabold tracking-tighter">
          {formatMs(state.secondsLeft)}
        </div>
        <div
          role="progressbar"
          aria-label="Sisa waktu iqamah"
          aria-valuenow={percent}
          aria-valuemin={0}
          aria-valuemax={100}
          className="bg-brand-2 h-4.5 w-250 overflow-hidden rounded-full"
        >
          <div className="bg-gold h-full rounded-full" style={{ width: `${percent}%` }} />
        </div>
        <div className="text-tv-26 text-on-brand-muted">dari {state.totalMinutes} menit</div>
      </div>
      <div className="grid grid-cols-3 gap-6">
        {cards.map((card) => (
          <div key={card.title} className="bg-canvas text-ink rounded-tv px-8 py-7">
            <div className="text-tv-30 font-extrabold">{card.title}</div>
            <div className="text-tv-22 text-ink-soft mt-2">{card.note}</div>
          </div>
        ))}
      </div>
    </div>
  )
}
