import { type ReactNode, useEffect, useState } from 'react'
import { cn } from '@/utils/cn'

export const STAGE = { width: 1920, height: 1080 }

function fitScale() {
  return Math.min(window.innerWidth / STAGE.width, window.innerHeight / STAGE.height)
}

/*
 * The TV is designed at 1920x1080 px. The stage keeps that layout and scales it to fit any
 * screen (720p TV, 4K, a laptop), letterboxing the rest. A tap (or a remote's OK) goes
 * fullscreen; a TV remote can't double-click, and Esc or Back leaves.
 */
export function TvStage({ className, children }: { className?: string; children: ReactNode }) {
  const [scale, setScale] = useState(fitScale)

  useEffect(() => {
    const onResize = () => setScale(fitScale())
    const enterFullscreen = (e: MouseEvent) => {
      // controls drawn over the stage (the dev preview's panel) keep their clicks
      if (document.fullscreenElement || (e.target as Element).closest?.('[data-tv-controls]'))
        return
      document.documentElement.requestFullscreen?.().catch(() => {})
    }
    window.addEventListener('resize', onResize)
    window.addEventListener('click', enterFullscreen)
    return () => {
      window.removeEventListener('resize', onResize)
      window.removeEventListener('click', enterFullscreen)
    }
  }, [])

  return (
    <div className="fixed inset-0 flex cursor-none items-center justify-center overflow-hidden bg-black select-none">
      <div
        // overflow-hidden + each mode's min-h-0: content can never grow the stage past 1080 px
        // (it would shift the whole screen whenever, say, a longer quote rotates in).
        className={cn('flex shrink-0 flex-col overflow-hidden', className)}
        style={{ width: STAGE.width, height: STAGE.height, transform: `scale(${scale})` }}
      >
        {children}
      </div>
    </div>
  )
}
