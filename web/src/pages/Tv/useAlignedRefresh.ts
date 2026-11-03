import { type QueryClient, useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'

/** How often the TV reloads its data. */
export const TV_REFRESH_MS = 5 * 60_000

/** The refresh round a moment falls in, by the server's clock: the same on every TV. */
export const refreshRound = (serverMs: number) => Math.floor(serverMs / TV_REFRESH_MS)

/** The TV's data: the public reads and the mocked ones (useTvData). */
const isTvQuery = (key: readonly unknown[]) => key[0] === 'public' || key[0] === 'tv'

/** Reloads every TV query on screen now; a failed reload keeps the last data. */
export function refetchTvData(queryClient: QueryClient) {
  return queryClient.refetchQueries({ predicate: (q) => isTvQuery(q.queryKey), type: 'active' })
}

/**
 * Reload the TV's data on the server clock's 5-minute marks (…:00, …:05, …) rather than 5 minutes
 * after this TV opened, so after an admin change every TV switches to the new slides and running
 * text within seconds of the others, and their slide shows stay in step. A failed reload keeps
 * the last data on screen.
 */
export function useAlignedRefresh(now: Date) {
  const queryClient = useQueryClient()
  const round = refreshRound(now.getTime())
  const last = useRef(round)

  useEffect(() => {
    if (round === last.current) return
    last.current = round
    refetchTvData(queryClient)
  }, [round, queryClient])
}
