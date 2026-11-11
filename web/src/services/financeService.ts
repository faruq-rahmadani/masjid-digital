import { apiDelete, apiGet, apiPost, apiPut, type DataResponse } from '@/api/client'
import type {
  Account,
  AccountKind,
  BookClosing,
  Bookkeeping,
  Category,
  CategoryType,
  FinanceDisplay,
  FinanceSummary,
  MoveFundsRequest,
  OpeningBalance,
  Pocket,
  SaveTransactionRequest,
  SurplusPlan,
  Transaction,
  TransactionRevision,
  TvPeriod,
} from '@/types/finance'

// Keuangan (E7): one function per /v1/finance endpoint.

const base = '/v1/finance'

async function data<T>(request: Promise<DataResponse<T>>) {
  return (await request).data
}

export const listAccounts = () => data(apiGet<DataResponse<Account[]>>(`${base}/accounts`))
export const createAccount = (body: { name: string; kind: AccountKind; number: string }) =>
  data(apiPost<DataResponse<Account>>(`${base}/accounts`, body))
export const updateAccount = (id: number, body: { name: string; number: string }) =>
  data(apiPut<DataResponse<Account>>(`${base}/accounts/${id}`, body))
export const archiveAccount = (id: number) =>
  data(apiPost<DataResponse<Account>>(`${base}/accounts/${id}/archive`))
export const restoreAccount = (id: number) =>
  data(apiDelete<DataResponse<Account>>(`${base}/accounts/${id}/archive`))

export type PocketFields = {
  name: string
  purpose: string
  target_amount: number | null
  surplus_plan: SurplusPlan
  show_on_tv: boolean
}

export const listPockets = () => data(apiGet<DataResponse<Pocket[]>>(`${base}/pockets`))
export const createPocket = (body: PocketFields) =>
  data(apiPost<DataResponse<Pocket>>(`${base}/pockets`, body))
/** A built-in kas only takes show_on_tv; the rest is ignored. */
export const updatePocket = (id: number, body: PocketFields) =>
  data(apiPut<DataResponse<Pocket>>(`${base}/pockets/${id}`, body))
export const archivePocket = (id: number) =>
  data(apiPost<DataResponse<Pocket>>(`${base}/pockets/${id}/archive`))
export const restorePocket = (id: number) =>
  data(apiDelete<DataResponse<Pocket>>(`${base}/pockets/${id}/archive`))
export const finishPocket = (
  id: number,
  body: { account_id: number; donor_consent: boolean; note: string },
) => data(apiPost<DataResponse<Pocket>>(`${base}/pockets/${id}/finish`, body))

export const listCategories = () => data(apiGet<DataResponse<Category[]>>(`${base}/categories`))
export const createCategory = (body: { name: string; type: CategoryType }) =>
  data(apiPost<DataResponse<Category>>(`${base}/categories`, body))
export const renameCategory = (id: number, name: string) =>
  data(apiPut<DataResponse<Category>>(`${base}/categories/${id}`, { name }))
export const deleteCategory = (id: number) => apiDelete(`${base}/categories/${id}`)
export const archiveCategory = (id: number) =>
  data(apiPost<DataResponse<Category>>(`${base}/categories/${id}/archive`))
export const restoreCategory = (id: number) =>
  data(apiDelete<DataResponse<Category>>(`${base}/categories/${id}/archive`))

export const getBookkeeping = () => data(apiGet<DataResponse<Bookkeeping>>(`${base}/bookkeeping`))
export const startBookkeeping = (body: {
  start_month: string
  accounts: OpeningBalance[]
  pockets: OpeningBalance[]
}) => data(apiPost<DataResponse<Bookkeeping>>(`${base}/bookkeeping`, body))

export const getSummary = (month: string) =>
  data(apiGet<DataResponse<FinanceSummary>>(`${base}/summary`, { params: { month } }))

export const listTransactions = (month: string, includeDeleted: boolean) =>
  data(
    apiGet<DataResponse<Transaction[]>>(`${base}/transactions`, {
      params: { month, include_deleted: includeDeleted },
    }),
  )
export const createTransaction = (body: SaveTransactionRequest) =>
  data(apiPost<DataResponse<Transaction[]>>(`${base}/transactions`, body))
export const updateTransaction = (id: number, body: SaveTransactionRequest) =>
  data(apiPut<DataResponse<Transaction[]>>(`${base}/transactions/${id}`, body))
export const deleteTransaction = (id: number, reason: string) =>
  apiDelete(`${base}/transactions/${id}`, { data: { reason } })
export const listRevisions = (id: number) =>
  data(apiGet<DataResponse<TransactionRevision[]>>(`${base}/transactions/${id}/revisions`))

/** The receipt photo, already resized to ≤ 1 MB in the browser. */
export async function uploadReceipt(id: number, file: Blob) {
  const body = new FormData()
  body.append('file', file, 'bukti.jpg')
  return apiPut(`${base}/transactions/${id}/receipt`, body)
}

/** The receipt needs the token, so it is fetched as a blob for an object URL. */
export const getReceipt = (id: number) =>
  apiGet<Blob>(`${base}/transactions/${id}/receipt`, { responseType: 'blob' })

export const moveFunds = (body: MoveFundsRequest) =>
  data(apiPost<DataResponse<Transaction[]>>(`${base}/transfers`, body))
export const updateMoveFunds = (groupId: string, body: MoveFundsRequest) =>
  data(apiPut<DataResponse<Transaction[]>>(`${base}/transfers/${groupId}`, body))

export const listClosings = () => data(apiGet<DataResponse<BookClosing[]>>(`${base}/closings`))
export const closeMonth = (body: {
  month: string
  accounts: { account_id: number; counted: number }[]
}) => data(apiPost<DataResponse<BookClosing>>(`${base}/closings`, body))

export const getTvSetting = () =>
  data(apiGet<DataResponse<{ period: TvPeriod }>>(`${base}/tv-setting`))
export const updateTvSetting = (period: TvPeriod) =>
  data(apiPut<DataResponse<{ period: TvPeriod }>>(`${base}/tv-setting`, { period }))

/** The TV kas header; public, the TV has no login. */
export const getFinanceDisplay = () =>
  data(apiGet<DataResponse<FinanceDisplay>>('/public/v1/finance/display'))
