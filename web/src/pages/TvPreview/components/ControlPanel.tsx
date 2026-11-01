import { Pause, Play } from 'lucide-react'
import { useState } from 'react'
import { Badge, Button, SegmentedControl } from '@/components'
import type { DisplayState } from '@/pages/Tv/displayMode'
import type { SholatDisplay } from '@/types/iqamah'
import type { Scenario, ScenarioGroup } from '../scenarios'

const SPEEDS = [
  { value: '1', label: '1×' },
  { value: '10', label: '10×' },
  { value: '60', label: '60×' },
] as const

const SHOLAT_DISPLAYS = [
  { value: 'dim', label: 'Teks redup' },
  { value: 'black', label: 'Gelap total' },
] as const

const DUTY_OPTIONS = [
  { value: 'filled', label: 'Terisi' },
  { value: 'empty', label: 'Kosong' },
] as const

const MODE_LABELS: Record<DisplayState['mode'], string> = {
  normal: 'Normal',
  adzan: 'Adzan',
  iqamah: 'Hitung mundur iqamah',
  sholat: 'Mode sholat',
  jumat: 'Layar Jumat',
}

const JUMAT_LABELS = { before: 'Layar Jumat', khutbah: 'Khutbah' } as const

type ControlPanelProps = {
  state: DisplayState
  clockLabel: string
  groups: ScenarioGroup[]
  speed: number
  paused: boolean
  sholatDisplay: SholatDisplay
  hasDuty: boolean
  onScenario: (scenario: Scenario) => void
  onSpeed: (speed: number) => void
  onTogglePaused: () => void
  onSholatDisplay: (display: SholatDisplay) => void
  onHasDuty: (hasDuty: boolean) => void
}

const modeLabel = (state: DisplayState) =>
  state.mode === 'jumat' ? JUMAT_LABELS[state.phase] : MODE_LABELS[state.mode]

// Floats over the TV stage. The stage hides the cursor and toggles fullscreen on double-click,
// so the panel shows the cursor again and keeps its double-clicks to itself.
export function ControlPanel(props: ControlPanelProps) {
  const [open, setOpen] = useState(true)

  if (!open) {
    return (
      <div className="fixed right-4 bottom-4 z-10 cursor-auto">
        <Button size="sm" variant="secondary" onClick={() => setOpen(true)}>
          Panel pratinjau
        </Button>
      </div>
    )
  }

  return (
    <aside
      aria-label="Panel pratinjau TV"
      onDoubleClick={(e) => e.stopPropagation()}
      className="bg-card text-ink border-line fixed right-4 bottom-4 z-10 flex max-h-[calc(100vh-2rem)] w-90 cursor-auto flex-col gap-4 overflow-y-auto rounded-xl border p-4 shadow-lg select-text"
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <h1 className="text-md font-bold">Pratinjau layar TV</h1>
          <p className="text-ink-muted text-sm tabular-nums">{props.clockLabel}</p>
        </div>
        <Button size="sm" variant="ghost" onClick={() => setOpen(false)}>
          Sembunyikan
        </Button>
      </div>

      <div className="flex items-center gap-2">
        <Badge>{modeLabel(props.state)}</Badge>
      </div>

      <div className="flex items-center gap-2">
        <Button
          size="md"
          variant="secondary"
          onClick={props.onTogglePaused}
          aria-label={props.paused ? 'Lanjutkan' : 'Jeda'}
        >
          {props.paused ? <Play size={16} /> : <Pause size={16} />}
        </Button>
        <SegmentedControl
          label="Kecepatan"
          className="flex-1"
          options={SPEEDS}
          value={String(props.speed) as (typeof SPEEDS)[number]['value']}
          onChange={(value) => props.onSpeed(Number(value))}
        />
      </div>

      {props.groups.map((group) => (
        <section key={group.label} className="flex flex-col gap-2">
          <h2 className="text-ink-muted text-xs font-bold tracking-wide uppercase">
            {group.label}
          </h2>
          <div className="flex flex-wrap gap-1.5">
            {group.scenarios.map((scenario) => (
              <Button
                key={scenario.id}
                size="sm"
                variant="secondary"
                onClick={() => props.onScenario(scenario)}
              >
                {scenario.label}
              </Button>
            ))}
          </div>
        </section>
      ))}

      <section className="flex flex-col gap-2">
        <h2 className="text-ink-muted text-xs font-bold tracking-wide uppercase">
          Tampilan saat sholat
        </h2>
        <SegmentedControl
          label="Tampilan saat sholat"
          options={SHOLAT_DISPLAYS}
          value={props.sholatDisplay}
          onChange={props.onSholatDisplay}
        />
      </section>

      <section className="flex flex-col gap-2">
        <h2 className="text-ink-muted text-xs font-bold tracking-wide uppercase">Petugas Jumat</h2>
        <SegmentedControl
          label="Petugas Jumat"
          options={DUTY_OPTIONS}
          value={props.hasDuty ? 'filled' : 'empty'}
          onChange={(value) => props.onHasDuty(value === 'filled')}
        />
      </section>

      <p className="text-ink-muted text-xs">
        Data contoh, bukan dari server. Bunyi alarm belum tersedia di layar TV.
      </p>
    </aside>
  )
}
