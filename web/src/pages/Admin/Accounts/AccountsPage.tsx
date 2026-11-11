import { Plus } from 'lucide-react'
import { useState } from 'react'
import {
  Badge,
  Button,
  Callout,
  Card,
  List,
  ListRow,
  PageBody,
  PageHeader,
  SectionHeader,
  Skeleton,
} from '@/components'
import { useAccounts, useBookkeeping } from '@/hooks/useFinance'
import { useSession } from '@/stores/sessionStore'
import type { Account } from '@/types/finance'
import { accessTo } from '@/utils/auth/permissions'
import { formatMonth } from '@/utils/finance'
import { formatRupiah } from '@/utils/format/money'
import { AccountDrawer } from './drawers/AccountDrawer'
import { StartBookkeepingDrawer } from './drawers/StartBookkeepingDrawer'

// Akun (E7-10): where the money sits (Tunai, bank accounts) with their balances, and "Mulai
// pembukuan" once, with the opening balance per account and per kas.
export function AccountsPage() {
  const accounts = useAccounts()
  const bookkeeping = useBookkeeping()
  const session = useSession()
  const canEdit = accessTo(session.role, 'finance') === 'edit'
  const [editing, setEditing] = useState<Account | 'new' | null>(null)
  const [starting, setStarting] = useState(false)

  const active = accounts.data?.filter((a) => a.archived_at === null) ?? []
  const archived = accounts.data?.filter((a) => a.archived_at !== null) ?? []
  const error = accounts.error ?? bookkeeping.error

  return (
    <>
      <PageHeader title="Akun" meta={accounts.data && `${active.length} akun`} />
      <PageBody>
        {error && <Callout tone="danger">{error.message}</Callout>}
        {(!accounts.data || !bookkeeping.data) && !error && <Skeleton className="h-40" />}

        {bookkeeping.data && !bookkeeping.data.started && (
          <Callout
            tone="notice"
            title="Pembukuan belum dimulai"
            action={
              canEdit && (
                <Button size="md" onClick={() => setStarting(true)}>
                  Mulai pembukuan
                </Button>
              )
            }
          >
            Tambahkan dulu semua rekening masjid, lalu isi saldo awalnya sekali.
          </Callout>
        )}
        {bookkeeping.data?.started && (
          <p className="text-ink-soft text-xs">
            Pembukuan mulai {formatMonth(bookkeeping.data.start_month)}
          </p>
        )}

        {accounts.data && (
          <Card padding="list">
            <List>
              {active.map((a) => (
                <AccountRow key={a.id} account={a} onOpen={() => setEditing(a)} />
              ))}
            </List>
          </Card>
        )}
        {canEdit && (
          <Button variant="secondary" size="lg" block onClick={() => setEditing('new')}>
            <Plus aria-hidden className="size-4" /> Tambah akun
          </Button>
        )}
        {archived.length > 0 && (
          <>
            <SectionHeader title={`Diarsipkan (${archived.length})`} />
            <Card padding="list">
              <List>
                {archived.map((a) => (
                  <AccountRow key={a.id} account={a} onOpen={() => setEditing(a)} />
                ))}
              </List>
            </Card>
          </>
        )}
      </PageBody>

      {editing && (
        <AccountDrawer
          account={editing === 'new' ? null : editing}
          canEdit={canEdit}
          onClose={() => setEditing(null)}
        />
      )}
      {starting && <StartBookkeepingDrawer onClose={() => setStarting(false)} />}
    </>
  )
}

function AccountRow({ account, onOpen }: { account: Account; onOpen: () => void }) {
  return (
    <ListRow
      onClick={onOpen}
      className="cursor-pointer"
      title={
        <span className="flex items-center gap-2">
          {account.name}
          <Badge tone="neutral" size="sm">
            {account.kind === 'cash' ? 'Tunai' : 'Bank'}
          </Badge>
        </span>
      }
      description={account.number || undefined}
      trailing={<span className="font-bold">{formatRupiah(account.balance)}</span>}
    />
  )
}
