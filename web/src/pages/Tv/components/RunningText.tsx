import { useRunningTexts } from '@/hooks/useTvData'

const SEPARATOR = '   •   '

// INFO bar. The text is laid out twice and slid by half its width, so the loop has no seam.
export function RunningText() {
  const texts = useRunningTexts().data ?? []
  if (texts.length === 0) return null
  const line = texts.join(SEPARATOR) + SEPARATOR
  // About 8 characters a second reads comfortably from the back of the room.
  const duration = `${Math.max(20, line.length / 8)}s`

  return (
    <div className="bg-gold text-on-gold flex h-13 shrink-0 items-center gap-5 overflow-hidden rounded-xl px-2.5">
      <div className="bg-on-gold text-on-brand text-tv-16 shrink-0 rounded-md px-4 py-2 font-extrabold tracking-widest">
        INFO
      </div>
      <div className="min-w-0 flex-1 overflow-hidden">
        <div
          className="animate-marquee text-tv-22 flex w-max font-semibold whitespace-pre"
          style={{ animationDuration: duration }}
        >
          <span>{line}</span>
          <span aria-hidden>{line}</span>
        </div>
      </div>
    </div>
  )
}
