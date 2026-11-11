import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as financeService from '@/services/financeService'

// Every write moves balances somewhere else on screen (kas, akun, ringkasan), so a successful
// write refetches everything under ['finance'] instead of patching one list.
export const financeKeys = {
  all: ['finance'] as const,
  accounts: ['finance', 'accounts'] as const,
  pockets: ['finance', 'pockets'] as const,
  categories: ['finance', 'categories'] as const,
  bookkeeping: ['finance', 'bookkeeping'] as const,
  closings: ['finance', 'closings'] as const,
  tvSetting: ['finance', 'tv-setting'] as const,
  summary: (month: string) => ['finance', 'summary', month] as const,
  transactions: (month: string, deleted: boolean) =>
    ['finance', 'transactions', month, deleted] as const,
  revisions: (id: number) => ['finance', 'revisions', id] as const,
  receipt: (id: number) => ['finance', 'receipt', id] as const,
}

export const useAccounts = () =>
  useQuery({ queryKey: financeKeys.accounts, queryFn: financeService.listAccounts })
export const usePockets = () =>
  useQuery({ queryKey: financeKeys.pockets, queryFn: financeService.listPockets })
export const useCategories = () =>
  useQuery({ queryKey: financeKeys.categories, queryFn: financeService.listCategories })
export const useBookkeeping = () =>
  useQuery({ queryKey: financeKeys.bookkeeping, queryFn: financeService.getBookkeeping })
export const useClosings = () =>
  useQuery({ queryKey: financeKeys.closings, queryFn: financeService.listClosings })
export const useTvSetting = () =>
  useQuery({ queryKey: financeKeys.tvSetting, queryFn: financeService.getTvSetting })

export const useFinanceSummary = (month: string, enabled = true) =>
  useQuery({
    queryKey: financeKeys.summary(month),
    queryFn: () => financeService.getSummary(month),
    enabled,
  })

export const useTransactions = (month: string, includeDeleted: boolean) =>
  useQuery({
    queryKey: financeKeys.transactions(month, includeDeleted),
    queryFn: () => financeService.listTransactions(month, includeDeleted),
  })

export const useRevisions = (id: number) =>
  useQuery({
    queryKey: financeKeys.revisions(id),
    queryFn: () => financeService.listRevisions(id),
  })

export const useReceipt = (id: number, enabled: boolean) =>
  useQuery({
    queryKey: financeKeys.receipt(id),
    queryFn: () => financeService.getReceipt(id),
    enabled,
    staleTime: Number.POSITIVE_INFINITY,
  })

/** A finance write that refetches every finance query once it succeeds. */
export function useFinanceMutation<TVars, TResult>(fn: (vars: TVars) => Promise<TResult>) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: fn,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: financeKeys.all }),
  })
}

/** Accounts, kas, categories and closings by ID, plus the first date still open for changes. */
export function useFinanceLookups() {
  const accounts = useAccounts()
  const pockets = usePockets()
  const categories = useCategories()
  const closings = useClosings()
  const ready = accounts.data && pockets.data && categories.data && closings.data
  const error = accounts.error ?? pockets.error ?? categories.error ?? closings.error
  if (!ready) return { ready: false as const, error }
  const last = closings.data.at(-1)
  return {
    ready: true as const,
    error,
    accounts: accounts.data,
    pockets: pockets.data,
    categories: categories.data,
    closings: closings.data,
    account: new Map(accounts.data.map((a) => [a.id, a])),
    pocket: new Map(pockets.data.map((p) => [p.id, p])),
    category: new Map(categories.data.map((c) => [c.id, c])),
    /** YYYY-MM of the last closing (the opening before any month is closed), null before pembukuan. */
    lastClosed: last?.month ?? null,
  }
}

export type FinanceLookups = Extract<ReturnType<typeof useFinanceLookups>, { ready: true }>
