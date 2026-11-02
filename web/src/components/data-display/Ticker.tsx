// ۞ is in Amiri's Arabic subset; unicode-range keeps the font file from loading until it shows.
import '@fontsource/amiri/arabic-700.css'
import { useLayoutEffect, useRef, useState } from 'react'
import type { ScrollSpeed } from '@/types/runningText'
import { cn } from '@/utils/cn'

const SEPARATOR = '   •   '

/** Characters scrolled per second. Medium reads comfortably from the back of the room. */
const CHARS_PER_SECOND: Record<ScrollSpeed, number> = { slow: 5, medium: 8, fast: 12 }

/** Seconds for one loop of `line` (all texts, then the loop mark). */
export function scrollSeconds(line: string, speed: ScrollSpeed) {
  return (line.length + SEPARATOR.length) / CHARS_PER_SECOND[speed]
}

const sizes = {
  tv: {
    bar: 'h-13 gap-5 rounded-xl px-2.5',
    badge: 'text-tv-16 rounded-md px-4 py-2 tracking-widest',
    text: 'text-tv-22',
  },
  preview: {
    bar: 'h-8 gap-2 rounded-lg px-1.5',
    badge: 'rounded-sm px-1.5 py-0.5 text-[10px]',
    text: 'text-xs',
  },
} as const

type TickerProps = {
  texts: readonly string[]
  speed: ScrollSpeed
  /** `tv` on the 1920×1080 stage, `preview` in the admin's Running Text screen. */
  size?: keyof typeof sizes
  className?: string
}

/*
 * The INFO bar: one unbroken stream. Each copy is every text joined by • and closed by ۞, which
 * marks the return to the first text. Enough copies are laid out to cover the bar plus one, and
 * the strip slides by exactly one copy, so the loop has no seam or gap however short the texts.
 */
export function Ticker({ texts, speed, size = 'tv', className }: TickerProps) {
  const viewportRef = useRef<HTMLDivElement>(null)
  const copyRef = useRef<HTMLDivElement>(null)
  // Unmeasured (first paint, jsdom): two copies sliding by half, which is one copy.
  const [layout, setLayout] = useState({ copies: 2, shift: '-50%' })

  // Re-measure when the bar resizes (TV stage scaling, phone rotation) or the text changes.
  const line = texts.join(SEPARATOR)
  useLayoutEffect(() => {
    const viewport = viewportRef.current
    const copy = copyRef.current
    if (!line || !viewport || !copy || typeof ResizeObserver === 'undefined') return
    const measure = () => {
      const copyWidth = copy.offsetWidth
      if (copyWidth === 0) return
      setLayout({
        copies: Math.max(2, Math.ceil(viewport.offsetWidth / copyWidth) + 1),
        shift: `-${copyWidth}px`,
      })
    }
    measure()
    const observer = new ResizeObserver(measure)
    observer.observe(viewport)
    observer.observe(copy)
    return () => observer.disconnect()
  }, [line])

  if (texts.length === 0) return null
  const s = sizes[size]

  return (
    <div
      className={cn(
        'bg-gold text-on-gold flex shrink-0 items-center overflow-hidden',
        s.bar,
        className,
      )}
    >
      <div className={cn('bg-on-gold text-on-brand shrink-0 font-extrabold', s.badge)}>INFO</div>
      <div ref={viewportRef} className="min-w-0 flex-1 overflow-hidden">
        <div
          className={cn('animate-marquee flex w-max font-semibold whitespace-pre', s.text)}
          style={
            {
              animationDuration: `${scrollSeconds(line, speed)}s`,
              '--marquee-shift': layout.shift,
            } as React.CSSProperties
          }
        >
          {Array.from({ length: layout.copies }, (_, i) => (
            <div
              // biome-ignore lint/suspicious/noArrayIndexKey: identical copies, only their count changes
              key={i}
              ref={i === 0 ? copyRef : undefined}
              aria-hidden={i > 0 || undefined}
              className="flex shrink-0 items-center"
            >
              <span>{line}</span>
              <span aria-hidden className="font-arabic px-[0.9em] leading-none font-bold">
                ۞
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
