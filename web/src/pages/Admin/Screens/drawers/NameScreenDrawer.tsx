import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Button, Callout, ConfirmDialog, Field, Input, Sheet, Stack, useToast } from '@/components'
import { useUpdateScreen } from '@/hooks/useScreens'
import type { Screen } from '@/types/screen'
import { type ScreenNameForm, screenNameSchema } from '../schema'
import { currentMode } from '../utils'

type NameScreenDrawerProps = {
  screen: Screen
  now: Date
  onClose: () => void
}

// "Beri nama" for a new screen, "Ganti nama" for a named one. Saved right away after a confirm.
export function NameScreenDrawer({ screen, now, onClose }: NameScreenDrawerProps) {
  const update = useUpdateScreen()
  const toast = useToast()
  const isNew = screen.name === ''
  const form = useForm<ScreenNameForm>({
    resolver: zodResolver(screenNameSchema),
    // a new screen starts from its ?layar= hint
    defaultValues: { name: screen.name || screen.name_hint },
  })
  const { errors } = form.formState
  const [pending, setPending] = useState<ScreenNameForm | null>(null)

  const save = ({ name }: ScreenNameForm) =>
    update.mutate(
      // keep the mode in force now; a mode that already ended must not start again
      { id: screen.id, body: { name, forced_mode: currentMode(screen, now) } },
      {
        onSuccess: () => {
          toast({ message: 'Nama layar disimpan', description: name })
          onClose()
        },
        onError: (error) => form.setError('root', { message: error.message }),
      },
    )

  return (
    <Sheet
      open
      onClose={onClose}
      title={isNew ? 'Beri nama layar' : 'Ganti nama layar'}
      footer={
        <Button type="submit" form="screen-name" size="xl" block loading={update.isPending}>
          Simpan nama
        </Button>
      }
    >
      <form id="screen-name" onSubmit={form.handleSubmit(setPending)} noValidate>
        <Stack gap="sm">
          <Field label="Nama layar" error={errors.name?.message}>
            <Input placeholder="TV Ruang Utama" autoComplete="off" {...form.register('name')} />
          </Field>
          {errors.root?.message && <Callout tone="danger">{errors.root.message}</Callout>}
        </Stack>
      </form>
      <ConfirmDialog
        open={pending !== null}
        onClose={() => setPending(null)}
        title="Simpan nama layar?"
        description={pending && `Layar ini akan tampil sebagai "${pending.name}".`}
        confirmLabel="Simpan"
        onConfirm={() => pending && save(pending)}
      />
    </Sheet>
  )
}
