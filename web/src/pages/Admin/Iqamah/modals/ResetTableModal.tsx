import { ConfirmDialog } from '@/components'

type ResetTableModalProps = {
  open: boolean
  onClose: () => void
  onConfirm: () => void
}

export function ResetTableModal({ open, onClose, onConfirm }: ResetTableModalProps) {
  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      title="Atur ulang iqamah & mode sholat?"
      description="Semua waktu kembali ke durasi bawaan. Belum tersimpan sampai menekan Simpan."
      confirmLabel="Atur ulang"
      danger
      onConfirm={onConfirm}
    />
  )
}
