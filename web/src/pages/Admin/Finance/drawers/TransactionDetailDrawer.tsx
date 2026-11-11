import { useEffect, useState } from 'react'
import {
  Badge,
  Button,
  Callout,
  ConfirmDialog,
  Money,
  Sheet,
  Stack,
  UploadTile,
  useToast,
} from '@/components'
import {
  type FinanceLookups,
  useFinanceMutation,
  useReceipt,
  useRevisions,
} from '@/hooks/useFinance'
import { deleteTransaction, uploadReceipt } from '@/services/financeService'
import type { Transaction } from '@/types/finance'
import { formatDayShort, isPindah, resizeImage } from '@/utils/finance'
import { formatRupiah } from '@/utils/format/money'
import type { FinanceDrawer } from '../FinancePage'
import { isLocked, rowTitle } from '../utils'

type TransactionDetailDrawerProps = {
  tx: Transaction
  /** Every row of the month, for the rest of its group */
  all: Transaction[]
  lookups: FinanceLookups
  canEdit: boolean
  today: string
  onEdit: (next: FinanceDrawer) => void
  onClose: () => void
}

// One transaction with its group, receipt and history. Edit and delete in an open month (with a
// reason); a closed month is fixed with a correction entry instead.
export function TransactionDetailDrawer({
  tx,
  all,
  lookups,
  canEdit,
  today,
  onEdit,
  onClose,
}: TransactionDetailDrawerProps) {
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const toast = useToast()
  const revisions = useRevisions(tx.id)
  const receipt = useReceipt(tx.id, tx.has_photo)
  const remove = useFinanceMutation((reason: string) => deleteTransaction(tx.id, reason))
  const upload = useFinanceMutation(async (file: File) =>
    uploadReceipt(tx.id, await resizeImage(file)),
  )
  const [receiptUrl, setReceiptUrl] = useState<string | null>(null)

  useEffect(() => {
    if (!receipt.data) return
    const url = URL.createObjectURL(receipt.data)
    setReceiptUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [receipt.data])

  const category = lookups.category.get(tx.category_id)
  const pindah = category ? isPindah(category) : false
  const group = tx.group_id ? all.filter((t) => t.group_id === tx.group_id) : [tx]
  const deleted = tx.deleted_at !== null
  const locked = isLocked(tx, lookups.lastClosed)
  const editable = canEdit && !deleted && !locked

  const edit = () => {
    if (pindah) return onEdit({ kind: 'move', edit: group })
    onEdit({
      kind: 'transaction',
      draft: { type: tx.kind, pocketId: tx.pocket_id, date: tx.date, edit: group },
    })
  }
  const correct = () =>
    onEdit({
      kind: 'transaction',
      draft: { type: tx.kind, pocketId: tx.pocket_id, date: today, correctsId: tx.id },
    })

  const details: [string, string | undefined][] = [
    ['Tanggal', formatDayShort(tx.date)],
    ['Kas', lookups.pocket.get(tx.pocket_id)?.name],
    ['Akun', lookups.account.get(tx.account_id)?.name],
    ['Kategori', category?.name],
    [pindah ? 'Alasan' : 'Keterangan', tx.note || undefined],
    ['Catatan donatur', tx.donor_note || undefined],
    ['Kerelaan', tx.donor_consent ? 'Donatur sudah merelakan' : undefined],
    ['Koreksi untuk', tx.corrects_id ? `Transaksi #${tx.corrects_id}` : undefined],
  ]

  return (
    <Sheet
      open
      onClose={onClose}
      title={rowTitle(tx, all, lookups)}
      footer={
        editable && (
          <div className="grid grid-cols-2 gap-2.5">
            <Button variant="secondary" size="lg" onClick={edit}>
              Edit
            </Button>
            <Button variant="danger" size="lg" onClick={() => setDeleting(true)}>
              {pindah ? 'Batalkan' : 'Hapus'}
            </Button>
          </div>
        )
      }
    >
      <Stack gap="md">
        <div className="flex flex-col gap-1.5">
          <Money amount={tx.amount} flow={tx.kind} withSymbol className="text-3xl font-extrabold" />
          <div className="flex flex-wrap gap-1.5">
            {tx.updated_at !== tx.created_at && !deleted && <Badge tone="gold">Diedit</Badge>}
            {deleted && <Badge tone="danger">Dihapus</Badge>}
            {locked && <Badge tone="notice">Terkunci</Badge>}
          </div>
        </div>

        <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
          {details
            .filter(([, value]) => value)
            .map(([label, value]) => (
              <div key={label} className="contents">
                <dt className="text-ink-soft">{label}</dt>
                <dd className="font-semibold">{value}</dd>
              </div>
            ))}
        </dl>

        {group.length > 1 && !pindah && (
          <Callout tone="notice" title="Dibagi ke beberapa kas">
            {group
              .map((t) => `${lookups.pocket.get(t.pocket_id)?.name} ${formatRupiah(t.amount)}`)
              .join(' · ')}
            . Diedit dan dihapus bersama.
          </Callout>
        )}

        {locked && canEdit && !deleted && (
          <Callout
            tone="notice"
            action={
              <Button size="md" variant="secondary" onClick={correct}>
                Buat koreksi
              </Button>
            }
          >
            Bulan ini sudah tutup buku. Perbaiki dengan transaksi koreksi di bulan berjalan.
          </Callout>
        )}

        {receiptUrl && <img src={receiptUrl} alt="Foto bukti" className="w-full rounded-lg" />}
        {!tx.has_photo && editable && (
          <UploadTile
            variant="dropzone"
            label={upload.isPending ? 'Mengunggah…' : 'Tambah foto bukti'}
            accept="image/*"
            disabled={upload.isPending}
            onFiles={([file]) =>
              file &&
              upload.mutate(file, {
                onSuccess: () => {
                  toast({ message: 'Foto bukti disimpan' })
                  onClose()
                },
                onError: (e) => setError(e.message),
              })
            }
          />
        )}

        {revisions.data && revisions.data.length > 0 && (
          <Stack gap="sm">
            <p className="text-ink-soft text-xs font-semibold">Riwayat</p>
            {revisions.data.map((r) => (
              <p key={r.id} className="text-ink-soft text-xs">
                {new Date(r.created_at).toLocaleString('id-ID')} ·{' '}
                {r.action === 'edit' ? 'diedit' : 'dihapus'} (sebelumnya{' '}
                {formatRupiah(r.previous.amount)}). <b>Alasan:</b> {r.reason}
              </p>
            ))}
          </Stack>
        )}
        {error && <Callout tone="danger">{error}</Callout>}
      </Stack>

      <ConfirmDialog
        open={deleting}
        onClose={() => setDeleting(false)}
        title={pindah ? 'Batalkan pindah dana?' : 'Hapus transaksi?'}
        description={
          pindah
            ? 'Kedua sisi pindah dana dibatalkan dan saldo kembali seperti semula.'
            : 'Transaksi ditandai dihapus dan tidak dihitung di saldo. Datanya tetap tersimpan untuk riwayat.'
        }
        confirmLabel={pindah ? 'Batalkan' : 'Hapus'}
        danger
        requireReason
        onConfirm={(reason) => {
          setDeleting(false)
          remove.mutate(reason, {
            onSuccess: () => {
              toast({ message: pindah ? 'Pindah dana dibatalkan' : 'Transaksi dihapus' })
              onClose()
            },
            onError: (e) => setError(e.message),
          })
        }}
      />
    </Sheet>
  )
}
