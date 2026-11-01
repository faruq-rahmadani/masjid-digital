import { Ticker } from '@/components'
import { usePublicRunningTexts } from '@/hooks/useRunningTexts'

// INFO bar along the bottom of the TV; hidden while there is no active text.
export function RunningText() {
  const display = usePublicRunningTexts().data
  if (!display) return null
  return <Ticker texts={display.texts} speed={display.speed} />
}
