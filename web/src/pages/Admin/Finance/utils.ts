import type { FinanceLookups } from '@/hooks/useFinance'
import type { Transaction } from '@/types/finance'
import { isPindah } from '@/utils/finance'

export type KindFilter = 'all' | 'in' | 'out' | 'pindah'

export type RowFilter = {
  /** null: every kas */
  pocketId: number | null
  kind: KindFilter
  search: string
}

/** The rows the list shows, newest first. */
export function visibleRows(txs: Transaction[], filter: RowFilter, lookups: FinanceLookups) {
  const search = filter.search.trim().toLowerCase()
  return txs
    .filter((tx) => filter.pocketId === null || tx.pocket_id === filter.pocketId)
    .filter((tx) => {
      const category = lookups.category.get(tx.category_id)
      const pindah = category ? isPindah(category) : false
      if (filter.kind === 'pindah') return pindah
      if (filter.kind === 'all') return true
      return !pindah && tx.kind === filter.kind
    })
    .filter((tx) => !search || rowTitle(tx, txs, lookups).toLowerCase().includes(search))
    .sort((a, b) => b.date.localeCompare(a.date) || b.id - a.id)
}

/** The other row of a Pindah dana pair, when it is in the list. */
export function partnerOf(tx: Transaction, txs: Transaction[]) {
  if (!tx.group_id) return undefined
  return txs.find((t) => t.group_id === tx.group_id && t.id !== tx.id)
}

/** What a row is called: the note, or for a Pindah dana where the money went or came from. */
export function rowTitle(tx: Transaction, txs: Transaction[], lookups: FinanceLookups) {
  const category = lookups.category.get(tx.category_id)
  if (category && isPindah(category)) {
    const partner = partnerOf(tx, txs)
    if (partner && partner.pocket_id === tx.pocket_id) {
      const account = lookups.account.get(partner.account_id)?.name ?? 'akun lain'
      return tx.kind === 'out' ? `Pindah ke ${account}` : `Pindahan dari ${account}`
    }
    const pocket = partner
      ? (lookups.pocket.get(partner.pocket_id)?.name ?? 'kas lain')
      : 'kas lain'
    return tx.kind === 'out' ? `Pindah ke ${pocket}` : `Pindahan dari ${pocket}`
  }
  return tx.note || category?.name || 'Transaksi'
}

/** A row dated before the first open date belongs to a closed month. */
export const isLocked = (tx: Transaction, lastClosed: string | null) =>
  lastClosed !== null && tx.date.slice(0, 7) <= lastClosed
