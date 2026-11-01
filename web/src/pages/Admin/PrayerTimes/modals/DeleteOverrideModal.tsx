import { ConfirmDialog } from '@/components'
import type { PrayerTimeOverride } from '@/types/prayer'
import { formatLongDay } from '@/utils/format/day'
import { SCHEDULE_NAMES } from '@/utils/prayer/calculate'

type DeleteOverrideModalProps = {
  /** The override to delete; null keeps the modal closed. */
  override: PrayerTimeOverride | null
  onClose: () => void
  onConfirm: (override: PrayerTimeOverride) => void
}

export function DeleteOverrideModal({ override, onClose, onConfirm }: DeleteOverrideModalProps) {
  return (
    <ConfirmDialog
      open={override !== null}
      onClose={onClose}
      title="Hapus jadwal khusus?"
      description={
        override &&
        `${SCHEDULE_NAMES[override.prayer]} ${formatLongDay(override.date)} kembali ke jam hitungan. Layar TV ikut berubah.`
      }
      confirmLabel="Hapus"
      danger
      onConfirm={() => override && onConfirm(override)}
    />
  )
}
