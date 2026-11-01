import { useEffect, useState } from 'react'
import { useClockOffset } from '@/hooks/useServerClock'

/**
 * The server-corrected current time, refreshed every `intervalMs` (default 30 s, enough for a
 * minute countdown). A phone's clock can be wrong too, and it decides "today" and the next
 * prayer; this shares the TV's offset (useClockOffset), cached across screens and reloads.
 */
export function useNow(intervalMs = 30_000) {
  const offset = useClockOffset()
  const [now, setNow] = useState(() => new Date(Date.now() + offset))
  useEffect(() => {
    setNow(new Date(Date.now() + offset))
    const timer = setInterval(() => setNow(new Date(Date.now() + offset)), intervalMs)
    return () => clearInterval(timer)
  }, [intervalMs, offset])
  return now
}
