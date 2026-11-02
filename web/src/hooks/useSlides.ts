import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { optimistic } from '@/hooks/optimistic'
import * as slideService from '@/services/slideService'
import type { CreateSlideRequest, Slide, UpdateSlideRequest } from '@/types/slide'

export const slideKeys = {
  list: ['slides'] as const,
  public: ['public', 'slides'] as const,
}

export function useSlides() {
  return useQuery({
    queryKey: slideKeys.list,
    queryFn: slideService.listSlides,
  })
}

export function useCreateSlide() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ file, ...body }: CreateSlideRequest & { file: File }) =>
      slideService.createSlide(file, body),
    onSuccess: (created) =>
      queryClient.setQueryData<Slide[]>(slideKeys.list, (list) => [...(list ?? []), created]),
  })
}

// Toggling active saves on tap, so the button flips at once and rolls back on failure.
export function useUpdateSlide() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...body }: UpdateSlideRequest & { id: number }) =>
      slideService.updateSlide(id, body),
    onMutate: ({ id, ...body }) =>
      optimistic<Slide[]>(queryClient, slideKeys.list, (list) =>
        list.map((s) => (s.id === id ? { ...s, ...body } : s)),
      ),
    onError: (_, __, context) => context?.rollback(),
    onSuccess: (saved) =>
      queryClient.setQueryData<Slide[]>(slideKeys.list, (list) =>
        list?.map((s) => (s.id === saved.id ? saved : s)),
      ),
  })
}

export function useDeleteSlide() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => slideService.deleteSlide(id),
    onSuccess: (_, id) =>
      queryClient.setQueryData<Slide[]>(slideKeys.list, (list) => list?.filter((s) => s.id !== id)),
  })
}

export function useReorderSlides() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (ids: number[]) => slideService.reorderSlides(ids),
    onMutate: (ids) =>
      optimistic<Slide[]>(queryClient, slideKeys.list, (list) =>
        ids.flatMap((id) => list.find((s) => s.id === id) ?? []),
      ),
    onError: (_, __, context) => {
      context?.rollback()
      // most likely someone else added or deleted a slide meanwhile
      queryClient.invalidateQueries({ queryKey: slideKeys.list })
    },
    onSuccess: (list) => queryClient.setQueryData(slideKeys.list, list),
  })
}

// The TV reloads it on the server clock's 5-minute marks (useAlignedRefresh).
export function usePublicSlides() {
  return useQuery({
    queryKey: slideKeys.public,
    queryFn: slideService.getPublicSlides,
  })
}
