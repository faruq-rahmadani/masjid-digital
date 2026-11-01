import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as fridayDutyService from '@/services/fridayDutyService'
import type { FridayDuty, SaveFridayDutyRequest } from '@/types/fridayDuty'

const fridayDutyKeys = {
  all: ['friday-duties'] as const,
  between: (from: string, to: string) => ['friday-duties', from, to] as const,
  upcoming: ['public', 'friday-duties'] as const,
}

export function useFridayDuties(from: string, to: string) {
  return useQuery({
    queryKey: fridayDutyKeys.between(from, to),
    queryFn: () => fridayDutyService.listFridayDuties(from, to),
  })
}

export function useSaveFridayDuty() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ date, body }: { date: string; body: SaveFridayDutyRequest }) =>
      fridayDutyService.saveFridayDuty(date, body),
    // an upsert can add a Friday to any cached range: refetch them, as after a create
    onSuccess: () => queryClient.invalidateQueries({ queryKey: fridayDutyKeys.all }),
  })
}

export function useDeleteFridayDuty() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (date: string) => fridayDutyService.deleteFridayDuty(date),
    onSuccess: (_, date) =>
      queryClient.setQueriesData<FridayDuty[]>({ queryKey: fridayDutyKeys.all }, (list) =>
        list?.filter((d) => d.date !== date),
      ),
  })
}

// The TV runs for days: refetch on a timer and keep the last data when offline.
const TV_REFRESH = { refetchInterval: 5 * 60_000, refetchIntervalInBackground: true } as const

export function useUpcomingFridayDuties() {
  return useQuery({
    queryKey: fridayDutyKeys.upcoming,
    queryFn: fridayDutyService.listUpcomingFridayDuties,
    ...TV_REFRESH,
  })
}
