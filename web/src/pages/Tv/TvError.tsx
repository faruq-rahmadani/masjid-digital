import { useEffect } from 'react'
import { TvStage } from './components/TvStage'

/** How long the error shows before the TV reloads itself. */
const RELOAD_AFTER_MS = 30_000

// The TV's error page: nobody is there to press a button, so it reloads by itself, but only
// once the network is back. Reloading while offline would trade a working page for a browser
// error page.
export function TvError() {
  useEffect(() => {
    let timer: ReturnType<typeof setTimeout> | undefined
    const reloadSoon = () => {
      clearTimeout(timer)
      timer = setTimeout(() => {
        if (navigator.onLine) window.location.reload()
      }, RELOAD_AFTER_MS)
    }
    reloadSoon()
    window.addEventListener('online', reloadSoon)
    return () => {
      clearTimeout(timer)
      window.removeEventListener('online', reloadSoon)
    }
  }, [])

  return (
    <TvStage className="bg-brand">
      <div
        role="alert"
        className="text-on-brand m-auto flex flex-col items-center gap-4 text-center"
      >
        <p className="text-tv-40 font-bold">Layar sedang memuat ulang</p>
        <p className="text-on-brand-muted text-tv-26">
          Terjadi gangguan. Layar memuat ulang sendiri saat internet tersambung.
        </p>
      </div>
    </TvStage>
  )
}
