import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as screenService from '@/services/screenService'
import type { Screen, ScreenCommand, UpdateScreenRequest } from '@/types/screen'

const screenKeys = {
  all: ['screens'] as const,
}

/** Online status changes on its own, so the list refreshes while it is open. */
const STATUS_REFRESH_MS = 15_000

/** `enabled` false for a role that can't see the screens (the API would answer 403). */
export function useScreens(enabled = true) {
  return useQuery({
    queryKey: screenKeys.all,
    queryFn: screenService.listScreens,
    refetchInterval: STATUS_REFRESH_MS,
    enabled,
  })
}

/** Patches the saved screen into the cached list. */
function usePatchScreen() {
  const queryClient = useQueryClient()
  return (saved: Screen) =>
    queryClient.setQueryData<Screen[]>(screenKeys.all, (list) =>
      list?.map((s) => (s.id === saved.id ? saved : s)),
    )
}

export function useUpdateScreen() {
  const patch = usePatchScreen()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: UpdateScreenRequest }) =>
      screenService.updateScreen(id, body),
    onSuccess: patch,
  })
}

export function useBlockScreen() {
  const patch = usePatchScreen()
  return useMutation({ mutationFn: screenService.blockScreen, onSuccess: patch })
}

export function useUnblockScreen() {
  const patch = usePatchScreen()
  return useMutation({ mutationFn: screenService.unblockScreen, onSuccess: patch })
}

export function useSendScreenCommand() {
  return useMutation({
    mutationFn: ({ command, id }: { command: ScreenCommand; id?: number }) =>
      screenService.sendScreenCommand(command, id),
  })
}
