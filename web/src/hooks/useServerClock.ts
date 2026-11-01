import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import * as timeService from '@/services/timeService'

/*
 * A TV's own clock is often unset or drifting, so the TV display runs on the server's time:
 * offset = server time − device time, measured at the middle of the request. Kept in
 * localStorage, so a TV that reloads while the server is unreachable still starts corrected.
 */

const STORAGE_KEY = 'masjid-digital:clock-offset'
const RESYNC_MS = 10 * 60_000
// Past this round trip the midpoint guess could be seconds off; keep the previous offset instead.
const MAX_ROUND_TRIP_MS = 3_000

function loadOffset(): number | undefined {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    const value = Number(stored)
    return stored !== null && Number.isFinite(value) ? value : undefined
  } catch {
    return undefined
  }
}

function saveOffset(offset: number) {
  try {
    localStorage.setItem(STORAGE_KEY, String(offset))
  } catch {
    // Storage blocked: the offset still applies until reload.
  }
}

/** Measures how far the device clock is from the server's, in milliseconds. */
export async function measureClockOffset(): Promise<number> {
  const sentAt = Date.now()
  const serverNow = await timeService.getServerTime()
  const receivedAt = Date.now()
  if (receivedAt - sentAt > MAX_ROUND_TRIP_MS) throw new Error('round trip too slow to sync')
  return serverNow - (sentAt + receivedAt) / 2
}

/** ms to add to Date.now() to get the server's time; 0 until the first sync ever succeeds. */
export function useClockOffset() {
  const offset = useQuery({
    queryKey: ['server-clock-offset'],
    queryFn: async () => {
      const measured = await measureClockOffset()
      saveOffset(measured)
      return measured
    },
    initialData: loadOffset,
    // A stored offset would count as fresh for staleTime; still measure right away on open.
    initialDataUpdatedAt: 0,
    refetchInterval: RESYNC_MS,
    refetchIntervalInBackground: true,
  })
  return offset.data ?? 0
}

/**
 * The server-corrected time, ticking on each whole second of it (so a clock showing seconds
 * changes with the real second, not up to 1 s late).
 */
export function useServerNow() {
  const offset = useClockOffset()
  const [now, setNow] = useState(() => new Date(Date.now() + offset))

  useEffect(() => {
    let timer: ReturnType<typeof setTimeout>
    const tick = () => {
      const corrected = Date.now() + offset
      setNow(new Date(corrected))
      timer = setTimeout(tick, 1000 - (corrected % 1000))
    }
    tick()
    return () => clearTimeout(timer)
  }, [offset])

  return now
}
