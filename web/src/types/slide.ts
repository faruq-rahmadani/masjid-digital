// Slide (E6-04): images rotating in the TV's main panel, the same on every screen. Mirrors
// docs/openapi.yaml.

export type Slide = {
  id: number
  title: string
  /** Optional second line under the title on the TV; "" when not set. */
  caption: string
  duration_seconds: number
  /** Whether the TV shows the title and caption; off for posters with their own text. */
  show_caption: boolean
  active: boolean
  position: number
  created_by: number
  /** "" when nobody created it. */
  created_by_name: string
  created_at: string
  updated_at: string
}

/** What the admin fills in for a new slide; it is active once saved. */
export type CreateSlideRequest = Pick<
  Slide,
  'title' | 'caption' | 'duration_seconds' | 'show_caption'
>

export type UpdateSlideRequest = CreateSlideRequest & Pick<Slide, 'active'>

/** What the TV shows: an active slide; its image is at slideImageUrl(id). */
export type DisplaySlide = Pick<
  Slide,
  'id' | 'title' | 'caption' | 'duration_seconds' | 'show_caption'
>
