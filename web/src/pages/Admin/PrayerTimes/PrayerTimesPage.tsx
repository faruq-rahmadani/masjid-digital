import { useMemo, useState } from 'react'
import {
  ActionBar,
  Button,
  Callout,
  Card,
  LeaveGuard,
  PageBody,
  PageHeader,
  SectionHeader,
  Skeleton,
  useToast,
} from '@/components'
import { useMosqueProfile } from '@/hooks/useMosqueProfile'
import { useNow } from '@/hooks/useNow'
import {
  useDeletePrayerOverride,
  usePrayerOverrides,
  usePrayerSettings,
  useUpdatePrayerSettings,
} from '@/hooks/usePrayerTimes'
import type { MosqueProfile } from '@/types/mosque'
import type { PrayerSettings, PrayerTimeOverride, WeeklyPrayerOverride } from '@/types/prayer'
import { DEFAULT_PRAYER_SETTINGS, prayerSchedule } from '@/utils/prayer/calculate'
import { addDays, localDate } from '@/utils/prayer/calendar'
import { ContextCard } from './components/ContextCard'
import { CorrectionCard } from './components/CorrectionCard'
import { DateOverridesCard } from './components/DateOverridesCard'
import { HijriCard } from './components/HijriCard'
import { PreviewCard } from './components/PreviewCard'
import { WeeklyRulesCard } from './components/WeeklyRulesCard'
import { AddOverrideDrawer } from './drawers/AddOverrideDrawer'
import { AddWeeklyRuleDrawer } from './drawers/AddWeeklyRuleDrawer'
import { DeleteOverrideModal } from './modals/DeleteOverrideModal'
import { RemoveWeeklyRuleModal } from './modals/RemoveWeeklyRuleModal'
import { ResetCorrectionsModal } from './modals/ResetCorrectionsModal'
import { SaveSettingsModal } from './modals/SaveSettingsModal'

/** How far ahead the date overrides list looks; the TV itself gets 30 days. */
const OVERRIDE_DAYS = 60

// Waktu sholat (E4). Super admin and takmir; the route is guarded by RequireAccess.
export function PrayerTimesPage() {
  return (
    <>
      <PageHeader title="Waktu Sholat" />
      <PrayerTimesContent />
    </>
  )
}

/** Error, then loading, then the form once the profile and settings are both known. */
function PrayerTimesContent() {
  const profile = useMosqueProfile()
  const settings = usePrayerSettings()
  const failed = profile.error ?? settings.error

  if (failed) {
    return (
      <PageBody>
        <Callout
          tone="danger"
          title="Waktu sholat gagal dimuat"
          action={
            <Button
              size="md"
              variant="secondary"
              onClick={() => {
                profile.refetch()
                settings.refetch()
              }}
            >
              Coba lagi
            </Button>
          }
        >
          {failed.message}
        </Callout>
      </PageBody>
    )
  }

  if (profile.data === undefined || settings.data === undefined) {
    return (
      <PageBody>
        <Card aria-busy aria-label="Memuat waktu sholat" className="flex flex-col gap-3">
          <Skeleton className="h-8" />
          <Skeleton className="h-8" />
          <Skeleton className="h-8" />
        </Card>
      </PageBody>
    )
  }

  // Mounted once loaded, so the draft starts from the saved settings.
  return <PrayerTimesForm profile={profile.data} saved={settings.data} />
}

type PrayerTimesFormProps = {
  profile: MosqueProfile | null
  saved: PrayerSettings
}

