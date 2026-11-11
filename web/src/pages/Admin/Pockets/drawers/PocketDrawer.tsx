import { useState } from 'react'
import {
  AmountInput,
  Button,
  Callout,
  Checkbox,
  ConfirmDialog,
  Field,
  Input,
  RadioCard,
  Sheet,
  Stack,
  Textarea,
  useToast,
} from '@/components'
import { useFinanceMutation } from '@/hooks/useFinance'
import {
  archivePocket,
  createPocket,
  type PocketFields,
  restorePocket,
  updatePocket,
} from '@/services/financeService'
import type { Pocket, SurplusPlan } from '@/types/finance'
import { isBuiltIn, isOpen, MAX_AMOUNT } from '@/utils/finance'
import { formatRupiah } from '@/utils/format/money'

type PocketDrawerProps = {
  /** null: Kas baru */
  pocket: Pocket | null
  canEdit: boolean
  onFinish: (pocket: Pocket) => void
  onClose: () => void
}

type Pending = 'save' | 'archive' | 'restore'

// Kas baru / Ubah kas. Users add terikat kas only; a built-in kas is locked except "tampil di TV".
// The type never changes; a kas is archived only at Rp 0.
export function PocketDrawer({ pocket, canEdit, onFinish, onClose }: PocketDrawerProps) {
  const builtIn = pocket ? isBuiltIn(pocket.kind) : false
  const archived = pocket ? !isOpen(pocket) : false
  const readOnly = !canEdit || archived
  const [fields, setFields] = useState<PocketFields>({
    name: pocket?.name ?? '',
    purpose: pocket?.purpose ?? '',
    target_amount: pocket?.target_amount ?? null,
    surplus_plan: pocket?.surplus_plan ?? 'to_umum',
    show_on_tv: pocket?.show_on_tv ?? false,
  })
  const [pending, setPending] = useState<Pending | null>(null)
  const [error, setError] = useState<string | null>(null)
  const toast = useToast()
  const set = (patch: Partial<PocketFields>) => setFields((f) => ({ ...f, ...patch }))

  const save = useFinanceMutation((body: PocketFields) =>
    pocket ? updatePocket(pocket.id, body) : createPocket(body),
  )
  const archive = useFinanceMutation(archivePocket)
  const restore = useFinanceMutation(restorePocket)

  const validate = () => {
    if (builtIn) return null
    if (!fields.name.trim()) return 'Isi nama kas.'
    if (!fields.purpose.trim()) return 'Isi peruntukan kas.'
    return null
  }

  const done = (message: string) => ({
    onSuccess: () => {
      toast({ message })
      onClose()
    },
    onError: (e: Error) => setError(e.message),
  })

  const run = () => {
    setPending(null)
    if (pending === 'archive' && pocket) return archive.mutate(pocket.id, done('Kas diarsipkan'))
    if (pending === 'restore' && pocket) return restore.mutate(pocket.id, done('Kas dipulihkan'))
    save.mutate(fields, done(pocket ? 'Kas disimpan' : 'Kas dibuat'))
  }

  return (
    <Sheet
      open
      onClose={onClose}
      title={pocket ? pocket.name : 'Kas baru'}
      footer={
        !readOnly && (
          <Button
            size="xl"
            block
            loading={save.isPending}
            onClick={() => {
              const problem = validate()
              setError(problem)
              if (!problem) setPending('save')
            }}
          >
            {pocket ? 'Simpan' : 'Buat kas'}
          </Button>
        )
      }
    >
      <Stack gap="md">
        {builtIn && (
          <Callout tone="notice">
            Kas bawaan beserta aturannya tidak bisa diubah. Yang bisa diatur hanya tampil di layar
            TV.
          </Callout>
        )}
        {!pocket && (
          <p className="text-ink-soft text-sm">
            Kas terikat menampung donasi untuk satu tujuan. Pindah ke kas lain butuh kerelaan
            donatur.
          </p>
        )}
        {pocket && <p className="text-2xl font-extrabold">{formatRupiah(pocket.balance)}</p>}

        {!builtIn && (
          <>
            <Field label="Nama kas">
              <Input
                placeholder="Renovasi Atap"
                maxLength={60}
                disabled={readOnly}
                value={fields.name}
                onChange={(e) => set({ name: e.currentTarget.value })}
              />
            </Field>
            <Field label="Peruntukan">
              <Textarea
                rows={2}
                maxLength={200}
                placeholder="Donasi khusus penggantian atap dan talang masjid"
                disabled={readOnly}
                value={fields.purpose}
                onChange={(e) => set({ purpose: e.currentTarget.value })}
              />
            </Field>
            <Field label="Target" hint="Opsional. Kosongkan bila tidak ada target">
              <AmountInput
                value={fields.target_amount}
                max={MAX_AMOUNT}
                disabled={readOnly}
                onChange={(target_amount) => set({ target_amount })}
              />
            </Field>
            <fieldset className="flex flex-col gap-2">
              <legend className="text-ink-soft mb-1 text-xs font-semibold">
                Rencana sisa dana setelah program selesai
              </legend>
              {(
                [
                  ['to_umum', 'Dialihkan ke Kas Umum', 'Disampaikan ke donatur sejak awal.'],
                  [
                    'ask_donor',
                    'Minta persetujuan donatur dulu',
                    'Sisa dana baru dipindah setelah donatur setuju.',
                  ],
                ] as const satisfies readonly [SurplusPlan, string, string][]
              ).map(([value, title, description]) => (
                <RadioCard
                  key={value}
                  name="surplus_plan"
                  title={title}
                  description={description}
                  checked={fields.surplus_plan === value}
                  disabled={readOnly}
                  onChange={() => set({ surplus_plan: value })}
                />
              ))}
            </fieldset>
          </>
        )}

        <Checkbox
          label="Tampil di layar TV"
          checked={fields.show_on_tv}
          disabled={readOnly}
          onChange={(e) => set({ show_on_tv: e.currentTarget.checked })}
        />
        <p className="text-ink-soft -mt-2 text-xs">
          Ringkasan kas ini tampil bergiliran di header TV.
        </p>

        {pocket && canEdit && !builtIn && !archived && (
          <Stack gap="sm">
            <Button variant="outline" size="lg" block onClick={() => onFinish(pocket)}>
              Selesaikan program
            </Button>
            {pocket.balance === 0 ? (
              <Button variant="danger" size="lg" block onClick={() => setPending('archive')}>
                Arsipkan kas
              </Button>
            ) : (
              <p className="text-ink-soft text-xs">
                Kas bisa diarsipkan setelah saldonya Rp 0. Saldo sekarang{' '}
                {formatRupiah(pocket.balance)}.
              </p>
            )}
          </Stack>
        )}
        {pocket && canEdit && archived && (
          <Button variant="secondary" size="lg" block onClick={() => setPending('restore')}>
            Pulihkan dari arsip
          </Button>
        )}
        {error && <Callout tone="danger">{error}</Callout>}
      </Stack>

      <ConfirmDialog
        open={pending !== null}
        onClose={() => setPending(null)}
        title={confirmTitle(pending, pocket)}
        confirmLabel={pending === 'archive' ? 'Arsipkan' : 'Simpan'}
        danger={pending === 'archive'}
        onConfirm={run}
      />
    </Sheet>
  )
}

function confirmTitle(pending: Pending | null, pocket: Pocket | null) {
  if (pending === 'archive') return `Arsipkan ${pocket?.name}?`
  if (pending === 'restore') return `Pulihkan ${pocket?.name}?`
  if (pocket) return 'Simpan perubahan kas?'
  return 'Buat kas baru?'
}
