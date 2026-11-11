import { useState } from 'react'
import {
  AmountInput,
  Button,
  Callout,
  ConfirmDialog,
  Field,
  Sheet,
  Skeleton,
  Stack,
  useToast,
} from '@/components'
import { type FinanceLookups, useFinanceMutation, useFinanceSummary } from '@/hooks/useFinance'
import { closeMonth } from '@/services/financeService'
import { formatMonth } from '@/utils/finance'
import { formatRupiah } from '@/utils/format/money'

type CloseMonthDrawerProps = {
  month: string
  lookups: FinanceLookups
  onClose: () => void
}

// Tutup buku (E7-11): per account the recorded balance next to the counted one. They must match;
// a difference is recorded as a transaction first. Kas balances are kept with the closing.
export function CloseMonthDrawer({ month, lookups, onClose }: CloseMonthDrawerProps) {
  const summary = useFinanceSummary(month)
  const [counted, setCounted] = useState<Record<number, number | null>>({})
  const [error, setError] = useState<string | null>(null)
  const [confirming, setConfirming] = useState(false)
  const save = useFinanceMutation(closeMonth)
  const toast = useToast()

  const accounts =
    summary.data?.accounts.filter((a) => {
      const account = lookups.account.get(a.account_id)
      return account && (account.archived_at === null || a.closing !== 0)
    }) ?? []
  const mismatched = accounts.filter((a) => (counted[a.account_id] ?? null) !== a.closing)

  return (
    <Sheet
      open
      onClose={onClose}
      title={`Tutup buku ${formatMonth(month)}`}
      footer={
        <Button
          size="xl"
          block
          loading={save.isPending}
          disabled={!summary.data || mismatched.length > 0}
          onClick={() => setConfirming(true)}
        >
          Tutup buku
        </Button>
      }
    >
      {!summary.data ? (
        <Skeleton className="h-40" />
      ) : (
        <Stack gap="md">
          <p className="text-ink-soft text-sm">
            Hitung uang tunai dan cek saldo rekening per akhir {formatMonth(month)}, lalu isi di
            bawah.
          </p>
          {accounts.map((a) => {
            const value = counted[a.account_id] ?? null
            const diff = value === null ? null : value - a.closing
            return (
              <Field
                key={a.account_id}
                label={lookups.account.get(a.account_id)?.name}
                hint={`Catatan: ${formatRupiah(a.closing)}`}
                error={
                  diff
                    ? `Selisih ${formatRupiah(diff)}. Catat transaksi penyesuaian dulu.`
                    : undefined
                }
              >
                <AmountInput
                  value={value}
                  onChange={(v) => setCounted((all) => ({ ...all, [a.account_id]: v }))}
                />
              </Field>
            )
          })}

          <Stack gap="sm">
            <p className="text-ink-soft text-xs font-semibold">Saldo kas akhir bulan</p>
            {summary.data.pockets
              .filter((p) => p.closing !== 0 || p.in !== 0 || p.out !== 0)
              .map((p) => (
                <p key={p.pocket_id} className="flex justify-between text-sm">
                  <span>{lookups.pocket.get(p.pocket_id)?.name}</span>
                  <b>{formatRupiah(p.closing)}</b>
                </p>
              ))}
          </Stack>
          <p className="text-ink-soft text-xs">
            Pemasukan {formatRupiah(summary.data.total_in)} · pengeluaran{' '}
            {formatRupiah(summary.data.total_out)}
          </p>
          {error && <Callout tone="danger">{error}</Callout>}
        </Stack>
      )}

      <ConfirmDialog
        open={confirming}
        onClose={() => setConfirming(false)}
        title={`Tutup buku ${formatMonth(month)}?`}
        description="Transaksi bulan ini di semua kas akan terkunci. Kesalahan setelahnya diperbaiki dengan transaksi koreksi."
        confirmLabel="Tutup buku"
        danger
        onConfirm={() => {
          setConfirming(false)
          save.mutate(
            {
              month,
              accounts: accounts.map((a) => ({
                account_id: a.account_id,
                counted: counted[a.account_id] ?? 0,
              })),
            },
            {
              onSuccess: () => {
                toast({ message: `${formatMonth(month)} sudah tutup buku` })
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
