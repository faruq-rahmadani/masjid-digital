import type { SegmentOption } from '@/components'
import type { RunningText, ScrollSpeed } from '@/types/runningText'

/** Same limit as the API (domain.MaxRunningTextLength). */
export const MAX_TEXT_LENGTH = 200

export const SPEED_OPTIONS: readonly SegmentOption<ScrollSpeed>[] = [
  { value: 'slow', label: 'Lambat' },
  { value: 'medium', label: 'Sedang' },
  { value: 'fast', label: 'Cepat' },
]

/** The TV polls every 5 minutes (usePublicRunningTexts); push updates come with Layar TV (E9). */
export const TV_DELAY = 'Tampil di TV dalam ±5 menit'

export type TextFilter = 'active' | 'inactive' | 'all'

export const matchesFilter = (text: RunningText, filter: TextFilter) =>
  filter === 'all' || text.active === (filter === 'active')

/**
 * The API saves the order of every text, but a filtered list only reorders the ones it shows:
 * put `moved` (the shown ids in their new order) back into the slots the shown texts held.
 */
export function mergeOrder(all: readonly number[], moved: readonly number[]) {
  const shown = new Set(moved)
  let next = 0
  return all.map((id) => (shown.has(id) ? (moved[next++] ?? id) : id))
}
