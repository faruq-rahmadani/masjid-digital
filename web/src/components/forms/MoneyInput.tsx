import { useId, useState } from 'react'
import { cn } from '@/utils/cn'
import { type CashFlow, formatNumber } from '@/utils/format/money'

type MoneyInputProps = {
  label?: string
  /** Integer Rupiah, or null when empty. */
  value: number | null
  onChange: (value: number | null) => void
  /** Colors the amount: pemasukan green, pengeluaran rust. */
  flow?: CashFlow
  /** Guards against runaway input (PRD: server also validates a max). */
  max?: number
  className?: string
}

/*
 * The big "Nominal" entry on Tambah Transaksi. Digits only, grouped as you type, and the whole
 * value is selected on focus so retyping replaces it. A competitor's mask appended digits instead
 * and saved Rp 1.000.015.000 (docs/product/prd.md#settled-decisions).
 */
export function MoneyInput({
  label = 'Nominal',
  value,
  onChange,
  flow = 'in',
  max = 999_999_999_999,
  className,
}: MoneyInputProps) {
  const id = useId()
  const [focused, setFocused] = useState(false)

  return (
    <div className={cn('bg-card rounded-xl px-4 py-4.5 text-center', className)}>
      <label htmlFor={id} className="text-ink-soft text-xs font-semibold">
        {label}
      </label>
      <div
        className={cn(
          'mt-1 flex items-baseline justify-center gap-2 text-4xl font-extrabold',
          flow === 'in' ? 'text-income' : 'text-expense',
        )}
      >
        <span aria-hidden>Rp</span>
        <input
          id={id}
          inputMode="numeric"
          autoComplete="off"
          placeholder="0"
          value={value === null ? '' : formatNumber(value)}
          onFocus={(event) => {
            setFocused(true)
            event.currentTarget.select()
          }}
          onBlur={() => setFocused(false)}
          onChange={(event) => {
            const digits = event.currentTarget.value.replace(/\D/g, '')
            if (digits === '') return onChange(null)
            const next = Number(digits)
            if (next <= max) onChange(next)
          }}
          // The visible width follows the number so "Rp" stays attached to it.
          style={{
            width: `${Math.max(1, (value === null ? 1 : formatNumber(value).length) + 0.5)}ch`,
          }}
          className={cn(
            'placeholder:text-line min-w-0 bg-transparent text-left outline-none',
            focused && 'decoration-gold underline decoration-2 underline-offset-8',
          )}
        />
      </div>
    </div>
  )
}
