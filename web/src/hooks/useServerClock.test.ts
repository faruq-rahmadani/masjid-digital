import { afterEach, describe, expect, it, vi } from 'vitest'
import * as timeService from '@/services/timeService'
import { measureClockOffset } from './useServerClock'

afterEach(() => {
  vi.restoreAllMocks()
})

// Device clock reads `sent` before the request and `received` after it.
function deviceClock(sent: number, received: number) {
  vi.spyOn(Date, 'now').mockReturnValueOnce(sent).mockReturnValueOnce(received)
}

describe('measureClockOffset', () => {
  it('compares the server time with the middle of the round trip', async () => {
    deviceClock(1_000, 1_200) // device midpoint 1_100
    vi.spyOn(timeService, 'getServerTime').mockResolvedValue(61_100)

    expect(await measureClockOffset()).toBe(60_000) // device is a minute behind
  })

  it('rejects a reply too slow to trust', async () => {
    deviceClock(1_000, 5_000)
    vi.spyOn(timeService, 'getServerTime').mockResolvedValue(3_000)

    await expect(measureClockOffset()).rejects.toThrow()
  })
})
