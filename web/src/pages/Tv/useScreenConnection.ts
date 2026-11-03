import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef, useState } from 'react'
import { screenEventsUrl } from '@/services/screenService'
import type { ScreenState } from '@/types/screen'
import { env } from '@/utils/env'
import { refetchTvData } from './useAlignedRefresh'

const DEVICE_KEY = 'masjid-digital:device-id'
/** After a refused connection (not a network drop, which the browser retries itself). */
const CLOSED_RETRY_MS = 30_000

export type ScreenConnection = {
  /** Its name and forced mode, from the API; null until the first connect. */
  state: ScreenState | null
  /** An admin blocked this TV; it keeps checking back every 5 minutes. */
  blocked: boolean
  connected: boolean
  /** Device time the connection dropped; null while connected or before the first connect. */
  disconnectedAt: number | null
}

/**
 * The TV's one live connection to the API (E9): it registers the TV on the Layar TV page, keeps
 * it online there, and brings its forced mode, content refreshes and commands. The TV runs on
 * without it, on its 5-minute aligned refresh.
 */
export function useScreenConnection(onTestSound: () => void): ScreenConnection {
  const queryClient = useQueryClient()
  const [connection, setConnection] = useState<ScreenConnection>({
    state: null,
    blocked: false,
    connected: false,
    disconnectedAt: null,
  })
  const testSound = useRef(onTestSound)
  testSound.current = onTestSound

  useEffect(() => {
    if (!('EventSource' in window)) return
    const nameHint = new URLSearchParams(window.location.search).get('layar') ?? ''
    const url = screenEventsUrl(deviceId(), nameHint, env.appVersion)
    let source: EventSource
    let retry: ReturnType<typeof setTimeout> | undefined
    const update = (patch: Partial<ScreenConnection>) => setConnection((c) => ({ ...c, ...patch }))

    const connect = () => {
      source = new EventSource(url)
      source.onopen = () => update({ connected: true, disconnectedAt: null })
      source.addEventListener('state', (e) =>
        update({ state: JSON.parse((e as MessageEvent).data), blocked: false }),
      )
      source.addEventListener('refresh', () => refetchTvData(queryClient))
      source.addEventListener('reload', () => window.location.reload())
      source.addEventListener('test_sound', () => testSound.current())
      source.addEventListener('blocked', () => update({ blocked: true }))
      source.onerror = () => {
        setConnection((c) => ({
          ...c,
          connected: false,
          disconnectedAt: c.disconnectedAt ?? Date.now(),
        }))
        // CONNECTING: the browser retries by itself. CLOSED: the API answered with an error.
        if (source.readyState === EventSource.CLOSED) retry = setTimeout(connect, CLOSED_RETRY_MS)
      }
    }

    connect()
    return () => {
      clearTimeout(retry)
      source.close()
    }
  }, [queryClient])

  return connection
}

/** This browser's ID, made once and kept, so the TV stays the same screen across reloads. */
function deviceId() {
  try {
    const stored = localStorage.getItem(DEVICE_KEY)
    if (stored) return stored
    const id = randomId()
    localStorage.setItem(DEVICE_KEY, id)
    return id
  } catch {
    // storage blocked: a new screen on every load, still better than none
    return randomId()
  }
}

/** crypto.randomUUID needs a secure context, which a TV on http:// may not have. */
function randomId() {
  if (typeof crypto.randomUUID === 'function' && window.isSecureContext) return crypto.randomUUID()
  const bytes = crypto.getRandomValues(new Uint8Array(16))
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('')
}

/** The forced mode while it lasts, judged by the server's clock. */
export function activeMode(state: ScreenState | null, now: Date) {
  if (!state?.forced_until || new Date(state.forced_until) <= now) return 'auto'
  return state.forced_mode
}
