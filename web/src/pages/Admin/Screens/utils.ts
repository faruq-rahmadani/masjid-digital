import type { SelectOption } from '@/components'
import type { Screen, ScreenMode } from '@/types/screen'
import { formatTimeAgo } from '@/utils/format/date'

export const MODE_OPTIONS: readonly SelectOption<ScreenMode>[] = [
  { value: 'auto', label: 'Otomatis (ikut jadwal)' },
  { value: 'normal', label: 'Tampilan normal' },
  { value: 'jumat', label: 'Mode Jumat' },
  { value: 'dark', label: 'Layar gelap' },
]

export const modeLabel = (mode: ScreenMode) =>
  MODE_OPTIONS.find((o) => o.value === mode)?.label ?? mode

/** The mode the TV is in now: a forced mode ends at forced_until. */
export function currentMode(screen: Screen, now: Date): ScreenMode {
  if (!screen.forced_until || new Date(screen.forced_until) <= now) return 'auto'
  return screen.forced_mode
}

/** Blocked screens are hidden from the list; new (unnamed) ones are shown apart. */
export function groupScreens(screens: Screen[]) {
  const valid = screens.filter((s) => s.blocked_at === null)
  return {
    named: valid.filter((s) => s.name !== ''),
    fresh: valid.filter((s) => s.name === ''),
    blocked: screens.filter((s) => s.blocked_at !== null),
  }
}

/** "Aktif sekarang · versi 1.4.0", "Aktif 2 jam lalu · versi 1.1.4" */
export function seenLine(screen: Screen, now: Date) {
  const seen = seenText(screen, now)
  return screen.app_version ? `${seen}  ·  versi ${screen.app_version}` : seen
}

function seenText(screen: Screen, now: Date) {
  if (screen.status === 'online') return 'Aktif sekarang'
  if (!screen.last_seen_at) return 'Belum terlihat sejak server dimulai ulang'
  return `Aktif ${formatTimeAgo(screen.last_seen_at, now)}`
}

/** "Chrome Android", "Safari macOS" from a user agent; enough to tell TVs apart. */
export function deviceLabel(userAgent: string) {
  const browser = BROWSERS.find(([pattern]) => pattern.test(userAgent))?.[1] ?? 'Browser'
  const system = SYSTEMS.find(([pattern]) => pattern.test(userAgent))?.[1]
  return system ? `${browser} ${system}` : browser
}

// Order matters: Edge and Samsung Internet also say Chrome, Chrome also says Safari.
const BROWSERS: [RegExp, string][] = [
  [/SamsungBrowser/, 'Samsung Internet'],
  [/Edg\//, 'Edge'],
  [/Firefox\//, 'Firefox'],
  [/Chrome\//, 'Chrome'],
  [/Safari\//, 'Safari'],
]

const SYSTEMS: [RegExp, string][] = [
  [/Tizen/, 'Tizen TV'],
  [/Web0S|webOS/, 'webOS TV'],
  [/Android/, 'Android'],
  [/CrOS/, 'ChromeOS'],
  [/Windows/, 'Windows'],
  [/iPhone|iPad/, 'iOS'],
  [/Mac OS X/, 'macOS'],
  [/Linux/, 'Linux'],
]
