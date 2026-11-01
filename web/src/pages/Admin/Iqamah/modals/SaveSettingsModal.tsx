import { ConfirmDialog } from '@/components'

type SaveSettingsModalProps = {
  open: boolean
  onClose: () => void
  onConfirm: () => void
}

export function SaveSettingsModal({ open, onClose, onConfirm }: SaveSettingsModalProps) {
  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      title="Simpan pengaturan iqamah?"
      description="Iqamah, mode sholat, tampilan layar dan alarm langsung dipakai di semua layar TV."
      confirmLabel="Simpan"
      onConfirm={onConfirm}
    />
  )
}
