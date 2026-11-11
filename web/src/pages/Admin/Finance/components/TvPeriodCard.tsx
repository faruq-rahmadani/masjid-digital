import { Card, Field, Select, useToast } from '@/components'
import { useFinanceMutation, useTvSetting } from '@/hooks/useFinance'
import { updateTvSetting } from '@/services/financeService'
import type { TvPeriod } from '@/types/finance'

const PERIODS = [
  { value: 'today', label: 'Hari ini' },
  { value: 'week', label: 'Minggu ini' },
  { value: 'month', label: 'Bulan ini' },
] as const satisfies readonly { value: TvPeriod; label: string }[]

// The period the TV kas header covers. Which kas show is set per kas ("Tampil di layar TV").
// A select saves right away, like the other one-tap settings.
export function TvPeriodCard() {
  const setting = useTvSetting()
  const save = useFinanceMutation(updateTvSetting)
  const toast = useToast()

  return (
    <Card padding="lg" className="flex flex-col gap-2.5">
      <div>
        <p className="text-sm font-bold">Kas di layar TV</p>
        <p className="text-ink-soft text-xs">
          Kas yang ditandai "Tampil di layar TV" bergiliran di header TV.
        </p>
      </div>
      <Field label="Periode yang ditampilkan">
        <Select
          options={PERIODS}
          value={setting.data?.period}
          disabled={!setting.data || save.isPending}
          onChange={(period) =>
            save.mutate(period, {
              onSuccess: () => toast({ message: 'Periode kas di TV disimpan' }),
              onError: (error) => toast({ tone: 'danger', message: error.message }),
            })
          }
        />
      </Field>
    </Card>
  )
}
