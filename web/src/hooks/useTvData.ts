import { useQuery } from '@tanstack/react-query'
import * as tvService from '@/services/tvService'

// The TV reloads these on the server clock's 5-minute marks (useAlignedRefresh); TanStack Query
// keeps `data` when a reload fails, so a network blip never blanks the screen.

/** The kas header; a finance change also reaches the TV through its live connection. */
export function useCashDisplay() {
  return useQuery({ queryKey: ['tv', 'cash'], queryFn: tvService.getCashDisplay })
}

export function useQuotes() {
  return useQuery({ queryKey: ['tv', 'quotes'], queryFn: tvService.listQuotes })
}
