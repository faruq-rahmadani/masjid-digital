import { Card, SectionHeader, Stepper } from '@/components'
import type { IqamahSettings } from '@/types/iqamah'
import { SettingRow } from './SettingRow'

type JumatCardProps = {
  settings: IqamahSettings
  saved: IqamahSettings
  onChange: (settings: IqamahSettings) => void
}

// Friday Dzuhur runs differently: no iqamah countdown, the khutbah comes after the adzan.
export function JumatCard({ settings, saved, onChange }: JumatCardProps) {
  return (
    <Card padding="md">
      <SectionHeader as="h3" title="Sholat Jumat" description="Iqamah setelah khutbah" />
      <div className="mt-2 flex flex-col">
        <SettingRow
          title="Layar Jumat"
          description="Tampil sebelum Dzuhur"
          changed={settings.jumat_before_min !== saved.jumat_before_min}
        >
          <Stepper
            label="Layar Jumat sebelum Dzuhur"
            value={settings.jumat_before_min}
            onChange={(v) => onChange({ ...settings, jumat_before_min: v })}
            min={0}
            max={120}
            step={5}
            format={(v) => (v === 0 ? 'Mati' : `${v} mnt`)}
            className="[&_output]:min-w-14"
          />
        </SettingRow>
        <SettingRow title="Khutbah" changed={settings.khutbah_min !== saved.khutbah_min}>
          <Stepper
            label="Durasi khutbah"
            value={settings.khutbah_min}
            onChange={(v) => onChange({ ...settings, khutbah_min: v })}
            min={10}
            max={60}
            step={5}
            format={(v) => `${v} mnt`}
            className="[&_output]:min-w-14"
          />
        </SettingRow>
        <SettingRow
          title="Mode sholat Jumat"
          description="Sampai 60 menit"
          changed={settings.sholat_min.jumat !== saved.sholat_min.jumat}
        >
          <Stepper
            label="Mode sholat Jumat"
            value={settings.sholat_min.jumat}
            onChange={(v) =>
              onChange({ ...settings, sholat_min: { ...settings.sholat_min, jumat: v } })
            }
            min={5}
            max={60}
            format={(v) => `${v} mnt`}
            className="[&_output]:min-w-14"
          />
        </SettingRow>
      </div>
    </Card>
  )
}
