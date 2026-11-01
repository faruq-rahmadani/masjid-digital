import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Button, Callout, ConfirmDialog, Field, Input, Sheet, Stack, useToast } from '@/components'
import { useSaveFridayDuty } from '@/hooks/useFridayDuties'
import type { FridayDuty } from '@/types/fridayDuty'
import { formatLongDay } from '@/utils/format/day'
import { type FridayDutyForm, fridayDutySchema } from '../schema'

type FridayDutyDrawerProps = {
  date: string
  /** The petugas already set for that Friday, to edit. */
  duty: FridayDuty | undefined
  onClose: () => void
}

export function FridayDutyDrawer({ date, duty, onClose }: FridayDutyDrawerProps) {
  const save = useSaveFridayDuty()
  const toast = useToast()
  const day = formatLongDay(date)
  const form = useForm<FridayDutyForm>({
    resolver: zodResolver(fridayDutySchema),
    defaultValues: {
      khatib: duty?.khatib ?? '',
      imam: duty?.imam ?? '',
      muadzin: duty?.muadzin ?? '',
      theme: duty?.theme ?? '',
    },
  })
  const { errors } = form.formState

  // Saved right away (no Simpan button), so the form asks once more before saving
  const [pending, setPending] = useState<FridayDutyForm | null>(null)
  const submit = form.handleSubmit(setPending)

  const saveDuty = (body: FridayDutyForm) =>
    save.mutate(
      { date, body },
      {
        onSuccess: () => {
          toast({ message: 'Petugas Jumat disimpan', description: day })
          onClose()
        },
        onError: (error) => form.setError('root', { message: error.message }),
      },
    )

  return (
    <Sheet
      open
      onClose={onClose}
      title={`Petugas ${day}`}
      footer={
        <Button type="submit" form="friday-duty" size="xl" block loading={save.isPending}>
          Simpan petugas
        </Button>
      }
    >
      <form id="friday-duty" onSubmit={submit} noValidate>
        <Stack gap="sm">
          <Field label="Khatib" error={errors.khatib?.message}>
            <Input placeholder="Ustadz ..." autoComplete="off" {...form.register('khatib')} />
          </Field>
          <Field label="Imam" error={errors.imam?.message}>
            <Input autoComplete="off" {...form.register('imam')} />
          </Field>
          <Field label="Muadzin" error={errors.muadzin?.message}>
            <Input autoComplete="off" {...form.register('muadzin')} />
          </Field>
          <Field label="Tema khutbah (opsional)" error={errors.theme?.message}>
            <Input autoComplete="off" {...form.register('theme')} />
          </Field>
          {errors.root?.message && <Callout tone="danger">{errors.root.message}</Callout>}
        </Stack>
      </form>
      <ConfirmDialog
        open={pending !== null}
        onClose={() => setPending(null)}
        title="Simpan petugas Jumat?"
        description={`Petugas ${day} langsung tampil di layar Jumat.`}
        confirmLabel="Simpan"
        onConfirm={() => pending && saveDuty(pending)}
      />
    </Sheet>
  )
}
