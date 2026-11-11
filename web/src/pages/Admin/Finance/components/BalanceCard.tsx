import { Badge, Card, ProgressBar, Stat } from '@/components'
import type { FinanceLookups } from '@/hooks/useFinance'
import type { FinanceSummary, Pocket } from '@/types/finance'
import { POCKET_KIND_LABEL, POCKET_KIND_TONE } from '@/utils/finance'
import { formatRupiah } from '@/utils/format/money'

type BalanceCardProps = {
  /** null: every kas together */
  pocket: Pocket | null
  summary: FinanceSummary
  lookups: FinanceLookups
}

// The saldo card: balance now, masuk and keluar today, and for a terikat kas its target.
export function BalanceCard({ pocket, summary, lookups }: BalanceCardProps) {
  const flow = pocket ? summary.pockets.find((p) => p.pocket_id === pocket.id) : undefined
  const balance = pocket ? pocket.balance : summary.balance
  const todayIn = pocket ? (flow?.today_in ?? 0) : summary.today_in
  const todayOut = pocket ? (flow?.today_out ?? 0) : summary.today_out
  const accounts = lookups.accounts.filter((a) => a.archived_at === null)

  return (
    <Card tone="brand" padding="xl" className="flex flex-col gap-3.5">
      <div>
        <div className="flex items-center gap-2">
          <p className="text-on-brand-muted text-sm">
            {pocket ? `Saldo ${pocket.name}` : 'Saldo semua kas'}
          </p>
          {pocket && (
            <Badge tone={POCKET_KIND_TONE[pocket.kind]} size="sm">
              {POCKET_KIND_LABEL[pocket.kind]}
            </Badge>
          )}
        </div>
        <p className="mt-1 text-3xl font-extrabold" data-testid="balance">
          {formatRupiah(balance)}
        </p>
        {pocket?.purpose && <p className="text-on-brand-muted mt-1 text-xs">{pocket.purpose}</p>}
        {!pocket && (
          <p className="text-on-brand-muted mt-1 text-xs">
            {accounts.map((a) => `${a.name} ${formatRupiah(a.balance)}`).join(' · ')}
          </p>
        )}
      </div>

      {pocket?.progress && pocket.target_amount !== null && (
        <div className="flex flex-col gap-1.5">
          <ProgressBar
            label={`Target ${pocket.name}`}
            value={pocket.progress.collected}
            max={pocket.target_amount}
            tone="gold"
          />
          <p className="text-on-brand-muted text-xs">
            {pocket.progress.target_reached
              ? `Target tercapai · terkumpul ${formatRupiah(pocket.progress.collected)}`
              : `Terkumpul ${formatRupiah(pocket.progress.collected)} dari ${formatRupiah(pocket.target_amount)}`}
          </p>
        </div>
      )}

      <div className="grid grid-cols-2 gap-2.5">
        <Card tone="brand-2" padding="md">
          <Stat label="Masuk hari ini" value={formatRupiah(todayIn)} />
        </Card>
        <Card tone="brand-2" padding="md">
          <Stat label="Keluar hari ini" value={formatRupiah(todayOut)} />
        </Card>
      </div>
    </Card>
  )
}
