import { useState } from 'react'
import { Button, RadioCard, Sheet, Stack, useToast } from '@/components'
import { useUpdateUser } from '@/hooks/useUsers'
import type { Role } from '@/types/auth'
import type { User } from '@/types/user'
import { ROLES, roleLabels } from '@/utils/auth/roles'

const roleDescriptions: Record<Role, string> = {
  super_admin: 'Semua menu, termasuk pengguna dan profil masjid',
  takmir: 'Waktu sholat, iqamah, konten dan layar TV',
  bendahara: 'Keuangan dan kategori',
}

type ChangeRoleDrawerProps = {
  user: User
  onClose: () => void
}

export function ChangeRoleDrawer({ user, onClose }: ChangeRoleDrawerProps) {
  const [role, setRole] = useState<Role>(user.role)
  const update = useUpdateUser()
  const toast = useToast()

  const save = () =>
    update.mutate(
      { user, changes: { role } },
      {
        onSuccess: () => {
          toast({
            message: `Peran ${user.name} diubah`,
            description: `Sekarang ${roleLabels[role]}. Sesi login-nya diakhiri.`,
          })
          onClose()
        },
        onError: (error) => toast({ tone: 'danger', message: error.message }),
      },
    )

  return (
    <Sheet
      open
      onClose={onClose}
      title={`Ubah peran ${user.name}`}
      footer={
        <Button
          size="xl"
          block
          loading={update.isPending}
          disabled={role === user.role}
          onClick={save}
        >
          Simpan peran
        </Button>
      }
    >
      <fieldset className="m-0 min-w-0 border-0 p-0">
        <legend className="sr-only">Peran</legend>
        <Stack gap="sm">
          {ROLES.map((value) => (
            <RadioCard
              key={value}
              name="role"
              value={value}
              checked={role === value}
              onChange={() => setRole(value)}
              title={roleLabels[value]}
              description={roleDescriptions[value]}
            />
          ))}
        </Stack>
      </fieldset>
    </Sheet>
  )
}
