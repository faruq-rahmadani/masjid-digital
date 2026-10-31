import type { LucideIcon } from 'lucide-react'
import type { ModuleKey } from '@/utils/auth/permissions'

/** One entry of the Home "Kelola" menu. */
export type MenuItem = {
  label: string
  path: string
  icon: LucideIcon
  module: ModuleKey
  /** False until the screen is built; the route shows "Segera hadir" meanwhile. */
  ready: boolean
}
