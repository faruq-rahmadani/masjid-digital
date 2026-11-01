import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { ApiError } from '@/api/client'
import {
  Button,
  Callout,
  Field,
  Input,
  PasswordInput,
  Select,
  Sheet,
  Stack,
  useToast,
} from '@/components'
import { useCreateUser } from '@/hooks/useUsers'
import { ROLES, roleLabels } from '@/utils/auth/roles'
import { normalizePhone } from '@/utils/format/phone'
import { type CreateUserForm, createUserSchema } from '../schema'

const EMPTY: CreateUserForm = { name: '', phone: '', email: '', role: 'takmir', password: '' }

export function AddUserDrawer({ onClose }: { onClose: () => void }) {
  const create = useCreateUser()
  const toast = useToast()
  const form = useForm({ resolver: zodResolver(createUserSchema), defaultValues: EMPTY })
  const { errors } = form.formState

  const submit = form.handleSubmit((values) =>
    create.mutate(
      { ...values, phone: normalizePhone(values.phone) ?? values.phone },
      {
        onSuccess: () => {
          toast({
            message: `${values.name} ditambahkan`,
            description: `Peran ${roleLabels[values.role]}`,
          })
          onClose()
        },
        onError: (error) => {
          // A taken phone or email belongs next to those fields; anything else above the button.
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
      title="Tambah pengguna"
      footer={
        <Button type="submit" form="add-user" size="xl" block loading={create.isPending}>
          Simpan pengguna
        </Button>
      }
    >
      <form id="add-user" onSubmit={submit} noValidate>
        <Stack gap="sm">
          <Field label="Nama" error={errors.name?.message}>
            <Input placeholder="Nama lengkap" autoComplete="off" {...form.register('name')} />
          </Field>
          <Field label="Nomor HP" hint="Untuk login" error={errors.phone?.message}>
            <Input
              type="tel"
              inputMode="tel"
              placeholder="0812-3456-7890"
              autoComplete="off"
              {...form.register('phone')}
            />
          </Field>
          <Field label="Email (opsional)" error={errors.email?.message}>
            <Input type="email" inputMode="email" autoComplete="off" {...form.register('email')} />
          </Field>
          <Field label="Peran">
            <Controller
              control={form.control}
              name="role"
              render={({ field }) => (
                <Select
                  options={ROLES.map((role) => ({ value: role, label: roleLabels[role] }))}
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  name={field.name}
                />
              )}
            />
          </Field>
          <Field
            label="Password awal"
            hint="Minimal 8 karakter. Berikan ke pengguna secara langsung."
            error={errors.password?.message}
          >
            <PasswordInput autoComplete="new-password" {...form.register('password')} />
          </Field>
          {errors.root?.message && <Callout tone="danger">{errors.root.message}</Callout>}
        </Stack>
      </form>
    </Sheet>
  )
}
