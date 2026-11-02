import { type QueryClient, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as runningTextService from '@/services/runningTextService'
import type { RunningText, RunningTextSetting, UpdateRunningTextRequest } from '@/types/runningText'

export const runningTextKeys = {
  list: ['running-texts'] as const,
  setting: ['running-text-setting'] as const,
  public: ['public', 'running-texts'] as const,
}

export function useRunningTexts() {
  return useQuery({ queryKey: runningTextKeys.list, queryFn: runningTextService.listRunningTexts })
}

export function useRunningTextSetting() {
  return useQuery({
    queryKey: runningTextKeys.setting,
    queryFn: runningTextService.getRunningTextSetting,
  })
}

export function useCreateRunningText() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (text: string) => runningTextService.createRunningText(text),
    onSuccess: (created) =>
      queryClient.setQueryData<RunningText[]>(runningTextKeys.list, (list) => [
        ...(list ?? []),
        created,
      ]),
  })
}

/** Swap the cached list for `next` and return a rollback for onError. */
async function optimistic<T>(
  queryClient: QueryClient,
  key: readonly unknown[],
  next: (old: T) => T,
) {
  await queryClient.cancelQueries({ queryKey: key })
  const previous = queryClient.getQueryData<T>(key)
  if (previous !== undefined) queryClient.setQueryData<T>(key, next(previous))
  return { rollback: () => queryClient.setQueryData(key, previous) }
}

// Toggling active saves on tap, so the checkbox flips at once and rolls back on failure.
export function useUpdateRunningText() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...body }: UpdateRunningTextRequest & { id: number }) =>
      runningTextService.updateRunningText(id, body),
    onMutate: ({ id, ...body }) =>
      optimistic<RunningText[]>(queryClient, runningTextKeys.list, (list) =>
        list.map((t) => (t.id === id ? { ...t, ...body } : t)),
      ),
    onError: (_, __, context) => context?.rollback(),
    onSuccess: (saved) =>
      queryClient.setQueryData<RunningText[]>(runningTextKeys.list, (list) =>
        list?.map((t) => (t.id === saved.id ? saved : t)),
      ),
  })
}

export function useDeleteRunningText() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => runningTextService.deleteRunningText(id),
    onSuccess: (_, id) =>
      queryClient.setQueryData<RunningText[]>(runningTextKeys.list, (list) =>
        list?.filter((t) => t.id !== id),
      ),
  })
}

export function useReorderRunningTexts() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (ids: number[]) => runningTextService.reorderRunningTexts(ids),
    onMutate: (ids) =>
      optimistic<RunningText[]>(queryClient, runningTextKeys.list, (list) =>
        ids.flatMap((id) => list.find((t) => t.id === id) ?? []),
      ),
    onError: (_, __, context) => {
      context?.rollback()
      // most likely someone else added or deleted a text meanwhile
      queryClient.invalidateQueries({ queryKey: runningTextKeys.list })
    },
    onSuccess: (list) => queryClient.setQueryData(runningTextKeys.list, list),
  })
}

export function useUpdateRunningTextSetting() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: RunningTextSetting) => runningTextService.updateRunningTextSetting(body),
    onMutate: (body) =>
      optimistic<RunningTextSetting>(queryClient, runningTextKeys.setting, () => body),
    onError: (_, __, context) => context?.rollback(),
    onSuccess: (saved) => queryClient.setQueryData(runningTextKeys.setting, saved),
  })
}

// Same refresh as the other public reads: the TV runs for days and keeps the last data offline.
const TV_REFRESH = { refetchInterval: 5 * 60_000, refetchIntervalInBackground: true } as const

export function usePublicRunningTexts() {
  return useQuery({
    queryKey: runningTextKeys.public,
    queryFn: runningTextService.getPublicRunningTexts,
    ...TV_REFRESH,
  })
}
