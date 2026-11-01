import { RotateCcw } from 'lucide-react'
import { Button, Card, SectionHeader, Stepper, Table, TBody, Td, THead, Th, Tr } from '@/components'
import type { IqamahSettings } from '@/types/iqamah'
import { cn } from '@/utils/cn'
import { PRAYER_KEYS, SCHEDULE_NAMES } from '@/utils/prayer/calculate'
import { DEFAULT_IQAMAH_SETTINGS } from '../defaults'

type IqamahTableProps = {
  settings: IqamahSettings
  /** What is saved, to mark the rows changed since. */
  saved: IqamahSettings
  onChange: (settings: IqamahSettings) => void
  /** Asks before putting the table back to the defaults. */
  onReset: () => void
}

const mnt = (v: number) => `${v} mnt`

// Iqamah and mode sholat per waktu. Jumat has its own card: its iqamah is "setelah khutbah".
export function IqamahTable({ settings, saved, onChange, onReset }: IqamahTableProps) {
  const isDefault = PRAYER_KEYS.every(
    (key) =>
      settings.iqamah_min[key] === DEFAULT_IQAMAH_SETTINGS.iqamah_min[key] &&
      settings.sholat_min[key] === DEFAULT_IQAMAH_SETTINGS.sholat_min[key],
  )

  return (
    <Card padding="list">
      <SectionHeader
        as="h2"
        title="Iqamah & mode sholat"
        description="Iqamah dihitung dari jam adzan"
        className="pt-3.5"
        action={
          <Button size="sm" variant="secondary" disabled={isDefault} onClick={onReset}>
            <RotateCcw aria-hidden />
            Atur ulang
          </Button>
        }
      />
      <Table caption="Iqamah dan mode sholat per waktu">
        <THead>
          <Tr>
            <Th>Waktu</Th>
            <Th>Iqamah</Th>
            <Th>Mode sholat</Th>
          </Tr>
        </THead>
        <TBody>
          {PRAYER_KEYS.map((key) => {
            const name = SCHEDULE_NAMES[key]
            const changed =
              settings.iqamah_min[key] !== saved.iqamah_min[key] ||
              settings.sholat_min[key] !== saved.sholat_min[key]
            return (
              <Tr
                key={key}
                aria-label={changed ? `${name}, belum disimpan` : undefined}
                className={cn(changed && 'bg-highlight [&_output]:text-gold-ink')}
              >
                <Th scope="row">{name}</Th>
                <Td>
                  <Stepper
                    label={`Iqamah ${name}`}
                    value={settings.iqamah_min[key]}
                    onChange={(v) =>
                      onChange({ ...settings, iqamah_min: { ...settings.iqamah_min, [key]: v } })
                    }
                    min={1}
                    max={30}
                    format={mnt}
                    className="[&_output]:min-w-12"
                  />
                </Td>
                <Td>
                  <Stepper
                    label={`Mode sholat ${name}`}
                    value={settings.sholat_min[key]}
                    onChange={(v) =>
                      onChange({ ...settings, sholat_min: { ...settings.sholat_min, [key]: v } })
                    }
                    min={5}
                    max={30}
                    format={mnt}
                    className="[&_output]:min-w-12"
                  />
                </Td>
              </Tr>
            )
          })}
        </TBody>
      </Table>
      <p className="text-ink-muted pt-1 pb-3 text-xs">
        Iqamah 1 sampai 30 menit, mode sholat 5 sampai 30 menit.
      </p>
    </Card>
  )
}
