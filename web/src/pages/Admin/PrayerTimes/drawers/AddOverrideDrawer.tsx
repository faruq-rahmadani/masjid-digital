import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { ApiError } from '@/api/client'
import {
  Button,
  Callout,
  ConfirmDialog,
  Field,
  Input,
  Select,
  Sheet,
  Stack,
  useToast,
} from '@/components'
import { useCreatePrayerOverride } from '@/hooks/usePrayerTimes'
import { formatLongDay } from '@/utils/format/day'
import { SCHEDULE_NAMES } from '@/utils/prayer/calculate'
import { type OverrideForm, overrideSchema } from '../schema'
import { PRAYER_OPTIONS } from '../utils'

type AddOverrideDrawerProps = {
  /** The mosque's today, the earliest date that can be picked. */
  today: string
  onClose: () => void
}

export function AddOverrideDrawer({ today, onClose }: AddOverrideDrawerProps) {
  const create = useCreatePrayerOverride()
  const toast = useToast()
  const form = useForm<OverrideForm>({
    resolver: zodResolver(overrideSchema),
    defaultValues: { date: today, prayer: 'dzuhur', time: '', remarks: '' },
  })
  const { errors } = form.formState

  // Saved right away (no Simpan button), so the form asks once more before creating it
  const [pending, setPending] = useState<OverrideForm | null>(null)
  const submit = form.handleSubmit(setPending)

  const createOverride = (values: OverrideForm) =>
    create.mutate(values, {
      onSuccess: (saved) => {
        toast({
          message: 'Jadwal khusus ditambahkan',
          description: `${formatLongDay(saved.date)}: ${SCHEDULE_NAMES[saved.prayer]} ${saved.time}`,
        })
        onClose()
      },
      onError: (error) => {
        // a duplicate is about the date and prayer picked
        if (error instanceof ApiError && error.code === 'ALREADY_EXISTS') {
          form.setError('date', { message: error.message })
          return
        }
        form.setError('root', { message: error.message })
      },
    })

  return (
    <Sheet
      open
      onClose={onClose}
      title="Tambah jadwal khusus"
      footer={
        <Button type="submit" form="add-override" size="xl" block loading={create.isPending}>
          Simpan jadwal khusus
        </Button>
      }
    >
      <form id="add-override" onSubmit={submit} noValidate>
        <Stack gap="sm">
          <Field label="Tanggal" error={errors.date?.message}>
            <Input type="date" min={today} {...form.register('date')} />
          </Field>
          <Field label="Waktu sholat">
            <Controller
              control={form.control}
              name="prayer"
              render={({ field }) => (
                <Select
                  options={PRAYER_OPTIONS}
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  name={field.name}
                />
              )}
            />
          </Field>
          <Field label="Jam" error={errors.time?.message}>
            <Input type="time" {...form.register('time')} />
          </Field>
          <Field label="Keterangan (opsional)" error={errors.remarks?.message}>
            <Input placeholder="Contoh: khatib tamu" {...form.register('remarks')} />
          </Field>
          {errors.root?.message && <Callout tone="danger">{errors.root.message}</Callout>}
        </Stack>
      </form>
      <ConfirmDialog
        open={pending !== null}
        onClose={() => setPending(null)}
        title="Simpan jadwal khusus?"
        description={
          pending &&
          `${SCHEDULE_NAMES[pending.prayer]} ${formatLongDay(pending.date)} jam ${pending.time} langsung tampil di layar TV.`
        }
        confirmLabel="Simpan"
        onConfirm={() => pending && createOverride(pending)}
      />
    </Sheet>
  )
}
