import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { ApiError } from '@/api/client'
import { Button, Callout, Field, Input, Sheet, Stack, useToast } from '@/components'
import { useUpdateUser } from '@/hooks/useUsers'
import type { User } from '@/types/user'
import { formatPhone, normalizePhone } from '@/utils/format/phone'
import { type UserDetailsForm, userDetailsSchema } from '../schema'

type EditUserDrawerProps = {
  user: User
  onClose: () => void
}

export function EditUserDrawer({ user, onClose }: EditUserDrawerProps) {
  const update = useUpdateUser()
  const toast = useToast()
  const form = useForm<UserDetailsForm>({
    resolver: zodResolver(userDetailsSchema),
    defaultValues: { name: user.name, phone: formatPhone(user.phone), email: user.email },
  })
  const { errors, isDirty } = form.formState

  const submit = form.handleSubmit((values) =>
    update.mutate(
      {
        user,
        changes: {
          name: values.name,
          phone: normalizePhone(values.phone) ?? values.phone,
          email: values.email,
        },
      },
      {
        onSuccess: () => {
          toast({ message: `Data ${values.name} disimpan` })
          onClose()
        },
        onError: (error) => {
          if (error instanceof ApiError && error.code === 'ALREADY_EXISTS') {
            form.setError('phone', { message: error.message })
          } else {
            form.setError('root', { message: error.message })
          }
        },
      },
    ),
  )

  return (
    <Sheet
      open
      onClose={onClose}
      title={`Ubah data ${user.name}`}
      footer={
        <Button
          type="submit"
          form="edit-user"
          size="xl"
          block
          loading={update.isPending}
          disabled={!isDirty}
        >
          Simpan perubahan
        </Button>
      }
    >
      <form id="edit-user" onSubmit={submit} noValidate>
        <Stack gap="sm">
          <Field label="Nama" error={errors.name?.message}>
            <Input autoComplete="off" {...form.register('name')} />
          </Field>
          <Field label="Nomor HP" hint="Untuk login" error={errors.phone?.message}>
            <Input type="tel" inputMode="tel" autoComplete="off" {...form.register('phone')} />
          </Field>
          <Field
            label="Email (opsional)"
            hint="Kosongkan untuk menghapus"
            error={errors.email?.message}
          >
            <Input type="email" inputMode="email" autoComplete="off" {...form.register('email')} />
          </Field>
          {errors.root?.message && <Callout tone="danger">{errors.root.message}</Callout>}
        </Stack>
      </form>
    </Sheet>
  )
}
