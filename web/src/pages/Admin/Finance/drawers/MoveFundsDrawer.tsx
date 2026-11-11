import { useState } from 'react'
import {
  Button,
  Callout,
  Checkbox,
  ConfirmDialog,
  Field,
  Input,
  MoneyInput,
  Select,
  Sheet,
  Stack,
  Textarea,
  useToast,
} from '@/components'
import { type FinanceLookups, useFinanceMutation } from '@/hooks/useFinance'
import { moveFunds, updateMoveFunds } from '@/services/financeService'
import type { MoveFundsRequest, Transaction } from '@/types/finance'
import { addMonths, canMoveFunds, isOpen, MAX_AMOUNT } from '@/utils/finance'
import { formatRupiah } from '@/utils/format/money'

type MoveFundsDrawerProps = {
  lookups: FinanceLookups
  today: string
  fromPocketId?: number
  /** The out and in rows of the pair being edited */
  edit?: Transaction[]
  onClose: () => void
}

// Pindah dana (E7-17): one form for moving between kas and between accounts (setor/tarik tunai
// is the same kas on both sides). Saved as an out + in pair; never into or out of zakat or
// fidyah; out of a terikat kas only with a reason and the donors' consent.
export function MoveFundsDrawer({
  lookups,
  today,
  fromPocketId,
  edit,
  onClose,
}: MoveFundsDrawerProps) {
  const out = edit?.find((t) => t.kind === 'out')
  const into = edit?.find((t) => t.kind === 'in')
  const pockets = lookups.pockets.filter(isOpen)
  const accounts = lookups.accounts.filter((a) => a.archived_at === null)
  const cash = accounts.find((a) => a.kind === 'cash') ?? accounts[0]

  const [fromPocket, setFromPocket] = useState(
    out?.pocket_id ?? fromPocketId ?? pockets[0]?.id ?? 0,
  )
  const [fromAccount, setFromAccount] = useState(out?.account_id ?? cash?.id ?? 0)
  const [toPocket, setToPocket] = useState<number | null>(into?.pocket_id ?? null)
  const [toAccount, setToAccount] = useState(into?.account_id ?? cash?.id ?? 0)
  const [amount, setAmount] = useState<number | null>(out?.amount ?? null)
  const [date, setDate] = useState(out?.date ?? today)
  const [note, setNote] = useState(out?.note ?? '')
  const [consent, setConsent] = useState(out?.donor_consent ?? false)
  const [error, setError] = useState<string | null>(null)
  const [confirming, setConfirming] = useState(false)
  const toast = useToast()

  const from = lookups.pocket.get(fromPocket)
  const to = toPocket === null ? undefined : lookups.pocket.get(toPocket)
  const between = Boolean(from && to && from.id !== to.id)
  const needsConsent = between && from?.kind === 'terikat'
  // a zakat kas only moves within itself (setor/tarik tunai)
  const destinations = pockets.filter(
    (p) => p.id === fromPocket || (from && canMoveFunds(from.kind) && canMoveFunds(p.kind)),
  )

  const save = useFinanceMutation((body: MoveFundsRequest) =>
    out?.group_id ? updateMoveFunds(out.group_id, body) : moveFunds(body),
  )

  const validate = () => {
    if (!to) return 'Pilih kas tujuan.'
    if (!amount) return 'Isi nominal.'
    if (to.id === from?.id && toAccount === fromAccount) return 'Pilih akun tujuan yang berbeda.'
    if (needsConsent && (!consent || !note.trim()))
      return 'Isi alasan dan centang persetujuan donatur.'
    return null
  }

  const submit = () => {
    const problem = validate()
    setError(problem)
    if (!problem) setConfirming(true)
  }

  const run = (reason: string) => {
    setConfirming(false)
    save.mutate(
      {
        from_pocket_id: fromPocket,
        from_account_id: fromAccount,
        to_pocket_id: toPocket ?? 0,
        to_account_id: toAccount,
        amount: amount ?? 0,
        date,
        note,
        donor_consent: needsConsent && consent,
        reason,
      },
      {
        onSuccess: () => {
          toast({ message: edit ? 'Pindah dana diubah' : 'Dana dipindahkan' })
          onClose()
        },
        onError: (e) => setError(e.message),
      },
    )
  }

  const accountOptions = accounts.map((a) => ({
    value: String(a.id),
    label: `${a.name} · ${formatRupiah(a.balance)}`,
  }))
  const openFrom = lookups.lastClosed ? `${addMonths(lookups.lastClosed, 1)}-01` : undefined

  return (
    <Sheet
      open
      onClose={onClose}
      title={edit ? 'Ubah pindah dana' : 'Pindah dana'}
      footer={
        <Button size="xl" block loading={save.isPending} onClick={submit}>
          {edit ? 'Simpan perubahan' : 'Pindahkan dana'}
        </Button>
      }
    >
      <Stack gap="md">
        <Callout tone="notice">
          Pindah dana bukan pemasukan atau pengeluaran masjid. Uangnya hanya berpindah kas atau
          akun.
        </Callout>

        <MoneyInput value={amount} onChange={setAmount} flow="out" max={MAX_AMOUNT} />

        <Field label="Dari kas" hint={from && `Saldo ${formatRupiah(from.balance)}`}>
          <Select
            options={pockets.map((p) => ({ value: String(p.id), label: p.name }))}
            value={String(fromPocket)}
            onChange={(v) => {
              setFromPocket(Number(v))
              setToPocket(null)
            }}
          />
        </Field>
        <Field label="Dari akun">
          <Select
            options={accountOptions}
            value={String(fromAccount)}
            onChange={(v) => setFromAccount(Number(v))}
          />
        </Field>

        {from && !canMoveFunds(from.kind) && (
          <Callout tone="notice">
            Dana zakat dan fidyah hanya untuk mustahiknya, jadi tidak bisa dipindah ke kas lain. Di
            sini hanya bisa setor atau tarik tunai di kas yang sama.
          </Callout>
        )}

        <Field label="Ke kas" hint={to && `Saldo ${formatRupiah(to.balance)}`}>
          <Select
            placeholder="Pilih kas tujuan"
            options={destinations.map((p) => ({
              value: String(p.id),
              label: p.id === fromPocket ? `${p.name} (setor/tarik tunai)` : p.name,
            }))}
            value={toPocket === null ? undefined : String(toPocket)}
            onChange={(v) => setToPocket(Number(v))}
          />
        </Field>
        <Field label="Ke akun">
          <Select
            options={accountOptions}
            value={String(toAccount)}
            onChange={(v) => setToAccount(Number(v))}
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
        <Field label="Alasan" hint={needsConsent ? 'Wajib untuk dana kas terikat' : undefined}>
          <Textarea
            rows={2}
            maxLength={200}
            placeholder="Contoh: Tambahan dana kegiatan, keputusan rapat takmir"
            value={note}
            onChange={(e) => setNote(e.currentTarget.value)}
          />
        </Field>
        {needsConsent && (
          <Checkbox
            label="Donatur sudah merelakan dana ini dialihkan"
            checked={consent}
            onChange={(e) => setConsent(e.currentTarget.checked)}
          />
        )}
        {error && <Callout tone="danger">{error}</Callout>}
      </Stack>

      <ConfirmDialog
        open={confirming}
        onClose={() => setConfirming(false)}
        title={edit ? 'Simpan perubahan pindah dana?' : 'Pindahkan dana?'}
        requireReason={Boolean(edit)}
        description={
          <>
            {formatRupiah(amount ?? 0)} dari {from?.name} ({lookups.account.get(fromAccount)?.name})
            ke {to?.name} ({lookups.account.get(toAccount)?.name}).
          </>
        }
        confirmLabel={edit ? 'Simpan' : 'Pindahkan'}
        onConfirm={run}
      />
    </Sheet>
  )
}
