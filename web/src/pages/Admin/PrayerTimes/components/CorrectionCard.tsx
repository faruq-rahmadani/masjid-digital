import { RotateCcw } from 'lucide-react'
import { Button, Card, SectionHeader, Stepper, Table, TBody, Td, THead, Th, Tr } from '@/components'
import type { PrayerSettings, PrayerTimeEntry } from '@/types/prayer'
import { cn } from '@/utils/cn'
import { DEFAULT_PRAYER_SETTINGS } from '@/utils/prayer/calculate'
import { SourceMark } from './SourceMark'

type CorrectionCardProps = {
  /** Today's times from the draft settings; null when the location is not set. */
  schedule: PrayerTimeEntry[] | null
  settings: PrayerSettings
  /** What is saved, to mark the rows changed since. */
  saved: PrayerSettings
  onChange: (settings: PrayerSettings) => void
  /** Asks before putting every correction and the Imsak offset back to the defaults. */
  onReset: () => void
}

const signed = (v: number) => (v > 0 ? `+${v}` : String(v))

const isDefaultCorrections = (settings: PrayerSettings) =>
  Object.values(settings.corrections).every((v) => v === 0)

// Minutes added to each computed time, every day. Today's resulting time sits next to each
// stepper, since the preview above is usually scrolled out of view by here.
export function CorrectionCard({
  schedule,
  settings,
  saved,
  onChange,
  onReset,
}: CorrectionCardProps) {
  const rows = schedule ?? []
  const isDefault =
    isDefaultCorrections(settings) &&
    settings.imsak_offset_min === DEFAULT_PRAYER_SETTINGS.imsak_offset_min

  return (
    <Card padding="list">
      <SectionHeader
        as="h2"
        title="Koreksi waktu"
        description="Berlaku setiap hari, -10 sampai +10 menit"
        className="pt-3.5"
        action={
          <Button size="sm" variant="secondary" disabled={isDefault} onClick={onReset}>
            <RotateCcw aria-hidden />
            Atur ulang
          </Button>
        }
      />
      <Table caption="Koreksi waktu dan jam hari ini">
        <THead>
          <Tr>
            <Th>Waktu</Th>
            <Th>Hari ini</Th>
            <Th>Koreksi menit</Th>
          </Tr>
        </THead>
        <TBody>
          {rows.map((entry) => {
            const changed =
              entry.key === 'imsak'
                ? settings.imsak_offset_min !== saved.imsak_offset_min
                : settings.corrections[entry.key] !== saved.corrections[entry.key]
            return (
              <Tr
                key={entry.key}
                aria-label={changed ? `${entry.name}, belum disimpan` : undefined}
                // a changed value stays visible down the page: it is what Simpan will send
                className={cn(changed && 'bg-highlight [&_output]:text-gold-ink')}
              >
                <Th scope="row">{entry.name}</Th>
                <Td>
                  <span className="inline-flex items-center gap-1">
                    {entry.time}
                    <SourceMark entry={entry} />
                  </span>
                </Td>
                <Td>
                  {entry.key === 'imsak' ? (
                    <Stepper
                      label="Imsak sebelum Subuh"
                      value={settings.imsak_offset_min}
                      onChange={(v) => onChange({ ...settings, imsak_offset_min: v })}
                      min={1}
                      max={30}
                      format={(v) => `${v} mnt`}
                      className="[&_output]:min-w-12"
                    />
                  ) : (
                    <Stepper
                      label={`Koreksi ${entry.name}`}
                      value={settings.corrections[entry.key]}
                      onChange={(v) =>
                        onChange({
                          ...settings,
                          corrections: { ...settings.corrections, [entry.key]: v },
                        })
                      }
                      min={-10}
                      max={10}
                      format={signed}
                    />
                  )}
                </Td>
              </Tr>
            )
          })}
        </TBody>
      </Table>
      <p className="text-ink-muted pt-1 pb-3 text-xs">Durasi iqamah diatur di menu Iqamah.</p>
    </Card>
  )
}
