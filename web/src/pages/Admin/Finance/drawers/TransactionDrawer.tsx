import { Trash2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import {
  AmountInput,
  Badge,
  Button,
  Callout,
  Checkbox,
  ConfirmDialog,
  Field,
  IconButton,
  Input,
  MoneyInput,
  SegmentedControl,
  Select,
  Sheet,
  Stack,
  Textarea,
  TextLink,
  UploadTile,
  useToast,
} from '@/components'
import { type FinanceLookups, useFinanceMutation } from '@/hooks/useFinance'
import { ROUTES } from '@/route'
import { createTransaction, updateTransaction, uploadReceipt } from '@/services/financeService'
import type { CategoryType, SaveTransactionRequest, Transaction } from '@/types/finance'
import {
  addMonths,
  CONFIRM_AMOUNT_ABOVE,
  formCategories,
  isOpen,
  MAX_AMOUNT,
  resizeImage,
} from '@/utils/finance'
import { formatRupiah } from '@/utils/format/money'

export type TransactionDraft = {
  type: CategoryType
  pocketId: number
  date: string
  /** The rows being edited: one, or every row of a split */
  edit?: Transaction[]
  /** A transaction in a closed month this one corrects */
  correctsId?: number
}

type Line = { pocketId: number; categoryId: number | null; amount: number | null }

type TransactionDrawerProps = {
  draft: TransactionDraft
  lookups: FinanceLookups
  today: string
  canManageCategories: boolean
  onClose: () => void
}

// Tambah transaksi (E7-07): pemasukan or pengeluaran into one kas, or a pemasukan split over
// several kas ("Bagi ke beberapa kas"). Saved after a confirm; an edit asks for its reason.
export function TransactionDrawer({
  draft,
  lookups,
  today,
  canManageCategories,
  onClose,
}: TransactionDrawerProps) {
  const editing = draft.edit
  const first = editing?.[0]
  const openAccounts = lookups.accounts.filter(
    (a) => a.archived_at === null || a.id === first?.account_id,
  )
  const openPockets = lookups.pockets.filter(isOpen)
  const defaultAccount = openAccounts.find((a) => a.kind === 'cash') ?? openAccounts[0]

  const [type, setType] = useState<CategoryType>(first?.kind ?? draft.type)
  const [accountId, setAccountId] = useState(first?.account_id ?? defaultAccount?.id ?? 0)
  const [date, setDate] = useState(first?.date ?? draft.date)
  const [note, setNote] = useState(first?.note ?? '')
  const [donorNote, setDonorNote] = useState(first?.donor_note ?? '')
  const [lines, setLines] = useState<Line[]>(
    editing?.map((t) => ({
      pocketId: t.pocket_id,
      categoryId: t.category_id,
      amount: t.amount,
    })) ?? [{ pocketId: draft.pocketId, categoryId: null, amount: null }],
  )
  const [photo, setPhoto] = useState<File | null>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [errors, setErrors] = useState<string[]>([])
  const [confirming, setConfirming] = useState(false)
  const toast = useToast()

  const create = useFinanceMutation(createTransaction)
  const update = useFinanceMutation((body: SaveTransactionRequest) =>
    updateTransaction(first?.id ?? 0, body),
  )
  const saving = create.isPending || update.isPending

  useEffect(() => {
    if (!photo) return setPreview(null)
    const url = URL.createObjectURL(photo)
    setPreview(url)
    return () => URL.revokeObjectURL(url)
  }, [photo])

  const split = lines.length > 1
  const total = lines.reduce((sum, l) => sum + (l.amount ?? 0), 0)
  const openFrom = lookups.lastClosed ? `${addMonths(lookups.lastClosed, 1)}-01` : undefined
  const setLine = (i: number, patch: Partial<Line>) =>
    setLines((all) => all.map((l, j) => (j === i ? { ...l, ...patch } : l)))

  const changeType = (next: CategoryType) => {
    setType(next)
    setLines((all) => [{ ...(all[0] as Line), categoryId: null }])
  }

  const validate = () => {
    const problems: string[] = []
    if (!accountId) problems.push('Pilih akun.')
    if (!date) problems.push('Isi tanggal.')
    lines.forEach((l, i) => {
      const where = split ? ` (baris ${i + 1})` : ''
      if (!l.amount) problems.push(`Isi nominal${where}.`)
      if (!l.categoryId) problems.push(`Pilih kategori${where}.`)
    })
    setErrors(problems)
    return problems.length === 0
  }

  const save = async (reason: string) => {
    const body: SaveTransactionRequest = {
      kind: type,
      date,
      account_id: accountId,
      lines: lines.map((l) => ({
        pocket_id: l.pocketId,
        category_id: l.categoryId ?? 0,
        amount: l.amount ?? 0,
      })),
      note,
      donor_note: type === 'in' ? donorNote : '',
      corrects_id: draft.correctsId ?? null,
      reason,
    }
    const mutation = editing ? update : create
    mutation.mutate(body, {
      onSuccess: async (rows) => {
        const saved = rows[0]
        if (photo && saved) {
          try {
            await uploadReceipt(saved.id, await resizeImage(photo))
          } catch (error) {
            toast({ tone: 'danger', message: `Transaksi disimpan, foto gagal: ${String(error)}` })
          }
        }
        toast({ message: editing ? 'Perubahan disimpan' : 'Transaksi disimpan' })
        onClose()
      },
      onError: (error) => setErrors([error.message]),
    })
  }

  return (
    <Sheet
      open
      onClose={onClose}
      title={drawerTitle(draft)}
      footer={
        <Button size="xl" block loading={saving} onClick={() => validate() && setConfirming(true)}>
          {editing ? 'Simpan perubahan' : 'Simpan transaksi'}
        </Button>
      }
    >
      <Stack gap="md">
        {editing ? (
          <Badge tone={type === 'in' ? 'success' : 'danger'}>
            {type === 'in' ? 'Pemasukan' : 'Pengeluaran'} · jenis tidak bisa diubah
          </Badge>
        ) : (
          <SegmentedControl
            label="Jenis transaksi"
            variant="track"
            options={[
              { value: 'in', label: 'Pemasukan' },
              { value: 'out', label: 'Pengeluaran' },
            ]}
            value={type}
            onChange={changeType}
            tone={(v) => (v === 'in' ? 'income' : 'expense')}
          />
        )}

        {draft.correctsId && (
          <Callout tone="notice">
            Koreksi untuk transaksi di bulan yang sudah tutup buku. Transaksi aslinya tetap ada.
          </Callout>
        )}

        {split ? (
          <Stack gap="sm">
            {lines.map((line, i) => (
              <SplitLine
                // biome-ignore lint/suspicious/noArrayIndexKey: lines have no id until saved
                key={i}
                index={i}
                line={line}
                type={type}
                lookups={lookups}
                pockets={openPockets.filter(
                  (p) => p.id === line.pocketId || !lines.some((l) => l.pocketId === p.id),
                )}
                canRemove={!editing && lines.length > 2}
                onChange={(patch) => setLine(i, patch)}
                onRemove={() => setLines((all) => all.filter((_, j) => j !== i))}
              />
            ))}
            {!editing && (
              <Button
                variant="ghost"
                size="md"
                onClick={() => {
                  const free = openPockets.find((p) => !lines.some((l) => l.pocketId === p.id))
                  if (free)
                    setLines((all) => [
                      ...all,
                      { pocketId: free.id, categoryId: null, amount: null },
                    ])
                }}
              >
                + Tambah kas
              </Button>
            )}
            <p className="text-ink-soft text-sm">
              Total diterima: <b className="text-ink">{formatRupiah(total)}</b>
            </p>
          </Stack>
        ) : (
          <>
            <MoneyInput
              value={lines[0]?.amount ?? null}
              onChange={(amount) => setLine(0, { amount })}
              flow={type}
              max={MAX_AMOUNT}
            />
            <SingleLine
              line={lines[0] as Line}
              type={type}
              lookups={lookups}
              pockets={openPockets}
              onChange={(patch) => setLine(0, patch)}
            />
          </>
        )}

        {type === 'in' && !editing && (
          <Checkbox
            label="Bagi ke beberapa kas"
            checked={split}
            onChange={(e) => {
              const on = e.currentTarget.checked
              setLines((all) => {
                const head = all[0] as Line
                if (!on) return [head]
                const other = openPockets.find((p) => p.id !== head.pocketId)
                return other ? [head, { pocketId: other.id, categoryId: null, amount: null }] : all
              })
            }}
          />
        )}

        <Field label="Akun" hint="Tempat uangnya: tunai atau rekening">
          <Select
            options={openAccounts.map((a) => ({
              value: String(a.id),
              label: `${a.name} · ${formatRupiah(a.balance)}`,
            }))}
            value={accountId ? String(accountId) : undefined}
            onChange={(v) => setAccountId(Number(v))}
          />
        </Field>
        <Field label="Tanggal">
          <Input
            type="date"
            value={date}
            min={openFrom}
            max={today}
            onChange={(e) => setDate(e.currentTarget.value)}
          />
        </Field>
        <Field label="Keterangan">
          <Textarea
            rows={2}
            maxLength={200}
            placeholder={
              type === 'in' ? "Contoh: Kotak amal ba'da Jumat" : 'Contoh: Token listrik November'
            }
            value={note}
            onChange={(e) => setNote(e.currentTarget.value)}
          />
        </Field>
        {type === 'in' && (
          <Field label="Catatan donatur" hint="Opsional, misalnya nama atau niat donatur">
            <Input
              maxLength={200}
              value={donorNote}
              onChange={(e) => setDonorNote(e.currentTarget.value)}
            />
          </Field>
        )}
        {!editing && (
          <Stack gap="sm">
            <UploadTile
              variant="dropzone"
              label="Tambah foto bukti"
              hint="Nota atau bukti transfer, otomatis diperkecil"
              accept="image/*"
              onFiles={([file]) => setPhoto(file ?? null)}
            />
            {preview && (
              <div className="flex items-center gap-3">
                <img
                  src={preview}
                  alt="Pratinjau foto bukti"
                  className="size-16 rounded-md object-cover"
                />
                <Button variant="ghost" size="sm" onClick={() => setPhoto(null)}>
                  Hapus foto
                </Button>
              </div>
            )}
          </Stack>
        )}
        {canManageCategories && (
          <TextLink to={ROUTES.categories} size="sm">
            Kelola kategori
          </TextLink>
        )}

        {errors.length > 0 && (
          <Callout tone="danger">
            {errors.length === 1 ? (
              errors[0]
            ) : (
              <ul className="list-disc pl-4">
                {errors.map((e) => (
                  <li key={e}>{e}</li>
                ))}
              </ul>
            )}
          </Callout>
        )}
      </Stack>

      <ConfirmDialog
        open={confirming}
        onClose={() => setConfirming(false)}
        title={editing ? 'Simpan perubahan?' : 'Simpan transaksi?'}
        requireReason={Boolean(editing)}
        description={
          <>
            {type === 'in' ? 'Pemasukan' : 'Pengeluaran'} <b>{formatRupiah(total)}</b>
            {split
              ? ` dibagi ke ${lines.length} kas.`
              : ` di ${lookups.pocket.get(lines[0]?.pocketId ?? 0)?.name}.`}
            {total > CONFIRM_AMOUNT_ABOVE && (
              <span className="mt-2 block font-semibold">
                Nominal di atas Rp 10.000.000, pastikan angkanya benar.
              </span>
            )}
          </>
        }
        confirmLabel="Simpan"
        onConfirm={(reason) => {
          setConfirming(false)
          save(reason)
        }}
      />
    </Sheet>
  )
}

type LineProps = {
  line: Line
  type: CategoryType
  lookups: FinanceLookups
  pockets: FinanceLookups['pockets']
  onChange: (patch: Partial<Line>) => void
}

function PocketAndCategory({ line, type, lookups, pockets, onChange }: LineProps) {
  const pocket = lookups.pocket.get(line.pocketId)
  const categories = pocket ? formCategories(lookups.categories, pocket.kind, type) : []
  return (
    <>
      <Field label="Kas" hint={pocket?.purpose}>
        <Select
          options={pockets.map((p) => ({ value: String(p.id), label: p.name }))}
          value={line.pocketId ? String(line.pocketId) : undefined}
          onChange={(v) => {
            const next = lookups.pocket.get(Number(v))
            const keep =
              next &&
              formCategories(lookups.categories, next.kind, type).some(
                (c) => c.id === line.categoryId,
              )
            onChange({ pocketId: Number(v), categoryId: keep ? line.categoryId : null })
          }}
        />
      </Field>
      <Field label="Kategori">
        <Select
          placeholder="Pilih kategori"
          options={categories.map((c) => ({ value: String(c.id), label: c.name }))}
          value={line.categoryId ? String(line.categoryId) : undefined}
          onChange={(v) => onChange({ categoryId: Number(v) })}
        />
      </Field>
    </>
  )
}

function SingleLine(props: LineProps) {
  return <PocketAndCategory {...props} />
}

function SplitLine({
  index,
  canRemove,
  onRemove,
  ...props
}: LineProps & { index: number; canRemove: boolean; onRemove: () => void }) {
  return (
    <div className="border-line flex flex-col gap-2.5 rounded-lg border p-3">
      <div className="flex items-center justify-between">
        <p className="text-sm font-bold">Kas {index + 1}</p>
        {canRemove && (
          <IconButton label={`Hapus kas ${index + 1}`} icon={<Trash2 />} onClick={onRemove} />
        )}
      </div>
      <PocketAndCategory {...props} />
      <Field label="Nominal">
        <AmountInput
          value={props.line.amount}
          max={MAX_AMOUNT}
          onChange={(amount) => props.onChange({ amount })}
        />
      </Field>
    </div>
  )
}

function drawerTitle(draft: TransactionDraft) {
  if (draft.edit) return 'Ubah transaksi'
  if (draft.correctsId) return 'Transaksi koreksi'
  return 'Tambah transaksi'
}
