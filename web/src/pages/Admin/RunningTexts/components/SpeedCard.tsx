import { Card, SegmentedControl, Ticker, useToast } from '@/components'
import { useUpdateRunningTextSetting } from '@/hooks/useRunningTexts'
import type { RunningText, ScrollSpeed } from '@/types/runningText'
import { SPEED_OPTIONS } from '../utils'

type SpeedCardProps = {
  speed: ScrollSpeed
  texts: RunningText[]
}

// One speed for every text, saved on tap. The preview is the TV's own INFO bar, scaled down.
export function SpeedCard({ speed, texts }: SpeedCardProps) {
  const update = useUpdateRunningTextSetting()
  const toast = useToast()
  const active = texts.filter((t) => t.active).map((t) => t.text)

  const change = (next: ScrollSpeed) =>
    update.mutate(
      { speed: next },
      { onError: (error) => toast({ tone: 'danger', message: error.message }) },
    )

  return (
    <Card padding="md" className="flex flex-col gap-2.5">
      <h2 className="text-md font-bold">Kecepatan gulir</h2>
      <SegmentedControl
        label="Kecepatan gulir"
        options={SPEED_OPTIONS}
        value={speed}
        onChange={change}
      />
      <p className="text-ink-muted mt-1 text-xs font-bold tracking-wide">PRATINJAU</p>
      {active.length > 0 ? (
        <Ticker texts={active} speed={speed} size="preview" />
      ) : (
        <p className="text-ink-muted text-sm">
          Tidak ada teks aktif, INFO bar di TV disembunyikan.
        </p>
      )}
    </Card>
  )
}
