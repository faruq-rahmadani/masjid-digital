import { useQuery } from '@tanstack/react-query'
import * as tvService from '@/services/tvService'

// The TV runs unattended for days: refetch on a timer, and keep the last good data on failure
// (TanStack Query keeps `data` when a refetch fails), so a network blip never blanks the screen.
const REFRESH = { refetchInterval: 5 * 60_000, refetchIntervalInBackground: true } as const

export function useCashSummary() {
  return useQuery({ queryKey: ['tv', 'cash'], queryFn: tvService.getCashSummary, ...REFRESH })
}

export function useSlides() {
  return useQuery({ queryKey: ['tv', 'slides'], queryFn: tvService.listSlides, ...REFRESH })
}

export function useQuotes() {
  return useQuery({ queryKey: ['tv', 'quotes'], queryFn: tvService.listQuotes, ...REFRESH })
}

export function useIqamahSettings() {
  return useQuery({ queryKey: ['tv', 'iqamah'], queryFn: tvService.getIqamahSettings, ...REFRESH })
}

export function useDutyRoster() {
  return useQuery({ queryKey: ['tv', 'duty'], queryFn: tvService.getDutyRoster, ...REFRESH })
}
