import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'

// <fieldset> gives the group role natively; reset its default border and padding.
const fieldsetReset = 'm-0 min-w-0 border-0 p-0'

export type SegmentOption<T extends string> = { value: T; label: ReactNode }

const groupStyles = {
  track: 'grid grid-flow-col auto-cols-fr gap-1.5 rounded-xl bg-track p-1', // Pemasukan | Pengeluaran
  split: 'grid grid-flow-col auto-cols-fr gap-2', // Hijri −2..+2, scroll speed
  pill: 'flex flex-wrap gap-2', // Semua / Masuk / Keluar filters
} as const

const itemStyles = {
  track: 'h-11 rounded-md text-md font-semibold text-ink-soft',
  split: 'h-10 rounded-md border border-line bg-card text-sm font-semibold text-ink',
  pill: 'h-9 shrink-0 whitespace-nowrap rounded-full border border-line bg-card px-3.5 text-sm font-semibold text-ink',
} as const

const activeStyles = {
  brand: 'border-brand bg-brand text-on-brand font-bold',
  income: 'border-income bg-income text-white font-bold',
  expense: 'border-expense bg-expense text-white font-bold',
} as const

type SegmentedControlProps<T extends string> = {
  /** Accessible name for the group. */
  label: string
  options: readonly SegmentOption<T>[]
  value: T
  onChange: (value: T) => void
  variant?: keyof typeof groupStyles
  /** Fill of the selected item; per option so Pemasukan/Pengeluaran can differ. */
  tone?: keyof typeof activeStyles | ((value: T) => keyof typeof activeStyles)
  className?: string
}

export function SegmentedControl<T extends string>({
  label,
  options,
  value,
  onChange,
  variant = 'split',
  tone = 'brand',
  className,
}: SegmentedControlProps<T>) {
  return (
    <fieldset aria-label={label} className={cn(fieldsetReset, groupStyles[variant], className)}>
      {options.map((option) => {
        const selected = option.value === value
        const activeTone = typeof tone === 'function' ? tone(option.value) : tone
        return (
          <button
            key={option.value}
            type="button"
            aria-pressed={selected}
            onClick={() => onChange(option.value)}
            className={cn(
              'transition-colors',
              itemStyles[variant],
              selected && activeStyles[activeTone],
            )}
          >
            {option.label}
          </button>
        )
      })}
    </fieldset>
  )
}
