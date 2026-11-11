import { useState } from 'react'
import {
  Button,
  Callout,
  Checkbox,
  ConfirmDialog,
  Field,
  Select,
  Sheet,
  Stack,
  Textarea,
  useToast,
} from '@/components'
import { useAccounts, useFinanceMutation } from '@/hooks/useFinance'
import { finishPocket } from '@/services/financeService'
import type { Pocket } from '@/types/finance'
import { formatRupiah } from '@/utils/format/money'

type FinishPocketDrawerProps = {
  pocket: Pocket
  onClose: () => void
}

// "Selesaikan program" (E7-19): the remainder moves to Kas Umum as a Pindah dana, then the kas is
// archived. "Minta persetujuan donatur dulu" needs the consent ticked and a note.
export function FinishPocketDrawer({ pocket, onClose }: FinishPocketDrawerProps) {
  const accounts = useAccounts()
  const open = accounts.data?.filter((a) => a.archived_at === null) ?? []
  const [accountId, setAccountId] = useState<number | null>(null)
  const [consent, setConsent] = useState(false)
  const [note, setNote] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [confirming, setConfirming] = useState(false)
  const finish = useFinanceMutation(() =>
    finishPocket(pocket.id, {
      account_id: accountId ?? open[0]?.id ?? 0,
      donor_consent: consent,
      note,
    }),
  )
  const toast = useToast()
  const askDonor = pocket.surplus_plan === 'ask_donor'

  return (
    <Sheet
      open
      onClose={onClose}
      title="Selesaikan program"
      footer={
        <Button
          size="xl"
          block
          loading={finish.isPending}
          onClick={() => {
            if (askDonor && (!consent || !note.trim())) {
              return setError('Centang persetujuan donatur dan isi catatannya.')
            }
            setError(null)
            setConfirming(true)
          }}
        >
          Selesaikan program
        </Button>
      }
    >
      <Stack gap="md">
        <Callout tone="notice" title={pocket.name}>
          Sisa dana <b>{formatRupiah(pocket.balance)}</b> dipindah ke Kas Umum, lalu kas ini
          diarsipkan.
        </Callout>
        {pocket.balance > 0 && (
          <Field
            label="Uang sisa ada di akun"
            hint="Uangnya tidak berpindah akun, hanya berpindah kas"
          >
            <Select
              options={open.map((a) => ({ value: String(a.id), label: a.name }))}
              value={String(accountId ?? open[0]?.id ?? '')}
              onChange={(v) => setAccountId(Number(v))}
            />
          </Field>
        )}
        {askDonor && (
          <>
            <Checkbox
              label="Donatur sudah setuju sisa dana dialihkan ke Kas Umum"
              checked={consent}
              onChange={(e) => setConsent(e.currentTarget.checked)}
            />
            <Field label="Catatan persetujuan">
              <Textarea
                rows={2}
                maxLength={200}
                placeholder="Contoh: Disetujui donatur di rapat 2 Nov"
                value={note}
                onChange={(e) => setNote(e.currentTarget.value)}
              />
            </Field>
          </>
        )}
        {error && <Callout tone="danger">{error}</Callout>}
      </Stack>
      <ConfirmDialog
        open={confirming}
        onClose={() => setConfirming(false)}
        title={`Selesaikan ${pocket.name}?`}
        description={`${formatRupiah(pocket.balance)} masuk ke Kas Umum dan kas ini diarsipkan.`}
        confirmLabel="Selesaikan"
        onConfirm={() => {
          setConfirming(false)
          finish.mutate(undefined, {
            onSuccess: () => {
              toast({ message: `${pocket.name} selesai` })
              onClose()
            },
            onError: (e) => setError(e.message),
          })
        }}
      />
    </Sheet>
  )
}
