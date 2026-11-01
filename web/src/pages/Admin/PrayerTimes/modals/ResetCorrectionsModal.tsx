import { ConfirmDialog } from '@/components'

type ResetCorrectionsModalProps = {
  open: boolean
  onClose: () => void
  onConfirm: () => void
}

export function ResetCorrectionsModal({ open, onClose, onConfirm }: ResetCorrectionsModalProps) {
  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      title="Atur ulang koreksi waktu?"
      description="Semua koreksi kembali ke 0 menit dan Imsak ke 10 menit sebelum Subuh. Belum tersimpan sampai menekan Simpan."
      confirmLabel="Atur ulang"
      danger
      onConfirm={onConfirm}
    />
  )
}
