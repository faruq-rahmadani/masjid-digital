// Running text (E6-05): texts scrolling in the TV's INFO bar. Mirrors docs/openapi.yaml.

export type ScrollSpeed = 'slow' | 'medium' | 'fast'

export type RunningText = {
  id: number
  text: string
  active: boolean
  position: number
  created_by: number
  /** "" when nobody created it. */
  created_by_name: string
  /** When it was last turned off, i.e. last shown on the TV; null while active. */
  deactivated_at: string | null
  created_at: string
  updated_at: string
}

export type UpdateRunningTextRequest = { text: string; active: boolean }

export type RunningTextSetting = { speed: ScrollSpeed }

/** What the TV shows: active texts in order. */
export type RunningTextDisplay = { speed: ScrollSpeed; texts: string[] }
