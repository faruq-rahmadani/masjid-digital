import { useMutation, useQuery } from '@tanstack/react-query'
import * as authService from '@/services/authService'

export function useProfile() {
  return useQuery({ queryKey: ['auth', 'profile'], queryFn: authService.getProfile })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: ({ current, next }: { current: string; next: string }) =>
      authService.changePassword(current, next),
  })
}
