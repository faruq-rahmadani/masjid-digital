// Layar TV (E9): the TVs that opened /tv and keep a live connection to the API.

/** `auto` follows the schedule; the others are forced by an admin until the mosque's midnight. */
export type ScreenMode = 'auto' | 'normal' | 'jumat' | 'dark'

export type ScreenStatus = 'online' | 'offline'

export type Screen = {
  id: number
  device_id: string
  /** "" until an admin names it (BARU). */
  name: string
  /** From /tv?layar=…, suggested when naming. */
  name_hint: string
  user_agent: string
  forced_mode: ScreenMode
  forced_until: string | null
  blocked_at: string | null
  first_seen_at: string
  updated_at: string
  /** Online while connected, and for 2 minutes after. */
  status: ScreenStatus
  /** null when the server's cache has no record of it. */
  last_seen_at: string | null
  app_version: string
}

export type UpdateScreenRequest = Pick<Screen, 'name' | 'forced_mode'>

export type ScreenCommand = 'reload' | 'test_sound'

/** What a TV hears about itself on its connection. */
export type ScreenState = Pick<Screen, 'name' | 'forced_mode' | 'forced_until'>
