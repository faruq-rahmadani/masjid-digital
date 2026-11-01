import type { FridayDuty } from '@/types/fridayDuty'
import { formatClock, formatHms, formatLongDate } from '../clock'
import { RunningText } from '../components/RunningText'
import { TopBar } from '../components/TopBar'
import type { DisplayState } from '../displayMode'
import type { TvContext } from '../types'

type JumatModeProps = {
  tv: TvContext
  state: Extract<DisplayState, { mode: 'jumat' }>
  /** Today's petugas Jumat; undefined when nobody filled them in. */
  duty: FridayDuty | undefined
}

// Jumat (E8-05, tv-jumat.html): before the adzan and during the khutbah.
export function JumatMode({ tv, state, duty }: JumatModeProps) {
  // only the petugas filled in; a Friday may have just the khatib
  const officers = [
    { label: 'KHATIB', name: duty?.khatib },
    { label: 'IMAM', name: duty?.imam },
    { label: 'MUADZIN', name: duty?.muadzin },
  ].filter((officer) => officer.name)

  return (
    <div className="bg-brand text-on-brand flex min-h-0 flex-1 flex-col gap-8 px-16 py-12">
      <TopBar
        name={tv.mosque.name}
        right={`${formatLongDate(tv.now, tv.timeZone)}  ·  ${formatClock(tv.clock.seconds)}`}
      />
      <div className="flex min-h-0 flex-1 gap-8">
        <div className="flex flex-1 flex-col justify-center gap-5">
          <div className="text-tv-36 text-gold font-semibold tracking-widest">SHOLAT JUMAT</div>
          <div className="text-tv-180 leading-none font-extrabold tracking-tighter">
            {state.prayer.time}
          </div>
          <div className="text-tv-34 text-on-brand-muted">
            {state.phase === 'before'
              ? `Adzan dalam ${formatHms(state.secondsLeft)}`
              : 'Khutbah sedang berlangsung'}
          </div>
          <div className="bg-canvas text-ink rounded-tv mt-6 max-w-225 px-9 py-7">
            <div className="text-tv-30 font-extrabold">Diam dan dengarkan khutbah</div>
            <div className="text-tv-24 text-ink-soft mt-2.5 leading-snug">
              Berbicara saat khatib berkhutbah membuat pahala Jumat sia-sia. (HR. Bukhari dan
              Muslim)
            </div>
          </div>
        </div>
        <div className="flex w-170 shrink-0 flex-col justify-center gap-5">
          {officers.map((officer) => (
            <div key={officer.label} className="bg-brand-2 rounded-tv px-9 py-8">
              <div className="text-tv-22 text-on-brand-muted tracking-wider">{officer.label}</div>
              <div className="text-tv-44 mt-1.5 font-extrabold">{officer.name}</div>
            </div>
          ))}
          {duty?.theme && (
            <div className="border-gold rounded-tv border-2 px-9 py-7">
              <div className="text-tv-22 text-gold tracking-wider">TEMA KHUTBAH</div>
              <div className="text-tv-32 mt-1.5 font-bold">{duty.theme}</div>
            </div>
          )}
        </div>
      </div>
      <RunningText />
    </div>
  )
}
