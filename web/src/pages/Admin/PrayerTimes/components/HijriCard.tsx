import { Card, Field, Input, SectionHeader, SegmentedControl } from '@/components'
import type { PrayerSettings } from '@/types/prayer'
import { hijriDate } from '@/utils/prayer/hijri'

const OFFSETS = ['-2', '-1', '0', '1', '2'] as const

type HijriCardProps = {
  today: string
  settings: PrayerSettings
  onChange: (settings: PrayerSettings) => void
}

// The Hijri date follows Umm al-Qura; shift it when the sidang isbat starts a month on another day.
export function HijriCard({ today, settings, onChange }: HijriCardProps) {
  return (
    <Card className="flex flex-col gap-3">
      <SectionHeader
        as="h3"
        title="Koreksi tanggal Hijriah"
        description="Disimpan dengan tombol Simpan"
      />
      <SegmentedControl
        label="Koreksi hari"
        variant="split"
        options={OFFSETS.map((v) => ({ value: v, label: Number(v) > 0 ? `+${v}` : v }))}
        value={String(settings.hijri_offset_days) as (typeof OFFSETS)[number]}
        onChange={(v) => onChange({ ...settings, hijri_offset_days: Number(v) })}
      />
      <Field label="Berlaku mulai" hint="Kosongkan agar berlaku untuk semua tanggal">
        <Input
          type="date"
          value={settings.hijri_offset_from}
          onChange={(e) => onChange({ ...settings, hijri_offset_from: e.currentTarget.value })}
        />
      </Field>
      <p className="text-ink-muted text-xs">
        Tampil di TV hari ini: <b className="text-ink">{hijriDate(today, settings)}</b>
      </p>
    </Card>
  )
}
