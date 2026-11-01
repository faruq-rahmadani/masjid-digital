import { formatClock } from '../clock'
import type { DisplayState } from '../displayMode'
import type { TvContext } from '../types'

type SholatModeProps = { tv: TvContext; state: Extract<DisplayState, { mode: 'sholat' }> }

// Sholat (E8-04, tv-sholat.html): dim text and a small clock, so the screen doesn't distract.
export function SholatMode({ tv, state }: SholatModeProps) {
  return (
    <div className="bg-tv-dark text-tv-dark-ink flex min-h-0 flex-1 flex-col items-center justify-center gap-4.5 text-center">
      <div className="text-tv-40 font-semibold tracking-wider">
        Sholat {state.jumat ? 'Jumat' : state.prayer.name} sedang berlangsung
      </div>
      <div className="text-tv-28">Mohon tenang dan jaga kekhusyukan</div>
      <div className="text-tv-64 mt-6 font-bold">{formatClock(tv.clock.seconds)}</div>
    </div>
  )
}
