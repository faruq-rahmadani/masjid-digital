// Keuangan (E7). Money is integer Rupiah; every balance comes computed from the API.

export type AccountKind = 'cash' | 'bank'

/** Where money sits: Tunai or a bank account. */
export type Account = {
  id: number
  name: string
  kind: AccountKind
  /** Bank account number; "" for cash. */
  number: string
  archived_at: string | null
  created_by: number
  created_at: string
  updated_at: string
  balance: number
}

export type PocketKind = 'umum' | 'zakat_mal' | 'zakat_fitrah' | 'fidyah' | 'terikat'
export type PocketStatus = 'active' | 'target_reached' | 'done' | 'archived'
export type SurplusPlan = 'to_umum' | 'ask_donor'

/** A kas: whose money it is and what it may be spent on. */
export type Pocket = {
  id: number
  kind: PocketKind
  name: string
  purpose: string
  target_amount: number | null
  surplus_plan: SurplusPlan | null
  show_on_tv: boolean
  status: PocketStatus
  created_by: number
  created_at: string
  updated_at: string
  balance: number
  /** Terikat kas with a target only. */
  progress: { collected: number; target_reached: boolean } | null
}

export type CategoryType = 'in' | 'out'
export type CategoryScope = 'umum_terikat' | 'zakat' | 'fidyah' | 'all'

export type Category = {
  id: number
  name: string
  type: CategoryType
  scope: CategoryScope
  /** System categories only, e.g. asnaf_fakir, pindah_out. */
  code: string | null
  archived_at: string | null
  created_by: number
  created_at: string
  updated_at: string
  transaction_count: number
}

export type Transaction = {
  id: number
  kind: CategoryType
  amount: number
  /** YYYY-MM-DD, the mosque's date. */
  date: string
  account_id: number
  pocket_id: number
  category_id: number
  /** Shared by the rows of a split income or a Pindah dana pair. */
  group_id: string | null
  donor_consent: boolean
  corrects_id: number | null
  note: string
  donor_note: string
  has_photo: boolean
  created_by: number
  updated_by: number
  created_at: string
  updated_at: string
  deleted_at: string | null
  deleted_by: number
}

export type TransactionLine = { pocket_id: number; category_id: number; amount: number }

export type SaveTransactionRequest = {
  kind: CategoryType
  date: string
  account_id: number
  lines: TransactionLine[]
  note: string
  donor_note: string
  corrects_id?: number | null
  /** Required when editing. */
  reason?: string
}

export type MoveFundsRequest = {
  from_pocket_id: number
  from_account_id: number
  to_pocket_id: number
  to_account_id: number
  amount: number
  date: string
  note: string
  donor_consent: boolean
  reason?: string
}

export type TransactionRevision = {
  id: number
  transaction_id: number
  action: 'edit' | 'delete'
  previous: Transaction
  reason: string
  user_id: number
  created_at: string
}

export type Flow = { opening: number; in: number; out: number; moved: number; closing: number }

export type FinanceWarning = {
  kind: 'amil_cap' | 'fitrah_deadline'
  pocket_id: number
  message: string
}

export type FinanceSummary = {
  month: string
  today: string
  closed: boolean
  total_in: number
  total_out: number
  today_in: number
  today_out: number
  /** Every kas together at the end of the month (now for the current month). */
  balance: number
  pockets: (Flow & { pocket_id: number; today_in: number; today_out: number })[]
  accounts: (Flow & { account_id: number })[]
  warnings: FinanceWarning[]
}

export type BookClosing = {
  id: number
  month: string
  kind: 'opening' | 'closing'
  total_in: number
  total_out: number
  closing_balance: number
  accounts: { account_id: number; closing_balance: number; counted_balance: number | null }[]
  pockets: { pocket_id: number; closing_balance: number }[]
  closed_by: number
  closed_at: string
}

export type Bookkeeping = { started: boolean; start_month: string; opening: BookClosing | null }

export type OpeningBalance = { id: number; amount: number }

export type TvPeriod = 'today' | 'week' | 'month'

export type FinanceDisplay = {
  period: TvPeriod
  from: string
  to: string
  pockets: (Flow & { id: number; name: string; kind: PocketKind })[]
}
