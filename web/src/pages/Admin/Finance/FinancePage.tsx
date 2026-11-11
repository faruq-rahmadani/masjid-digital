import { ArrowLeftRight, Minus, Plus } from 'lucide-react'
import { useState } from 'react'
import {
  Button,
  Callout,
  Card,
  Checkbox,
  NavCard,
  PageBody,
  PageHeader,
  SearchInput,
  SegmentedControl,
  Select,
  Skeleton,
  Stack,
} from '@/components'
import {
  type FinanceLookups,
  useBookkeeping,
  useFinanceLookups,
  useFinanceSummary,
  useTransactions,
} from '@/hooks/useFinance'
import { ROUTES } from '@/route'
import { useSession } from '@/stores/sessionStore'
import type { FinanceSummary, Transaction } from '@/types/finance'
import { accessTo, canOpen } from '@/utils/auth/permissions'
import { canMoveFunds, formatMonthShort, isOpen, monthsBetween } from '@/utils/finance'
import { formatRupiah } from '@/utils/format/money'
import { BalanceCard } from './components/BalanceCard'
import { CloseMonthCard } from './components/CloseMonthCard'
import { FinanceWarnings } from './components/FinanceWarnings'
import { TransactionList } from './components/TransactionList'
import { TvPeriodCard } from './components/TvPeriodCard'
import { CloseMonthDrawer } from './drawers/CloseMonthDrawer'
import { MoveFundsDrawer } from './drawers/MoveFundsDrawer'
import { TransactionDetailDrawer } from './drawers/TransactionDetailDrawer'
import { type TransactionDraft, TransactionDrawer } from './drawers/TransactionDrawer'
import { type KindFilter, visibleRows } from './utils'

// Keuangan (E7-06). Super admin and bendahara edit; takmir only reads. Balances and totals come
// computed from the API; this screen never adds money up itself.
export function FinancePage() {
  return (
    <>
      <PageHeader title="Keuangan" />
      <FinanceContent />
    </>
  )
}

type Drawer =
  | { kind: 'transaction'; draft: TransactionDraft }
  | { kind: 'move'; edit?: Transaction[] }
  | { kind: 'detail'; tx: Transaction }
  | { kind: 'close'; month: string }

function FinanceContent() {
  const session = useSession()
  const canEdit = accessTo(session.role, 'finance') === 'edit'
  const lookups = useFinanceLookups()
  const bookkeeping = useBookkeeping()
  const current = useFinanceSummary('')

  if (lookups.error || bookkeeping.error || current.error) {
    const error = lookups.error ?? bookkeeping.error ?? current.error
    return (
      <PageBody>
        <Callout tone="danger" title="Data keuangan gagal dimuat">
          {error?.message}
        </Callout>
      </PageBody>
    )
  }
  if (!lookups.ready || !bookkeeping.data || !current.data) {
    return (
      <PageBody>
        <Card aria-busy aria-label="Memuat keuangan" className="flex flex-col gap-3">
          <Skeleton className="h-28" />
          <Skeleton className="h-8" />
        </Card>
      </PageBody>
    )
  }
  if (!bookkeeping.data.started) {
    return (
      <PageBody>
        <Callout tone="notice" title="Pembukuan belum dimulai">
          {canEdit
            ? 'Isi saldo awal setiap akun dan kas dulu, setelah itu transaksi bisa dicatat.'
            : 'Bendahara belum mengisi saldo awal.'}
        </Callout>
        {canEdit && (
          <NavCard
            to={ROUTES.accounts}
            title="Mulai pembukuan"
            description="Saldo awal per akun dan per kas"
          />
        )}
      </PageBody>
    )
  }

  return (
    <Ledger
      lookups={lookups}
      current={current.data}
      startMonth={bookkeeping.data.start_month}
      canEdit={canEdit}
      canManageCategories={canOpen(session.role, 'categories')}
    />
  )
}

type LedgerProps = {
  lookups: FinanceLookups
  current: FinanceSummary
  startMonth: string
  canEdit: boolean
  canManageCategories: boolean
}

