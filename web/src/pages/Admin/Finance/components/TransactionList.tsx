import { Badge, Callout, Card, EmptyState, List, Money, Skeleton } from '@/components'
import type { FinanceLookups } from '@/hooks/useFinance'
import type { Transaction } from '@/types/finance'
import { cn } from '@/utils/cn'
import { formatDayShort } from '@/utils/finance'
import { isLocked, rowTitle } from '../utils'

type TransactionListProps = {
  /** undefined while loading */
  rows: Transaction[] | undefined
  /** Every row of the month, to name the other side of a Pindah dana. */
  all: Transaction[]
  lookups: FinanceLookups
  showPocket: boolean
  error?: string
  onOpen: (tx: Transaction) => void
}

export function TransactionList({
  rows,
  all,
  lookups,
  showPocket,
  error,
  onOpen,
}: TransactionListProps) {
  if (error) return <Callout tone="danger">{error}</Callout>
  if (!rows) {
    return (
      <Card aria-busy aria-label="Memuat transaksi" className="flex flex-col gap-3">
        <Skeleton className="h-8" />
        <Skeleton className="h-8" />
      </Card>
    )
  }
  if (rows.length === 0) {
    return (
      <EmptyState
        title="Belum ada transaksi"
        description="Transaksi bulan ini akan tampil di sini."
      />
    )
  }

  return (
    <Card padding="list">
      <List>
        {rows.map((tx) => {
          const category = lookups.category.get(tx.category_id)
          const parts = [
            formatDayShort(tx.date),
            category?.name,
            showPocket && lookups.pocket.get(tx.pocket_id)?.name,
            lookups.account.get(tx.account_id)?.name,
          ].filter(Boolean)
          const deleted = tx.deleted_at !== null
          return (
            <li key={tx.id}>
              <button
                type="button"
                onClick={() => onOpen(tx)}
                className="flex w-full items-center justify-between gap-3 py-2.75 text-left"
              >
                <span className="min-w-0">
                  <span
                    className={cn(
                      'block truncate text-sm font-semibold',
                      deleted && 'text-ink-muted line-through',
                    )}
                  >
                    {rowTitle(tx, all, lookups)}
                  </span>
                  <span className="text-ink-soft flex flex-wrap items-center gap-1.5 text-xs">
                    {parts.join('  ·  ')}
                    {tx.updated_at !== tx.created_at && !deleted && (
                      <Badge tone="gold" size="sm">
                        Diedit
                      </Badge>
                    )}
                    {deleted && (
                      <Badge tone="danger" size="sm">
                        Dihapus
                      </Badge>
                    )}
                    {isLocked(tx, lookups.lastClosed) && (
                      <Badge tone="notice" size="sm">
                        Terkunci
                      </Badge>
                    )}
                  </span>
                </span>
                <Money
                  amount={tx.amount}
                  flow={tx.kind}
                  className={cn('shrink-0', deleted && 'opacity-50')}
                />
              </button>
            </li>
          )
        })}
      </List>
    </Card>
  )
}
