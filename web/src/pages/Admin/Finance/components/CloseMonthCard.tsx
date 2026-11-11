import { Lock } from 'lucide-react'
import { Button, Callout } from '@/components'
import { addMonths, formatMonth } from '@/utils/finance'

type CloseMonthCardProps = {
  /** YYYY-MM of the last closing; null before pembukuan */
  lastClosed: string | null
  currentMonth: string
  onClose: (month: string) => void
}

// Tutup buku goes month by month: the one after the last closing, once it has ended.
export function CloseMonthCard({ lastClosed, currentMonth, onClose }: CloseMonthCardProps) {
  if (lastClosed === null) return null
  const next = addMonths(lastClosed, 1)
  if (next >= currentMonth) {
    return (
      <p className="text-ink-soft flex items-center gap-1.5 text-xs">
        <Lock aria-hidden className="size-3" /> Sudah tutup buku sampai {formatMonth(lastClosed)}
      </p>
    )
  }
  return (
    <Callout
      tone="notice"
      title={`Tutup buku ${formatMonth(next)}`}
      action={
        <Button size="md" variant="primary" onClick={() => onClose(next)}>
          Tutup buku
        </Button>
      }
    >
      Transaksi {formatMonth(next)} di semua kas akan terkunci.
    </Callout>
  )
}
