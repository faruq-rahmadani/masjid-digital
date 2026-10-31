import { LogOut } from 'lucide-react'
import { useState } from 'react'
import { Avatar, Button, Sheet, Stack } from '@/components'
import { useAuth } from '@/stores/sessionStore'
import type { Session } from '@/types/auth'
import { roleLabels, roleTones } from '@/utils/auth/roles'

type AccountDrawerProps = {
  session: Session
  name: string | undefined
  onClose: () => void
}

export function AccountDrawer({ session, name, onClose }: AccountDrawerProps) {
  const { logout } = useAuth()
  const [leaving, setLeaving] = useState(false)
  const role = roleLabels[session.role]

  return (
    <Sheet open onClose={onClose} title="Akun saya">
      <Stack gap="lg">
        <div className="bg-card flex items-center gap-3 rounded-xl p-4">
          <Avatar name={name ?? role} tone={roleTones[session.role]} size="lg" />
          <div className="min-w-0">
            {name && <p className="truncate font-bold">{name}</p>}
            <p className="text-ink-muted text-xs">{role}</p>
          </div>
        </div>
        <Button
          variant="danger"
          size="lg"
          block
          loading={leaving}
          onClick={async () => {
            setLeaving(true)
            await logout()
          }}
        >
          <LogOut aria-hidden /> Keluar
        </Button>
      </Stack>
    </Sheet>
  )
}
