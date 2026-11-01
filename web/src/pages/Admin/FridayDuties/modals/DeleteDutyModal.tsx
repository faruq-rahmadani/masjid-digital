import { ConfirmDialog } from '@/components'
import type { FridayDuty } from '@/types/fridayDuty'
import { formatLongDay } from '@/utils/format/day'

type DeleteDutyModalProps = {
  /** The petugas to delete; null keeps the modal closed. */
  duty: FridayDuty | null
  onClose: () => void
  onConfirm: (duty: FridayDuty) => void
}

export function DeleteDutyModal({ duty, onClose, onConfirm }: DeleteDutyModalProps) {
  return (
    <ConfirmDialog
      open={duty !== null}
      onClose={onClose}
      title="Hapus petugas Jumat?"
      description={duty && `Layar Jumat ${formatLongDay(duty.date)} tampil tanpa nama petugas.`}
      confirmLabel="Hapus"
      danger
      onConfirm={() => duty && onConfirm(duty)}
    />
  )
}
