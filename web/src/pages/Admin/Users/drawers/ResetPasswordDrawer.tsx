import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { Button, Checkbox, Field, Input, Sheet, Stack, useToast } from '@/components'
import { useResetPassword } from '@/hooks/useUsers'
import type { User } from '@/types/user'
import { passwordSchema } from '../schema'

const schema = z.object({ password: passwordSchema })

type ResetPasswordDrawerProps = {
  user: User
  onClose: () => void
}

// The API lets a super admin set a new password directly; there is no reset link yet.
export function ResetPasswordDrawer({ user, onClose }: ResetPasswordDrawerProps) {
  const [visible, setVisible] = useState(false)
  const reset = useResetPassword()
  const toast = useToast()
  const form = useForm({ resolver: zodResolver(schema), defaultValues: { password: '' } })

  const submit = form.handleSubmit(({ password }) =>
    reset.mutate(
      { id: user.id, password },
      {
        onSuccess: () => {
          toast({
            message: `Password ${user.name} diatur ulang`,
            description: 'Berikan password baru secara langsung.',
          })
          onClose()
        },
        onError: (error) => form.setError('password', { message: error.message }),
      },
    ),
  )

  return (
    <Sheet
      open
      onClose={onClose}
      title="Atur ulang password"
      footer={
        <Button size="xl" block type="submit" form="reset-password" loading={reset.isPending}>
          Simpan password
        </Button>
      }
    >
      <form id="reset-password" onSubmit={submit} noValidate>
        <Stack gap="sm">
          <p className="text-ink-muted">
            {user.name} akan keluar dari semua perangkat dan masuk lagi dengan password ini.
          </p>
          <Field
            label="Password baru"
            hint="Minimal 8 karakter"
            error={form.formState.errors.password?.message}
          >
            <Input
              type={visible ? 'text' : 'password'}
              autoComplete="new-password"
              {...form.register('password')}
            />
          </Field>
          <Checkbox
            label="Tampilkan password"
            checked={visible}
            onChange={(event) => setVisible(event.currentTarget.checked)}
          />
        </Stack>
      </form>
    </Sheet>
  )
}
