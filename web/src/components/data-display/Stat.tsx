import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'

const valueSizes = {
  md: 'text-lg font-bold', // masuk / keluar hari ini
  lg: 'text-2xl font-extrabold', // next prayer
  xl: 'text-3xl font-extrabold', // saldo kas
} as const

type StatProps = {
  label: ReactNode
  value: ReactNode
  size?: keyof typeof valueSizes
  /** Stats sit on brand surfaces in the mockup; pass false on a white card. */
  onBrand?: boolean
  align?: 'start' | 'end'
  className?: string
  valueClassName?: string
}

export function Stat({
  label,
  value,
  size = 'md',
  onBrand = true,
  align = 'start',
  className,
  valueClassName,
}: StatProps) {
  return (
    <div className={cn(align === 'end' && 'text-right', className)}>
      <div className={cn('text-xs', onBrand ? 'text-on-brand-muted' : 'text-ink-muted')}>
        {label}
      </div>
      <div className={cn('mt-0.5', valueSizes[size], valueClassName)}>{value}</div>
    </div>
  )
}
