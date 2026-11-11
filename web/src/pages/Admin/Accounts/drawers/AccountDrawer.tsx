import { useState } from 'react'
import {
  Button,
  Callout,
  ConfirmDialog,
  Field,
  Input,
  SegmentedControl,
  Sheet,
  Stack,
  useToast,
} from '@/components'
import { useFinanceMutation } from '@/hooks/useFinance'
import {
  archiveAccount,
  createAccount,
  restoreAccount,
  updateAccount,
} from '@/services/financeService'
import type { Account, AccountKind } from '@/types/finance'
import { formatRupiah } from '@/utils/format/money'

type AccountDrawerProps = {
  /** null: Tambah akun */
  account: Account | null
  canEdit: boolean
  onClose: () => void
}

type Pending = 'save' | 'archive' | 'restore'

// Tambah / ubah akun. The kind never changes; an account is archived only at Rp 0, and the last
// cash account always stays.
export function AccountDrawer({ account, canEdit, onClose }: AccountDrawerProps) {
  const [name, setName] = useState(account?.name ?? '')
  const [kind, setKind] = useState<AccountKind>(account?.kind ?? 'bank')
  const [number, setNumber] = useState(account?.number ?? '')
  const [pending, setPending] = useState<Pending | null>(null)
  const [error, setError] = useState<string | null>(null)
  const toast = useToast()
  const archived = account?.archived_at != null
  const readOnly = !canEdit || archived

  const save = useFinanceMutation(() =>
    account ? updateAccount(account.id, { name, number }) : createAccount({ name, kind, number }),
  )
  const archive = useFinanceMutation(archiveAccount)
  const restore = useFinanceMutation(restoreAccount)
  const done = (message: string) => ({
    onSuccess: () => {
      toast({ message })
      onClose()
    },
    onError: (e: Error) => setError(e.message),
  })

  const run = () => {
    setPending(null)
    if (pending === 'archive' && account) return archive.mutate(account.id, done('Akun diarsipkan'))
    if (pending === 'restore' && account) return restore.mutate(account.id, done('Akun dipulihkan'))
    save.mutate(undefined, done(account ? 'Akun disimpan' : 'Akun ditambahkan'))
  }

  return (
    <Sheet
      open
      onClose={onClose}
      title={account ? account.name : 'Tambah akun'}
      footer={
        !readOnly && (
          <Button
            size="xl"
            block
            loading={save.isPending}
            onClick={() => {
              if (!name.trim()) return setError('Isi nama akun.')
              setError(null)
              setPending('save')
            }}
          >
            Simpan
          </Button>
        )
      }
    >
      <Stack gap="md">
        {account && <p className="text-2xl font-extrabold">{formatRupiah(account.balance)}</p>}
        {!account && (
          <SegmentedControl
            label="Jenis akun"
            variant="split"
            options={[
              { value: 'bank', label: 'Rekening bank' },
              { value: 'cash', label: 'Tunai' },
            ]}
            value={kind}
            onChange={setKind}
          />
        )}
        <Field label="Nama akun">
          <Input
            placeholder={kind === 'bank' ? 'BSI Masjid' : 'Kotak Jumat'}
            maxLength={60}
            disabled={readOnly}
            value={name}
            onChange={(e) => setName(e.currentTarget.value)}
          />
        </Field>
        {kind === 'bank' && (
          <Field label="Nomor rekening" hint="Opsional">
            <Input
              inputMode="numeric"
              maxLength={40}
              disabled={readOnly}
              value={number}
              onChange={(e) => setNumber(e.currentTarget.value)}
            />
          </Field>
        )}
        {account && canEdit && !archived && (
          <Button variant="danger" size="lg" block onClick={() => setPending('archive')}>
            Arsipkan akun
          </Button>
        )}
        {account && canEdit && archived && (
          <Button variant="secondary" size="lg" block onClick={() => setPending('restore')}>
            Pulihkan akun
          </Button>
        )}
        {error && <Callout tone="danger">{error}</Callout>}
      </Stack>
      <ConfirmDialog
        open={pending !== null}
        onClose={() => setPending(null)}
        title={confirmTitle(pending, account)}
        description={
          pending === 'archive' ? 'Akun hanya bisa diarsipkan saat saldonya Rp 0.' : undefined
        }
        confirmLabel={pending === 'archive' ? 'Arsipkan' : 'Simpan'}
        danger={pending === 'archive'}
        onConfirm={run}
      />
    </Sheet>
  )
}

function confirmTitle(pending: Pending | null, account: Account | null) {
  if (pending === 'archive') return `Arsipkan ${account?.name}?`
  if (pending === 'restore') return `Pulihkan ${account?.name}?`
  if (account) return 'Simpan perubahan akun?'
  return 'Tambah akun ini?'
}
