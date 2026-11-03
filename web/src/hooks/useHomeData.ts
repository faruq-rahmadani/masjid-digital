import { useQuery } from '@tanstack/react-query'
import * as activityService from '@/services/activityService'

export function useRecentActivity() {
  return useQuery({ queryKey: ['activity', 'recent'], queryFn: activityService.listRecentActivity })
}
