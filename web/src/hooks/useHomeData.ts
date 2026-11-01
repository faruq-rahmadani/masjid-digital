import { useQuery } from '@tanstack/react-query'
import * as activityService from '@/services/activityService'
import * as screenService from '@/services/screenService'

export function useScreens() {
  return useQuery({ queryKey: ['screens'], queryFn: screenService.listScreens })
}

export function useRecentActivity() {
  return useQuery({ queryKey: ['activity', 'recent'], queryFn: activityService.listRecentActivity })
}
