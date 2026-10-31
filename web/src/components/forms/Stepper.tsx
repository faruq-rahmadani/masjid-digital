import { Minus, Plus } from 'lucide-react'
import { cn } from '@/utils/cn'

const sizes = {
  sm: { button: 'size-8 rounded-sm', value: 'min-w-6 text-base' }, // prayer-time correction table
  md: { button: 'size-9 rounded-md', value: 'min-w-14 text-base' }, // adzan duration
} as const

type StepperProps = {
  /** What is being changed, e.g. "Koreksi Subuh". Used for the button labels. */
  label: string
  value: number
  onChange: (value: number) => void
  min: number
  max: number
  step?: number
  /** Display text for the value, e.g. (v) => `${v} mnt`. */
  format?: (value: number) => string
  size?: keyof typeof sizes
  className?: string
}

export function Stepper({
  label,
  value,
  onChange,
  min,
  max,
  step = 1,
  format = String,
  size = 'sm',
  className,
}: StepperProps) {
  const set = (next: number) => onChange(Math.min(max, Math.max(min, next)))
  const s = sizes[size]

  return (
    <fieldset
      aria-label={label}
      className={cn('m-0 inline-flex min-w-0 items-center gap-1.5 border-0 p-0', className)}
    >
      <button
        type="button"
        aria-label={`Kurangi ${label}`}
        disabled={value <= min}
        onClick={() => set(value - step)}
        className={cn(
          'border-line bg-card text-ink inline-flex items-center justify-center border disabled:opacity-40',
          s.button,
        )}
      >
        <Minus aria-hidden className="size-4" />
      </button>
      <output aria-live="polite" className={cn('text-center font-bold', s.value)}>
        {format(value)}
      </output>
      <button
        type="button"
        aria-label={`Tambah ${label}`}
        disabled={value >= max}
        onClick={() => set(value + step)}
        className={cn(
          'border-line bg-card text-ink inline-flex items-center justify-center border disabled:opacity-40',
          s.button,
        )}
      >
        <Plus aria-hidden className="size-4" />
      </button>
    </fieldset>
  )
}
