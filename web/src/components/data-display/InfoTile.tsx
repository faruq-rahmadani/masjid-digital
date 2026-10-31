import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'

type InfoTileProps = {
  label: ReactNode
  value: ReactNode
  className?: string
}

// Read-only label/value box (Profil masjid: Koordinat, Zona waktu). Lay out two per row.
export function InfoTile({ label, value, className }: InfoTileProps) {
  return (
    <dl className={cn('bg-canvas rounded-md px-3 py-2.5', className)}>
      <dt className="text-2xs text-ink-muted font-semibold">{label}</dt>
      <dd className="mt-0.5 text-sm font-bold">{value}</dd>
    </dl>
  )
}
