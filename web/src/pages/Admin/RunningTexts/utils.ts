import type { SegmentOption } from '@/components'
import type { ScrollSpeed } from '@/types/runningText'

/** Same limit as the API (domain.MaxRunningTextLength). */
export const MAX_TEXT_LENGTH = 200

export const SPEED_OPTIONS: readonly SegmentOption<ScrollSpeed>[] = [
  { value: 'slow', label: 'Lambat' },
  { value: 'medium', label: 'Sedang' },
  { value: 'fast', label: 'Cepat' },
]

/** The TV polls every 5 minutes (usePublicRunningTexts); push updates come with Layar TV (E9). */
export const TV_DELAY = 'Tampil di TV dalam ±5 menit'
