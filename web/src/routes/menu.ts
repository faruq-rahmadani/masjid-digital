import { AlignLeft, Bell, Clock, CreditCard, House, Image, Monitor, Users } from 'lucide-react'
import { ROUTES } from '@/route'
import type { MenuItem } from './types'

// Home "Kelola" grid, in mockup order. Labels are UI copy; each role sees only what it may open.
export const MENU: readonly MenuItem[] = [
  { label: 'Slide', path: ROUTES.slides, icon: Image, module: 'content', ready: false },
  {
    label: 'Running text',
    path: ROUTES.runningTexts,
    icon: AlignLeft,
    module: 'content',
    ready: false,
  },
  { label: 'Keuangan', path: ROUTES.finance, icon: CreditCard, module: 'finance', ready: false },
  {
    label: 'Waktu sholat',
    path: ROUTES.prayerTimes,
    icon: Clock,
    module: 'prayerTimes',
    ready: false,
  },
  { label: 'Iqamah', path: ROUTES.iqamah, icon: Bell, module: 'iqamah', ready: false },
  { label: 'Profil masjid', path: ROUTES.profile, icon: House, module: 'profile', ready: true },
  { label: 'Pengguna', path: ROUTES.users, icon: Users, module: 'users', ready: true },
  { label: 'Layar TV', path: ROUTES.screens, icon: Monitor, module: 'screens', ready: false },
]
