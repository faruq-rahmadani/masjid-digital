import type { ComponentProps } from 'react'
import { type ControlSize, controlBase, controlSizes } from '@/components/forms/control-styles'
import { useFieldControl } from '@/components/forms/field-context'
import { cn } from '@/utils/cn'
import { formatNumber } from '@/utils/format/money'

type AmountInputProps = Omit<ComponentProps<'input'>, 'size' | 'value' | 'onChange' | 'max'> & {
  /** Integer Rupiah, or null when empty. */
  value: number | null
  onChange: (value: number | null) => void
  max?: number
  size?: ControlSize
}

/*
 * A row-sized Rupiah entry (split lines, saldo awal, saldo hitung), sitting in a Field. Same rules
 * as MoneyInput: digits only, grouped as you type, whole value selected on focus.
 */
export function AmountInput({
  value,
  onChange,
  max = 999_999_999_999,
  size = 'md',
  className,
  ...props
}: AmountInputProps) {
  return (
    <div className="relative">
      <span
        aria-hidden
        className="text-ink-muted pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-sm font-semibold"
      >
        Rp
      </span>
      <input
        inputMode="numeric"
        autoComplete="off"
        placeholder="0"
        className={cn(controlBase, controlSizes[size], 'pl-10 text-right font-semibold', className)}
        {...useFieldControl(props)}
        value={value === null ? '' : formatNumber(value)}
        // a click on an already focused field selects too, so typing never lands mid-number
        onFocus={(event) => event.currentTarget.select()}
        onClick={(event) => event.currentTarget.select()}
        onChange={(event) => {
          const digits = event.currentTarget.value.replace(/\D/g, '')
          if (digits === '') return onChange(null)
          const next = Number(digits)
          if (next <= max) onChange(next)
        }}
      />
    </div>
  )
}
