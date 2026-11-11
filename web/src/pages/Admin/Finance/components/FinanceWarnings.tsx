import { Banner } from '@/components'
import type { FinanceWarning } from '@/types/finance'

type FinanceWarningsProps = {
  warnings: FinanceWarning[]
  /** Only this kas's warnings; null shows them all. */
  pocketId?: number | null
}

// Syariah warnings (E7-21): amil above 1/8 of the zakat income, Zakat Fitrah left after 1 Syawal.
// They stay while the condition lasts, so they have no close button.
export function FinanceWarnings({ warnings, pocketId = null }: FinanceWarningsProps) {
  const shown = warnings.filter((w) => pocketId === null || w.pocket_id === pocketId)
  return shown.map((w) => (
    <Banner
      key={`${w.kind}-${w.pocket_id}`}
      tone="warning"
      title={w.kind === 'amil_cap' ? 'Bagian amil melebihi batas' : 'Zakat Fitrah belum disalurkan'}
    >
      {w.message}
    </Banner>
  ))
}
