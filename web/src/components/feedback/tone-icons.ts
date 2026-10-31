import { CircleAlert, CircleCheck, Info, TriangleAlert, type LucideIcon } from 'lucide-react'

// One meaning per tone across Toast and Banner. Pair the icon with text; never rely on color alone.
export type FeedbackTone = 'info' | 'success' | 'warning' | 'danger'

export const toneIcons: Record<FeedbackTone, LucideIcon> = {
  info: Info,
  success: CircleCheck,
  warning: TriangleAlert,
  danger: CircleAlert,
}
