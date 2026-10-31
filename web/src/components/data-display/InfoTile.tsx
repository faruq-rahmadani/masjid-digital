import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'

type InfoTileProps = {
  label: ReactNode
  value: ReactNode
  className?: string
}

// Read-only label/value box (Profil masjid: Koordinat, Zona waktu). Lay out two per row.
export function InfoTile({ label, value, className }: InfoTileProps) {
  return (
    <dl className={cn('rounded-md bg-canvas px-3 py-2.5', className)}>
      <dt className="text-2xs font-semibold text-ink-muted">{label}</dt>
      <dd className="mt-0.5 text-sm font-bold">{value}</dd>
    </dl>
  )
}
