import type { SholatDisplay } from '@/types/iqamah'
import { formatClock } from '../clock'
import type { DisplayState } from '../displayMode'
import type { TvContext } from '../types'

type SholatModeProps = {
  tv: TvContext
  state: Extract<DisplayState, { mode: 'sholat' }>
  display: SholatDisplay
}

// Sholat (E8-04, tv-sholat.html): dim text and a small clock, or a fully black screen, so the
// screen doesn't distract.
export function SholatMode({ tv, state, display }: SholatModeProps) {
  const label = `Sholat ${state.jumat ? 'Jumat' : state.prayer.name} sedang berlangsung`
  if (display === 'black') {
    return <div role="img" aria-label={label} className="flex-1 bg-black" />
  }

  return (
    <div className="bg-tv-dark text-tv-dark-ink flex min-h-0 flex-1 flex-col items-center justify-center gap-4.5 text-center">
      <div className="text-tv-40 font-semibold tracking-wider">{label}</div>
      <div className="text-tv-28">Mohon tenang dan jaga kekhusyukan</div>
      <div className="text-tv-64 mt-6 font-bold">{formatClock(tv.clock.seconds)}</div>
    </div>
  )
}
