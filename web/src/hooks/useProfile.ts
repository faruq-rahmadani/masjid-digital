import { useQuery } from '@tanstack/react-query'
import * as authService from '@/services/authService'

export function useProfile() {
  return useQuery({ queryKey: ['auth', 'profile'], queryFn: authService.getProfile })
}
