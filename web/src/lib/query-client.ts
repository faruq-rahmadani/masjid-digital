import { QueryClient } from '@tanstack/react-query'
import { ApiError } from './api-client'

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 60_000,
        // Retrying a 4xx won't change the answer; only retry network/5xx failures.
        retry: (failureCount, error) =>
          !(error instanceof ApiError && error.status < 500) && failureCount < 2,
        refetchOnWindowFocus: false,
      },
    },
  })
}
