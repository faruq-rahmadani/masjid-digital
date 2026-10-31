import { cn } from '@/lib/cn'
import { formatRupiah, formatSignedAmount, type CashFlow } from '@/lib/format'

type MoneyProps = {
  /** Integer Rupiah. */
  amount: number
  /** Signed and colored: pemasukan green, pengeluaran rust. Omit for a plain balance. */
  flow?: CashFlow
  /** Show "Rp" on signed amounts; list rows drop it. */
  withSymbol?: boolean
  className?: string
}

export function Money({ amount, flow, withSymbol = false, className }: MoneyProps) {
  if (!flow) return <span className={cn('tabular-nums', className)}>{formatRupiah(amount)}</span>

  return (
    <span
      className={cn(
        'font-bold whitespace-nowrap tabular-nums',
        flow === 'in' ? 'text-income' : 'text-expense',
        className,
      )}
    >
      {formatSignedAmount(amount, flow, { withSymbol })}
    </span>
  )
}
