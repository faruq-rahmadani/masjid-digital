import { Card, RadioCard, SectionHeader, Stack, Stepper } from '@/components'
import type { IqamahSettings, SholatDisplay } from '@/types/iqamah'
import { SettingRow } from './SettingRow'

type DisplayCardProps = {
  settings: IqamahSettings
  saved: IqamahSettings
  onChange: (settings: IqamahSettings) => void
}

const DISPLAYS: { value: SholatDisplay; title: string; description: string }[] = [
  { value: 'dim', title: 'Teks redup', description: '"Sholat sedang berlangsung" + jam kecil' },
  { value: 'black', title: 'Gelap total', description: 'Layar hitam penuh' },
]

export function DisplayCard({ settings, saved, onChange }: DisplayCardProps) {
  return (
    <Card padding="md" className="flex flex-col gap-3">
      <SectionHeader as="h3" title="Tampilan layar" />
      <SettingRow
        title="Durasi tampilan adzan"
        description='Layar "Telah masuk waktu"'
        changed={settings.adzan_display_min !== saved.adzan_display_min}
      >
        <Stepper
          label="Durasi tampilan adzan"
          value={settings.adzan_display_min}
          onChange={(v) => onChange({ ...settings, adzan_display_min: v })}
          min={1}
          max={10}
          format={(v) => `${v} mnt`}
          size="md"
        />
      </SettingRow>
      <fieldset className="m-0 min-w-0 border-0 p-0">
        <legend className="text-ink-soft mb-2 text-xs font-semibold">Tampilan saat sholat</legend>
        <Stack gap="sm">
          {DISPLAYS.map((display) => (
            <RadioCard
              key={display.value}
              name="sholat-display"
              value={display.value}
              checked={settings.sholat_display === display.value}
              onChange={() => onChange({ ...settings, sholat_display: display.value })}
              title={display.title}
              description={display.description}
            />
          ))}
        </Stack>
      </fieldset>
    </Card>
  )
}
