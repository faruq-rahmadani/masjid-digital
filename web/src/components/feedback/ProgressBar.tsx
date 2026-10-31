import { cn } from '@/utils/cn'

type ProgressBarProps = {
  value: number
  max?: number
  /** Accessible name, e.g. "Sisa waktu iqamah". */
  label: string
  tone?: 'gold' | 'brand' | 'income'
  className?: string
}

const fills = {
  gold: '[&::-webkit-progress-value]:bg-gold [&::-moz-progress-bar]:bg-gold',
  brand: '[&::-webkit-progress-value]:bg-brand [&::-moz-progress-bar]:bg-brand',
  income: '[&::-webkit-progress-value]:bg-income [&::-moz-progress-bar]:bg-income',
} as const

// Native <progress>, restyled. Iqamah countdown on TV, program targets later (P1).
export function ProgressBar({
  value,
  max = 100,
  label,
  tone = 'gold',
  className,
}: ProgressBarProps) {
  return (
    <progress
      value={value}
      max={max}
      aria-label={label}
      className={cn(
        'bg-line-soft h-2 w-full appearance-none overflow-hidden rounded-full',
        '[&::-webkit-progress-bar]:bg-line-soft [&::-moz-progress-bar]:rounded-full [&::-webkit-progress-value]:rounded-full',
        fills[tone],
        className,
      )}
    />
  )
}
