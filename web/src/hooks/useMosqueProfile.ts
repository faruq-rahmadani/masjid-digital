import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ApiError } from '@/api/client'
import * as mosqueProfileService from '@/services/mosqueProfileService'
import type { UpdateMosqueProfileRequest } from '@/types/mosque'

const mosqueKeys = {
  profile: ['mosque-profile'] as const,
  logo: ['mosque-profile', 'logo'] as const,
  publicProfile: ['public', 'mosque-profile'] as const,
  publicLogo: ['public', 'mosque-profile', 'logo'] as const,
}

// Both resources answer 404 until first saved; that is "empty", not an error.
async function orNull<T>(load: () => Promise<T>) {
  try {
    return await load()
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) return null
    throw error
  }
}

/** The mosque profile, or null while nobody has saved it yet. */
export function useMosqueProfile() {
  return useQuery({
    queryKey: mosqueKeys.profile,
    queryFn: () => orNull(mosqueProfileService.getMosqueProfile),
  })
}

export function useUpdateMosqueProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: UpdateMosqueProfileRequest) =>
      mosqueProfileService.updateMosqueProfile(body),
    onSuccess: (saved) => {
      queryClient.setQueryData(mosqueKeys.profile, saved)
      // the tab title (useDocumentBranding) reads the public twin
      queryClient.setQueryData(mosqueKeys.publicProfile, saved)
    },
  })
}

/** The logo as a data URI, or null while none is uploaded. */
export function useMosqueLogo() {
  return useQuery({
    queryKey: mosqueKeys.logo,
    queryFn: async () => (await orNull(mosqueProfileService.getMosqueLogo))?.url ?? null,
  })
}

export function useUploadMosqueLogo() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => mosqueProfileService.uploadMosqueLogo(file),
    // the favicon (useDocumentBranding) reads the public twin
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: mosqueKeys.logo }),
        queryClient.invalidateQueries({ queryKey: mosqueKeys.publicLogo }),
      ]),
  })
}

export function useDeleteMosqueLogo() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: mosqueProfileService.deleteMosqueLogo,
    onSuccess: () => {
      queryClient.setQueryData(mosqueKeys.logo, null)
      queryClient.setQueryData(mosqueKeys.publicLogo, null)
    },
  })
}

// Public endpoints (no login): the TV display and the login screen. The TV runs for days, so by
// default it refetches on a timer (a failed refetch keeps the last data on screen); pass
// { live: false } where one read is enough, such as login.
const TV_REFRESH = { refetchInterval: 5 * 60_000, refetchIntervalInBackground: true } as const

type PublicOptions = { live?: boolean }

export function usePublicMosqueProfile({ live = true }: PublicOptions = {}) {
  return useQuery({
    queryKey: mosqueKeys.publicProfile,
    queryFn: () => orNull(mosqueProfileService.getPublicMosqueProfile),
    ...(live && TV_REFRESH),
  })
}

export function usePublicMosqueLogo({ live = true }: PublicOptions = {}) {
  return useQuery({
    queryKey: mosqueKeys.publicLogo,
    queryFn: async () => (await orNull(mosqueProfileService.getPublicMosqueLogo))?.url ?? null,
    ...(live && TV_REFRESH),
  })
}
