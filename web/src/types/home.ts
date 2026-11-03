/** One activity log row: who changed what. Shown as "{actor_name} {summary}". */
export type Activity = {
  id: number
  actor_id: number // 0 when nobody was logged in
  actor_name: string // '' when nobody was logged in
  action: string // e.g. 'slide.created'
  entity_id: number // 0 for a setting
  summary: string
  created_at: string
}
