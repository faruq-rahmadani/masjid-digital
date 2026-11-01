import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useMosqueProfile, usePublicMosqueProfile } from '@/hooks/useMosqueProfile'
import * as prayerTimeService from '@/services/prayerTimeService'
import type { MosqueProfile } from '@/types/mosque'
import type {
  CreatePrayerOverrideRequest,
  PrayerSettings,
  PrayerTimeEntry,
  PrayerTimeOverride,
} from '@/types/prayer'
import { prayerSchedule } from '@/utils/prayer/calculate'

export const prayerKeys = {
  settings: ['prayer-settings'] as const,
  overrides: ['prayer-overrides'] as const,
  overridesBetween: (from: string, to: string) => ['prayer-overrides', from, to] as const,
  publicConfig: ['public', 'prayer-config'] as const,
}

export function usePrayerSettings() {
  return useQuery({ queryKey: prayerKeys.settings, queryFn: prayerTimeService.getPrayerSettings })
}

export function useUpdatePrayerSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: PrayerSettings) => prayerTimeService.updatePrayerSettings(body),
    onSuccess: (saved) => queryClient.setQueryData(prayerKeys.settings, saved),
  })
}

export function usePrayerOverrides(from: string, to: string) {
  return useQuery({
    queryKey: prayerKeys.overridesBetween(from, to),
    queryFn: () => prayerTimeService.listPrayerOverrides(from, to),
  })
}

export function useCreatePrayerOverride() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: CreatePrayerOverrideRequest) => prayerTimeService.createPrayerOverride(body),
    // ranges overlap, so refetch every cached one
    onSuccess: () => queryClient.invalidateQueries({ queryKey: prayerKeys.overrides }),
  })
}

export function useDeletePrayerOverride() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => prayerTimeService.deletePrayerOverride(id),
    onSuccess: (_, id) =>
      queryClient.setQueriesData<PrayerTimeOverride[]>({ queryKey: prayerKeys.overrides }, (list) =>
        list?.filter((o) => o.id !== id),
      ),
  })
}

// Same refresh as the other public reads: the TV runs for days and keeps the last data offline.
const TV_REFRESH = { refetchInterval: 5 * 60_000, refetchIntervalInBackground: true } as const

export function usePublicPrayerConfig() {
  return useQuery({
    queryKey: prayerKeys.publicConfig,
    queryFn: prayerTimeService.getPublicPrayerConfig,
    ...TV_REFRESH,
  })
}

/** `null` until the mosque's location is saved: there is nothing to compute from. */
type Schedule = { data: PrayerTimeEntry[] | null | undefined; isPending: boolean }

function useSchedule(
  date: string,
  profile: MosqueProfile | null | undefined,
  settings: PrayerSettings | undefined,
  overrides: PrayerTimeOverride[] | undefined,
): Schedule {
  const data = useMemo(() => {
    if (profile === undefined || !settings || !overrides) return undefined
    if (!profile || (!profile.latitude && !profile.longitude)) return null
    return prayerSchedule(date, profile, settings, overrides)
  }, [date, profile, settings, overrides])
  return { data, isPending: data === undefined }
}

/** The admin's view of a day's times (Home). `date` is the mosque's "YYYY-MM-DD". */
export function useDaySchedule(date: string) {
  const profile = useMosqueProfile()
  const settings = usePrayerSettings()
  const overrides = usePrayerOverrides(date, date)
  return useSchedule(date, profile.data, settings.data, overrides.data)
}

/** The TV's view of a day's times, from the public endpoints. */
export function usePublicDaySchedule(date: string) {
  const profile = usePublicMosqueProfile()
  const config = usePublicPrayerConfig()
  return useSchedule(date, profile.data, config.data?.settings, config.data?.overrides)
}
