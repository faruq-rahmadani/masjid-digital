import { Card } from '@/components'
import type { PrayerFlow } from '../flow'

// Brand like Waktu sholat's context card: "now" at the top, the next prayer as the TV will run
// it with the unsaved settings, so each stepper below shows its effect in real clock times.
export function FlowCard({ flow }: { flow: PrayerFlow | null }) {
  return (
    <Card tone="brand" className="flex flex-col gap-3">
      <div>
        <p className="text-on-brand-muted text-xs font-semibold">Alur di layar, waktu berikutnya</p>
        <p className="text-lg font-extrabold tabular-nums">
          {flow ? `${flow.name} ${flow.time}` : 'Isi lokasi di Profil masjid'}
        </p>
      </div>
      {flow && (
        <ol className="flex flex-col gap-1.5" aria-label={`Alur ${flow.name} di layar`}>
          {flow.steps.map((step) => (
            <li
              key={step.label}
              className="bg-brand-2 flex items-center justify-between rounded-md px-3 py-2 text-sm"
            >
              <span className="text-on-brand-muted font-semibold">{step.label}</span>
              <span className="font-bold tabular-nums">
                {step.from}–{step.to}
              </span>
            </li>
          ))}
        </ol>
      )}
    </Card>
  )
}
