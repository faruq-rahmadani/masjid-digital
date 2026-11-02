import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { refreshRound, TV_REFRESH_MS, useAlignedRefresh } from './useAlignedRefresh'

describe('refreshRound', () => {
  it('changes on the server clock’s 5-minute marks, whenever the TV opened', () => {
    const mark = Date.UTC(2026, 10, 2, 5, 10) // 12:10 WIB
    expect(refreshRound(mark - 1)).toBe(refreshRound(mark) - 1)
    expect(refreshRound(mark)).toBe(refreshRound(mark + TV_REFRESH_MS - 1))
  })
})

describe('useAlignedRefresh', () => {
  it('reloads the TV data once each time a 5-minute mark passes, not on open', () => {
    const queryClient = new QueryClient()
    const refetch = vi.spyOn(queryClient, 'refetchQueries').mockResolvedValue()
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    )
    const mark = Date.UTC(2026, 10, 2, 5, 10)

    const { rerender } = renderHook(({ now }) => useAlignedRefresh(now), {
      wrapper,
      initialProps: { now: new Date(mark - 2000) },
    })
    rerender({ now: new Date(mark - 1000) })
    expect(refetch).not.toHaveBeenCalled()

    rerender({ now: new Date(mark) })
    rerender({ now: new Date(mark + 1000) })
    expect(refetch).toHaveBeenCalledOnce()

    // only the TV's queries
    const filters = refetch.mock.calls[0]?.[0] as unknown as {
      predicate: (q: { queryKey: unknown[] }) => boolean
    }
    const { predicate } = filters
    expect(predicate({ queryKey: ['public', 'slides'] })).toBe(true)
    expect(predicate({ queryKey: ['tv', 'quotes'] })).toBe(true)
    expect(predicate({ queryKey: ['slides'] })).toBe(false)
    expect(predicate({ queryKey: ['server-clock-offset'] })).toBe(false)
  })
})
