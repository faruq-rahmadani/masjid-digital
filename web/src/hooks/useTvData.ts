import { useQuery } from '@tanstack/react-query'
import * as tvService from '@/services/tvService'

// The TV reloads these on the server clock's 5-minute marks (useAlignedRefresh); TanStack Query
// keeps `data` when a reload fails, so a network blip never blanks the screen.

export function useCashSummary() {
  return useQuery({ queryKey: ['tv', 'cash'], queryFn: tvService.getCashSummary })
}

export function useQuotes() {
  return useQuery({ queryKey: ['tv', 'quotes'], queryFn: tvService.listQuotes })
}
