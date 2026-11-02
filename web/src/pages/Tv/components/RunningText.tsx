import { useCallback } from 'react'
import { Ticker } from '@/components'
import { usePublicRunningTexts } from '@/hooks/useRunningTexts'
import { useClockOffset } from '@/hooks/useServerClock'

// INFO bar along the bottom of the TV; hidden while there is no active text. It scrolls in step
// with every other TV, by the server's clock.
export function RunningText() {
  const display = usePublicRunningTexts().data
  const offset = useClockOffset()
  const serverNow = useCallback(() => Date.now() + offset, [offset])
  if (!display) return null
  return <Ticker texts={display.texts} speed={display.speed} syncTo={serverNow} />
}
