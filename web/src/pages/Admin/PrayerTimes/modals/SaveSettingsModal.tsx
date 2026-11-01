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
      title="Simpan pengaturan waktu sholat?"
      description="Koreksi waktu, jadwal mingguan dan koreksi Hijriah langsung dipakai di semua layar TV."
      confirmLabel="Simpan"
      onConfirm={onConfirm}
    />
  )
}
