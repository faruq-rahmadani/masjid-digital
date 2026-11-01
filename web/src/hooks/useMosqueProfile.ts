import { useQuery } from '@tanstack/react-query'
import { ApiError } from '@/api/client'
import * as mosqueProfileService from '@/services/mosqueProfileService'

/** The mosque profile, or null while nobody has saved it yet (the API still not implemented yet). */
export function useMosqueProfile() {
  return useQuery({
    queryKey: ['mosque-profile'],
    queryFn: async () => {
      try {
        return await mosqueProfileService.getMosqueProfile()
      } catch (error) {
        if (error instanceof ApiError && error.status === 404) return null
        throw error
      }
    },
  })
}
