import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { ApiError } from '@/api/client'
import { Button, Callout, Field, PasswordInput, Sheet, Stack, useToast } from '@/components'
import { useChangePassword } from '@/hooks/useProfile'
import { passwordSchema } from '@/utils/auth/password'

const schema = z
  .object({
    current: z.string().min(1, 'Password lama wajib diisi'),
    next: passwordSchema,
  })
  .refine((v) => v.next !== v.current, {
    path: ['next'],
    message: 'Password baru harus berbeda dari password lama',
  })

// Any role changes their own password (PUT /auth/password). The session stays logged in.
export function ChangePasswordDrawer({ onClose }: { onClose: () => void }) {
  const change = useChangePassword()
  const toast = useToast()
  const form = useForm({ resolver: zodResolver(schema), defaultValues: { current: '', next: '' } })
  const { errors } = form.formState

  const submit = form.handleSubmit(({ current, next }) =>
    change.mutate(
      { current, next },
      {
        onSuccess: () => {
          toast({ message: 'Password diganti' })
          onClose()
        },
        onError: (error) => {
          // The API says which field is wrong only in its message.
          if (error instanceof ApiError && /password lama/i.test(error.message)) {
            form.setError('current', { message: error.message })
          } else if (error instanceof ApiError && error.code === 'INVALID_INPUT') {
            form.setError('next', { message: error.message })
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
      title="Ganti password"
      footer={
        <Button type="submit" form="change-password" size="xl" block loading={change.isPending}>
          Simpan password
        </Button>
      }
    >
      <form id="change-password" onSubmit={submit} noValidate>
        <Stack gap="sm">
          <Field label="Password lama" error={errors.current?.message}>
            <PasswordInput autoComplete="current-password" {...form.register('current')} />
          </Field>
          <Field label="Password baru" hint="Minimal 8 karakter" error={errors.next?.message}>
            <PasswordInput autoComplete="new-password" {...form.register('next')} />
          </Field>
          {errors.root?.message && <Callout tone="danger">{errors.root.message}</Callout>}
        </Stack>
      </form>
    </Sheet>
  )
}