function Ledger({ lookups, current, startMonth, canEdit, canManageCategories }: LedgerProps) {
  const [pocketId, setPocketId] = useState<number | null>(null)
  const [month, setMonth] = useState(current.month)
  const [kind, setKind] = useState<KindFilter>('all')
  const [search, setSearch] = useState('')
  const [showDeleted, setShowDeleted] = useState(false)
  const [drawer, setDrawer] = useState<Drawer | null>(null)
  const summary = useFinanceSummary(month, month !== current.month)
  const monthSummary = month === current.month ? current : summary.data
  const transactions = useTransactions(month, showDeleted)

  const openPockets = lookups.pockets.filter(isOpen)
  const pocket = pocketId === null ? null : (lookups.pocket.get(pocketId) ?? null)
  const today = current.today
  const rows = transactions.data
    ? visibleRows(transactions.data, { pocketId, kind, search }, lookups)
    : undefined
  const pocketFlow = monthSummary?.pockets.find((p) => p.pocket_id === pocketId)
  const monthIn = (pocketId === null ? monthSummary?.total_in : pocketFlow?.in) ?? 0
  const monthOut = (pocketId === null ? monthSummary?.total_out : pocketFlow?.out) ?? 0

  const newTransaction = (type: 'in' | 'out') =>
    setDrawer({
      kind: 'transaction',
      draft: { type, pocketId: pocket?.id ?? openPockets[0]?.id ?? 0, date: today },
    })

  return (
    <>
      <PageBody>
        <FinanceWarnings warnings={current.warnings} pocketId={pocketId} />

        <SegmentedControl
          label="Pilih kas"
          variant="pill"
          options={[
            { value: 'all', label: 'Semua kas' },
            ...openPockets.map((p) => ({ value: String(p.id), label: p.name })),
          ]}
          value={pocketId === null ? 'all' : String(pocketId)}
          onChange={(v) => setPocketId(v === 'all' ? null : Number(v))}
        />

        <BalanceCard pocket={pocket} summary={current} lookups={lookups} />

        {canEdit && (
          <Stack gap="sm">
            <div className="grid grid-cols-2 gap-2.5">
              <Button variant="income" size="lg" onClick={() => newTransaction('in')}>
                <Plus aria-hidden className="size-4" /> Pemasukan
              </Button>
              <Button variant="expense" size="lg" onClick={() => newTransaction('out')}>
                <Minus aria-hidden className="size-4" /> Pengeluaran
              </Button>
            </div>
            {pocket && !canMoveFunds(pocket.kind) && (
              <p className="text-ink-soft text-xs">
                Dana zakat tidak bisa dipindah ke kas lain. Salurkan lewat pengeluaran.
              </p>
            )}
            <Button variant="secondary" size="lg" block onClick={() => setDrawer({ kind: 'move' })}>
              <ArrowLeftRight aria-hidden className="size-4" /> Pindah dana
            </Button>
          </Stack>
        )}

        <SearchInput
          label="Cari transaksi"
          placeholder="Cari keterangan transaksi"
          value={search}
          onChange={(e) => setSearch(e.currentTarget.value)}
          onClear={() => setSearch('')}
        />
        <div className="flex flex-wrap items-center gap-2">
          <SegmentedControl
            label="Jenis transaksi"
            variant="pill"
            options={[
              { value: 'all', label: 'Semua' },
              { value: 'in', label: 'Masuk' },
              { value: 'out', label: 'Keluar' },
              { value: 'pindah', label: 'Pindah' },
            ]}
            value={kind}
            onChange={setKind}
          />
          <Select
            aria-label="Bulan"
            size="pill"
            className="ml-auto"
            options={monthsBetween(startMonth, current.month).map((m) => ({
              value: m,
              label: formatMonthShort(m),
            }))}
            value={month}
            onChange={setMonth}
          />
        </div>
        <p className="text-ink-soft text-xs">
          {formatMonthShort(month)}: masuk {formatRupiah(monthIn)} · keluar {formatRupiah(monthOut)}
          {monthSummary?.closed && ' · sudah tutup buku'}
        </p>

        <TransactionList
          rows={rows}
          all={transactions.data ?? []}
          lookups={lookups}
          showPocket={pocketId === null}
          error={transactions.error?.message}
          onOpen={(tx) => setDrawer({ kind: 'detail', tx })}
        />
        <Checkbox
          label="Tampilkan transaksi yang dihapus"
          checked={showDeleted}
          onChange={(e) => setShowDeleted(e.currentTarget.checked)}
        />

        {canEdit && <TvPeriodCard />}
        {canEdit && (
          <CloseMonthCard
            lastClosed={lookups.lastClosed}
            currentMonth={current.month}
            onClose={(m) => setDrawer({ kind: 'close', month: m })}
          />
        )}

        <Stack gap="sm">
          <NavCard to={ROUTES.pockets} title="Kas" description="Kas umum, zakat, dan kas terikat" />
          <NavCard to={ROUTES.accounts} title="Akun" description="Tunai dan rekening bank" />
          {canManageCategories && (
            <NavCard
              to={ROUTES.categories}
              title="Kategori"
              description="Kategori pemasukan dan pengeluaran"
            />
          )}
        </Stack>
      </PageBody>

      {drawer?.kind === 'transaction' && (
        <TransactionDrawer
          draft={drawer.draft}
          lookups={lookups}
          today={today}
          canManageCategories={canManageCategories}
          onClose={() => setDrawer(null)}
        />
      )}
      {drawer?.kind === 'move' && (
        <MoveFundsDrawer
          lookups={lookups}
          today={today}
          fromPocketId={pocket?.id}
          edit={drawer.edit}
          onClose={() => setDrawer(null)}
        />
      )}
      {drawer?.kind === 'detail' && (
        <TransactionDetailDrawer
          tx={drawer.tx}
          all={transactions.data ?? []}
          lookups={lookups}
          canEdit={canEdit}
          today={today}
          onEdit={(next) => setDrawer(next)}
          onClose={() => setDrawer(null)}
        />
      )}
      {drawer?.kind === 'close' && (
        <CloseMonthDrawer month={drawer.month} lookups={lookups} onClose={() => setDrawer(null)} />
      )}
    </>
  )
}

export type FinanceDrawer = Drawer
