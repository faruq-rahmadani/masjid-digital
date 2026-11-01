export type ScreenStatus = 'online' | 'offline'

/** A TV that opened /tv (E9-02: heartbeat, last seen). */
export type Screen = {
  id: number
  name: string
  status: ScreenStatus
  last_seen_at: string
}

export type Activity = {
  id: number
  actor_name: string
  summary: string
  created_at: string
}
