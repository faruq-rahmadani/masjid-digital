import type { QueryClient } from '@tanstack/react-query'

/** Swap the cached value for `next` and return a rollback for onError. */
export async function optimistic<T>(
  queryClient: QueryClient,
  key: readonly unknown[],
  next: (old: T) => T,
) {
  await queryClient.cancelQueries({ queryKey: key })
  const previous = queryClient.getQueryData<T>(key)
  if (previous !== undefined) queryClient.setQueryData<T>(key, next(previous))
  return { rollback: () => queryClient.setQueryData(key, previous) }
}
