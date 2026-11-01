import { Card, Checkbox, Field, SectionHeader, Select, Slider, Stack } from '@/components'
import type { AlarmSettings, AlarmSound, IqamahSettings } from '@/types/iqamah'

type AlarmCardProps = {
  settings: IqamahSettings
  onChange: (settings: IqamahSettings) => void
}

const SOUNDS: { value: AlarmSound; label: string }[] = [
  { value: 'beep', label: 'Beep pendek' },
  { value: 'bell', label: 'Bel lembut' },
]

// The details only show while the alarm is on: nothing to set for a silent screen.
export function AlarmCard({ settings, onChange }: AlarmCardProps) {
  const alarm = settings.alarm
  const set = (changes: Partial<AlarmSettings>) =>
    onChange({ ...settings, alarm: { ...alarm, ...changes } })

  return (
    <Card padding="md" className="flex flex-col gap-3">
      <div className="flex items-center justify-between gap-3">
        <SectionHeader as="h3" title="Alarm" />
        <Checkbox
          label="Aktif"
          checked={alarm.enabled}
          onChange={(e) => set({ enabled: e.currentTarget.checked })}
        />
      </div>
      {alarm.enabled && (
        <Stack gap="sm">
          <Checkbox
            label="Saat masuk waktu adzan"
            checked={alarm.at_adzan}
            onChange={(e) => set({ at_adzan: e.currentTarget.checked })}
          />
          <Checkbox
            label="10 detik terakhir sebelum iqamah"
            checked={alarm.before_iqamah}
            onChange={(e) => set({ before_iqamah: e.currentTarget.checked })}
          />
          <Field label="Bunyi">
            <Select
              options={SOUNDS}
              value={alarm.sound}
              onChange={(sound) => set({ sound })}
              aria-label="Bunyi"
            />
          </Field>
          <Field label={`Volume ${alarm.volume}`}>
            <Slider
              min={0}
              max={100}
              step={5}
              value={alarm.volume}
              onChange={(e) => set({ volume: Number(e.currentTarget.value) })}
              minLabel="Pelan"
              maxLabel="Keras"
              aria-label="Volume"
            />
          </Field>
          <p className="text-ink-muted text-xs">
            Layar TV perlu disentuh sekali setelah menyala agar bunyi bisa diputar.
          </p>
        </Stack>
      )}
    </Card>
  )
}