function PrayerTimesForm({ profile, saved }: PrayerTimesFormProps) {
  const [draft, setDraft] = useState(saved)
  const [drawer, setDrawer] = useState<'override' | 'weekly' | null>(null)
  const [confirmSave, setConfirmSave] = useState(false)
  const [confirmReset, setConfirmReset] = useState(false)
  const [removingRule, setRemovingRule] = useState<WeeklyPrayerOverride | null>(null)
  const [deleting, setDeleting] = useState<PrayerTimeOverride | null>(null)
  const update = useUpdatePrayerSettings()
  const remove = useDeletePrayerOverride()
  const toast = useToast()

  const now = useNow()
  const today = localDate(now, profile?.timezone ?? 'Asia/Jakarta')
  const overrides = usePrayerOverrides(today, addDays(today, OVERRIDE_DAYS))
  const location = profile && (profile.latitude || profile.longitude) ? profile : null

  const days = useMemo(() => {
    if (!location) return null
    return Array.from({ length: 7 }, (_, i) => {
      const date = addDays(today, i)
      return { date, schedule: prayerSchedule(date, location, draft, overrides.data) }
    })
  }, [location, today, draft, overrides.data])

  const dirty = JSON.stringify(draft) !== JSON.stringify(saved)

  const save = () =>
    update.mutate(draft, {
      onSuccess: () =>
        toast({ message: 'Waktu sholat disimpan', description: 'Layar TV ikut diperbarui' }),
      onError: (error) => toast({ tone: 'danger', message: error.message }),
    })

  const removeRule = (rule: WeeklyPrayerOverride) =>
    setDraft({ ...draft, weekly_overrides: draft.weekly_overrides.filter((r) => r !== rule) })

  const deleteOverride = (override: PrayerTimeOverride) =>
    remove.mutate(override.id, {
      onSuccess: () => toast({ message: 'Jadwal khusus dihapus' }),
      onError: (error) => toast({ tone: 'danger', message: error.message }),
    })

  return (
    <>
      <PageBody>
        {!location && (
          <Callout title="Lokasi masjid belum diisi">
            Waktu sholat dihitung dari lokasi masjid. Pengaturan tetap bisa disimpan.
          </Callout>
        )}
        <ContextCard location={location} />
        {days && <PreviewCard today={today} days={days} />}
        {days && (
          <CorrectionCard
            schedule={days[0]?.schedule ?? null}
            settings={draft}
            saved={saved}
            onChange={setDraft}
            onReset={() => setConfirmReset(true)}
          />
        )}

        <section aria-labelledby="special-settings" className="flex flex-col gap-3.5">
          <SectionHeader
            title={<span id="special-settings">Pengaturan khusus</span>}
            className="mt-2"
          />
          <WeeklyRulesCard
            rules={draft.weekly_overrides}
            onAdd={() => setDrawer('weekly')}
            onRemove={(index) => setRemovingRule(draft.weekly_overrides[index] ?? null)}
          />
          <HijriCard today={today} settings={draft} onChange={setDraft} />
          <DateOverridesCard
            overrides={overrides.data}
            error={overrides.error}
            onAdd={() => setDrawer('override')}
            onDelete={setDeleting}
          />
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

      {drawer === 'override' && <AddOverrideDrawer today={today} onClose={() => setDrawer(null)} />}
      {drawer === 'weekly' && (
        <AddWeeklyRuleDrawer
          rules={draft.weekly_overrides}
          onAdd={(rule) =>
            setDraft({ ...draft, weekly_overrides: [...draft.weekly_overrides, rule] })
          }
          onClose={() => setDrawer(null)}
        />
      )}
      <ResetCorrectionsModal
        open={confirmReset}
        onClose={() => setConfirmReset(false)}
        onConfirm={() =>
          setDraft({
            ...draft,
            corrections: DEFAULT_PRAYER_SETTINGS.corrections,
            imsak_offset_min: DEFAULT_PRAYER_SETTINGS.imsak_offset_min,
          })
        }
      />
      <SaveSettingsModal
        open={confirmSave}
        onClose={() => setConfirmSave(false)}
        onConfirm={save}
      />
      <RemoveWeeklyRuleModal
        rule={removingRule}
        onClose={() => setRemovingRule(null)}
        onConfirm={removeRule}
      />
      <DeleteOverrideModal
        override={deleting}
        onClose={() => setDeleting(null)}
        onConfirm={deleteOverride}
      />
      {/* not while saving: the draft equals the saved settings again once it lands */}
      <LeaveGuard when={dirty && !update.isPending} />
    </>
  )
}
