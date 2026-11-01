import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as iqamahService from '@/services/iqamahService'
import type { IqamahSettings } from '@/types/iqamah'

const iqamahKeys = {
  settings: ['iqamah-settings'] as const,
  publicSettings: ['public', 'iqamah-settings'] as const,
}

export function useIqamahSettings() {
  return useQuery({ queryKey: iqamahKeys.settings, queryFn: iqamahService.getIqamahSettings })
}

export function useUpdateIqamahSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: IqamahSettings) => iqamahService.updateIqamahSettings(body),
    onSuccess: (saved) => queryClient.setQueryData(iqamahKeys.settings, saved),
  })
}

// The TV runs for days: refetch on a timer and keep the last data when offline.
const TV_REFRESH = { refetchInterval: 5 * 60_000, refetchIntervalInBackground: true } as const

export function usePublicIqamahSettings() {
  return useQuery({
    queryKey: iqamahKeys.publicSettings,
    queryFn: iqamahService.getPublicIqamahSettings,
    ...TV_REFRESH,
  })
}
