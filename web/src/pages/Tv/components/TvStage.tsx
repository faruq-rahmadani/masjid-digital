import { type ReactNode, useEffect, useState } from 'react'
import { cn } from '@/utils/cn'

export const STAGE = { width: 1920, height: 1080 }

function fitScale() {
  return Math.min(window.innerWidth / STAGE.width, window.innerHeight / STAGE.height)
}

/*
 * The TV is designed at 1920x1080 px. The stage keeps that layout and scales it to fit any
 * screen (720p TV, 4K, a laptop), letterboxing the rest. Double-click toggles fullscreen.
 */
export function TvStage({ className, children }: { className?: string; children: ReactNode }) {
  const [scale, setScale] = useState(fitScale)

  useEffect(() => {
    const onResize = () => setScale(fitScale())
    const toggleFullscreen = () => {
      if (document.fullscreenElement) document.exitFullscreen().catch(() => {})
      else document.documentElement.requestFullscreen().catch(() => {})
    }
    window.addEventListener('resize', onResize)
    window.addEventListener('dblclick', toggleFullscreen)
    return () => {
      window.removeEventListener('resize', onResize)
      window.removeEventListener('dblclick', toggleFullscreen)
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
