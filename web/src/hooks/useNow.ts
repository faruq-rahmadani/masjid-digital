import { useEffect, useState } from 'react'

/** The current time, refreshed every `intervalMs` (default 30 s, enough for a minute countdown). */
export function useNow(intervalMs = 30_000) {
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), intervalMs)
    return () => clearInterval(timer)
  }, [intervalMs])
  return now
}
