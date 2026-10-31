import { ConfirmDialog } from '@/components'
import type { User } from '@/types/user'

type DeactivateUserModalProps = {
  /** The user to deactivate; null keeps the modal closed. */
  user: User | null
  onClose: () => void
  onConfirm: (user: User) => void
}

// Users are deactivated, never deleted: their past records keep their author (Go domain rule).
export function DeactivateUserModal({ user, onClose, onConfirm }: DeactivateUserModalProps) {
  return (
    <ConfirmDialog
      open={user !== null}
      onClose={onClose}
      title={`Nonaktifkan ${user?.name ?? ''}?`}
      description="Ia langsung keluar dan tidak bisa masuk lagi. Data yang pernah dicatat tetap tersimpan, dan akun bisa diaktifkan kembali."
      confirmLabel="Nonaktifkan"
      danger
      onConfirm={() => user && onConfirm(user)}
    />
  )
}
