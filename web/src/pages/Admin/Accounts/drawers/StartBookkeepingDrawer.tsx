import { useState } from 'react'
import {
  AmountInput,
  Button,
  Callout,
  ConfirmDialog,
  Field,
  Input,
  SectionHeader,
  Sheet,
  Skeleton,
  Stack,
  useToast,
} from '@/components'
import { useAccounts, useFinanceMutation, useFinanceSummary, usePockets } from '@/hooks/useFinance'
import { startBookkeeping } from '@/services/financeService'
import { formatMonth, isOpen } from '@/utils/finance'
import { formatRupiah } from '@/utils/format/money'

type Amounts = Record<number, number | null>

// Mulai pembukuan, once: the month the books start and what each account and each kas holds at
// its start. Both add up to the same money, so their totals must match. It is never pemasukan.
export function StartBookkeepingDrawer({ onClose }: { onClose: () => void }) {
  const accounts = useAccounts()
  const pockets = usePockets()
  const summary = useFinanceSummary('')
  const [month, setMonth] = useState<string | null>(null)
  const [accountAmounts, setAccountAmounts] = useState<Amounts>({})
  const [pocketAmounts, setPocketAmounts] = useState<Amounts>({})
  const [error, setError] = useState<string | null>(null)
  const [confirming, setConfirming] = useState(false)
  const toast = useToast()
  const start = useFinanceMutation(startBookkeeping)

  if (!accounts.data || !pockets.data || !summary.data) {
    return (
      <Sheet open onClose={onClose} title="Mulai pembukuan">
        <Skeleton className="h-40" />
      </Sheet>
    )
  }
  const startMonth = month ?? summary.data.month
  const openAccounts = accounts.data.filter((a) => a.archived_at === null)
  const openPockets = pockets.data.filter(isOpen)
  const sum = (amounts: Amounts) => Object.values(amounts).reduce<number>((t, v) => t + (v ?? 0), 0)
  const accountTotal = sum(accountAmounts)
  const pocketTotal = sum(pocketAmounts)
  const balanced = accountTotal === pocketTotal

  const toList = (amounts: Amounts) =>
    Object.entries(amounts)
      .filter(([, v]) => v)
      .map(([id, amount]) => ({ id: Number(id), amount: amount ?? 0 }))

  return (
    <Sheet
      open
      onClose={onClose}
      title="Mulai pembukuan"
      footer={
        <Button
          size="xl"
          block
          loading={start.isPending}
          disabled={!balanced}
          onClick={() => setConfirming(true)}
        >
          Mulai pembukuan
        </Button>
      }
    >
      <Stack gap="md">
        <p className="text-ink-soft text-sm">
          Saldo awal hanya diisi sekali dan bukan pemasukan. Isi berapa uang di setiap akun pada
          awal bulan mulai, lalu bagi ke kas pemiliknya.
        </p>
        <Field label="Bulan mulai dibukukan">
          <Input
            type="month"
            max={summary.data.month}
            value={startMonth}
            onChange={(e) => setMonth(e.currentTarget.value)}
          />
        </Field>

        <SectionHeader title="Saldo awal per akun" />
        {openAccounts.map((a) => (
          <Field key={a.id} label={a.name}>
            <AmountInput
              value={accountAmounts[a.id] ?? null}
              onChange={(v) => setAccountAmounts((all) => ({ ...all, [a.id]: v }))}
            />
          </Field>
        ))}

        <SectionHeader title="Dibagi ke kas" />
        {openPockets.map((p) => (
          <Field key={p.id} label={p.name}>
            <AmountInput
              value={pocketAmounts[p.id] ?? null}
              onChange={(v) => setPocketAmounts((all) => ({ ...all, [p.id]: v }))}
            />
          </Field>
        ))}

        <Callout tone={balanced ? 'notice' : 'danger'}>
          Total akun {formatRupiah(accountTotal)} · total kas {formatRupiah(pocketTotal)}
          {!balanced &&
            `. Selisih ${formatRupiah(Math.abs(accountTotal - pocketTotal))}, samakan dulu.`}
        </Callout>
        {error && <Callout tone="danger">{error}</Callout>}
      </Stack>

      <ConfirmDialog
        open={confirming}
        onClose={() => setConfirming(false)}
        title="Mulai pembukuan?"
        description={`Pembukuan mulai ${formatMonth(startMonth)} dengan saldo awal ${formatRupiah(accountTotal)}. Saldo awal tidak bisa diubah setelah ini.`}
        confirmLabel="Mulai"
        onConfirm={() => {
          setConfirming(false)
          start.mutate(
            {
              start_month: startMonth,
              accounts: toList(accountAmounts),
              pockets: toList(pocketAmounts),
            },
            {
              onSuccess: () => {
                toast({ message: 'Pembukuan dimulai' })
                onClose()
              },
              onError: (e) => setError(e.message),
            },
          )
        }}
      />
    </Sheet>
  )
}
