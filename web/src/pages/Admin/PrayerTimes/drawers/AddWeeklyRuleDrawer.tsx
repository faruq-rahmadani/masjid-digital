import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { Button, Field, Input, Select, Sheet, Stack } from '@/components'
import type { WeeklyPrayerOverride } from '@/types/prayer'
import { type WeeklyRuleForm, weeklyRuleSchema } from '../schema'
import { PRAYER_OPTIONS, WEEKDAY_OPTIONS } from '../utils'

type AddWeeklyRuleDrawerProps = {
  /** Rules already in the draft: one per weekday and prayer. */
  rules: WeeklyPrayerOverride[]
  onAdd: (rule: WeeklyPrayerOverride) => void
  onClose: () => void
}

// Adds to the draft only; "Simpan dan kirim ke layar" saves it with the other settings.
export function AddWeeklyRuleDrawer({ rules, onAdd, onClose }: AddWeeklyRuleDrawerProps) {
  const form = useForm<WeeklyRuleForm>({
    resolver: zodResolver(weeklyRuleSchema),
    defaultValues: { weekday: 'friday', prayer: 'dzuhur', time: '', remarks: '' },
  })
  const { errors } = form.formState

  const submit = form.handleSubmit((values) => {
    if (rules.some((r) => r.weekday === values.weekday && r.prayer === values.prayer)) {
      form.setError('prayer', { message: 'Hari dan waktu sholat ini sudah punya jadwal mingguan' })
      return
    }
    onAdd(values)
    onClose()
  })

  return (
    <Sheet
      open
      onClose={onClose}
      title="Tambah jadwal mingguan"
      footer={
        <Button type="submit" form="add-weekly-rule" size="xl" block>
          Tambahkan
        </Button>
      }
    >
      <form id="add-weekly-rule" onSubmit={submit} noValidate>
        <Stack gap="sm">
          <Field label="Setiap hari">
            <Controller
              control={form.control}
              name="weekday"
              render={({ field }) => (
                <Select
                  options={WEEKDAY_OPTIONS}
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  name={field.name}
                />
              )}
            />
          </Field>
          <Field label="Waktu sholat" error={errors.prayer?.message}>
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
            <Input placeholder="Contoh: Sholat Jumat" {...form.register('remarks')} />
          </Field>
        </Stack>
      </form>
    </Sheet>
  )
}
