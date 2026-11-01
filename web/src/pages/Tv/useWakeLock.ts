import { useEffect } from 'react'

/** Keeps the screen from sleeping while the TV page is open (where the browser supports it). */
export function useWakeLock() {
  useEffect(() => {
    if (!('wakeLock' in navigator)) return
    let lock: WakeLockSentinel | null = null
    let cancelled = false
    const acquire = () => {
      if (document.visibilityState !== 'visible') return
      navigator.wakeLock
        .request('screen')
        .then((sentinel) => {
          if (cancelled) sentinel.release().catch(() => {})
          else lock = sentinel
        })
        .catch(() => {})
    }
    // The browser drops the lock when the tab is hidden; take it again on return.
    acquire()
    document.addEventListener('visibilitychange', acquire)
    return () => {
      cancelled = true
      document.removeEventListener('visibilitychange', acquire)
      lock?.release().catch(() => {})
    }
  }, [])
}
