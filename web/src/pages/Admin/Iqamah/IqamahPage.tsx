import { useMemo, useState } from 'react'
import {
  ActionBar,
  Button,
  Callout,
  Card,
  LeaveGuard,
  NavCard,
  PageBody,
  PageHeader,
  SectionHeader,
  Skeleton,
  useToast,
} from '@/components'
import { useIqamahSettings, useUpdateIqamahSettings } from '@/hooks/useIqamah'
import { useMosqueProfile } from '@/hooks/useMosqueProfile'
import { useNow } from '@/hooks/useNow'
import { useDaySchedule } from '@/hooks/usePrayerTimes'
import { ROUTES } from '@/route'
import type { IqamahSettings } from '@/types/iqamah'
import type { PrayerKey, PrayerTimeEntry } from '@/types/prayer'
import { PRAYER_KEYS } from '@/utils/prayer/calculate'
import { addDays, localDate, minutesOf, secondsOfDay, weekdayOf } from '@/utils/prayer/calendar'
import { AlarmCard } from './components/AlarmCard'
import { DisplayCard } from './components/DisplayCard'
import { FlowCard } from './components/FlowCard'
import { IqamahTable } from './components/IqamahTable'
import { JumatCard } from './components/JumatCard'
import { DEFAULT_IQAMAH_SETTINGS } from './defaults'
import { prayerFlow } from './flow'
import { ResetTableModal } from './modals/ResetTableModal'
import { SaveSettingsModal } from './modals/SaveSettingsModal'

// Iqamah & Alarm (E5). Super admin and takmir; the route is guarded by RequireAccess.
export function IqamahPage() {
  return (
    <>
      <PageHeader title="Iqamah & Alarm" />
      <IqamahContent />
    </>
  )
}

/** Error, then loading, then the form once the settings are known. */
function IqamahContent() {
  const settings = useIqamahSettings()

  if (settings.error) {
    return (
      <PageBody>
        <Callout
          tone="danger"
          title="Pengaturan iqamah gagal dimuat"
          action={
            <Button size="md" variant="secondary" onClick={() => settings.refetch()}>
              Coba lagi
            </Button>
          }
        >
          {settings.error.message}
        </Callout>
      </PageBody>
    )
  }

  if (!settings.data) {
    return (
      <PageBody>
        <Card aria-busy aria-label="Memuat pengaturan iqamah" className="flex flex-col gap-3">
          <Skeleton className="h-8" />
          <Skeleton className="h-8" />
          <Skeleton className="h-8" />
        </Card>
      </PageBody>
    )
  }

  // Mounted once loaded, so the draft starts from the saved settings.
  return <IqamahForm saved={settings.data} />
}

type IqamahFormProps = { saved: IqamahSettings }

function IqamahForm({ saved }: IqamahFormProps) {
  const [draft, setDraft] = useState(saved)
  const [confirmSave, setConfirmSave] = useState(false)
  const [confirmReset, setConfirmReset] = useState(false)
  const update = useUpdateIqamahSettings()
  const toast = useToast()

  const now = useNow()
  const mosque = useMosqueProfile()
  const timezone = mosque.data?.timezone ?? 'Asia/Jakarta'
  const today = localDate(now, timezone)
  const schedule = useDaySchedule(today)

  const flow = useMemo(() => {
    const next = upcomingPrayer(schedule.data, secondsOfDay(now, timezone) / 60, today)
    if (!next) return null
    return prayerFlow(next.prayer, draft, weekdayOf(next.date) === 'friday')
  }, [schedule.data, now, timezone, today, draft])

  const dirty = JSON.stringify(draft) !== JSON.stringify(saved)

  const save = () =>
    update.mutate(draft, {
      onSuccess: () =>
        toast({ message: 'Pengaturan iqamah disimpan', description: 'Layar TV ikut diperbarui' }),
      onError: (error) => toast({ tone: 'danger', message: error.message }),
    })

  const resetTable = () =>
    setDraft({
      ...draft,
      iqamah_min: DEFAULT_IQAMAH_SETTINGS.iqamah_min,
      sholat_min: { ...DEFAULT_IQAMAH_SETTINGS.sholat_min, jumat: draft.sholat_min.jumat },
    })

  return (
    <>
      <PageBody>
        <FlowCard flow={flow} />
        <IqamahTable
          settings={draft}
          saved={saved}
          onChange={setDraft}
          onReset={() => setConfirmReset(true)}
        />

        <section aria-labelledby="jumat-settings" className="flex flex-col gap-3.5">
          <SectionHeader title={<span id="jumat-settings">Jumat</span>} className="mt-2" />
          <JumatCard settings={draft} saved={saved} onChange={setDraft} />
          <NavCard
            to={ROUTES.fridayDuties}
            title="Petugas Jumat"
            description="Khatib, imam, muadzin dan tema khutbah"
          />
        </section>

        <section aria-labelledby="screen-settings" className="flex flex-col gap-3.5">
          <SectionHeader title={<span id="screen-settings">Layar & alarm</span>} className="mt-2" />
          <DisplayCard settings={draft} saved={saved} onChange={setDraft} />
          <AlarmCard settings={draft} onChange={setDraft} />
        </section>
      </PageBody>

      <ActionBar>
        {dirty && (
          <p role="status" className="text-notice-ink mb-2 text-center text-xs font-semibold">
            Ada perubahan belum disimpan
          </p>
        )}
        <Button
          size="xl"
          block
          loading={update.isPending}
          disabled={!dirty}
          onClick={() => setConfirmSave(true)}
        >
          Simpan dan kirim ke layar
        </Button>
      </ActionBar>

      <SaveSettingsModal
        open={confirmSave}
        onClose={() => setConfirmSave(false)}
        onConfirm={save}
      />
      <ResetTableModal
        open={confirmReset}
        onClose={() => setConfirmReset(false)}
        onConfirm={resetTable}
      />
      {/* not while saving: the draft equals the saved settings again once it lands */}
      <LeaveGuard when={dirty && !update.isPending} />
    </>
  )
}

/**
 * The next adzan after `minutesNow` and the date it falls on. After Isya it is tomorrow's Subuh,
 * shown with today's time (a minute or so off), which is enough for the flow card.
 */
function upcomingPrayer(
  schedule: PrayerTimeEntry[] | null | undefined,
  minutesNow: number,
  today: string,
) {
  const prayers = (schedule ?? []).filter((e): e is PrayerTimeEntry & { key: PrayerKey } =>
    (PRAYER_KEYS as readonly string[]).includes(e.key),
  )
  const later = prayers.find((p) => minutesOf(p.time) > minutesNow)
  if (later) return { prayer: later, date: today }
  const first = prayers[0]
  if (!first) return null
  return { prayer: first, date: addDays(today, 1) }
}
