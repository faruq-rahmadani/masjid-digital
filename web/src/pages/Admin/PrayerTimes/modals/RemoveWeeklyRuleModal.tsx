import { ConfirmDialog } from '@/components'
import type { WeeklyPrayerOverride } from '@/types/prayer'
import { SCHEDULE_NAMES } from '@/utils/prayer/calculate'
import { WEEKDAY_LABELS } from '../utils'

type RemoveWeeklyRuleModalProps = {
  /** The rule to remove; null keeps the modal closed. */
  rule: WeeklyPrayerOverride | null
  onClose: () => void
  onConfirm: (rule: WeeklyPrayerOverride) => void
}

export function RemoveWeeklyRuleModal({ rule, onClose, onConfirm }: RemoveWeeklyRuleModalProps) {
  return (
    <ConfirmDialog
      open={rule !== null}
      onClose={onClose}
      title="Hapus jadwal mingguan?"
      description={
        rule &&
        `${SCHEDULE_NAMES[rule.prayer]} setiap ${WEEKDAY_LABELS[rule.weekday]} kembali ke jam hitungan setelah pengaturan disimpan.`
      }
      confirmLabel="Hapus"
      danger
      onConfirm={() => rule && onConfirm(rule)}
    />
  )
}
