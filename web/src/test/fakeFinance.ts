import { ApiError } from '@/api/client'
import type { PocketFields } from '@/services/financeService'
import type {
  Account,
  AccountKind,
  BookClosing,
  Category,
  CategoryType,
  FinanceSummary,
  MoveFundsRequest,
  OpeningBalance,
  Pocket,
  SaveTransactionRequest,
  Transaction,
  TransactionRevision,
  TvPeriod,
} from '@/types/finance'
import { loadSession } from '@/utils/auth/token'

// A strict, stateful fake of the Keuangan API (financeService) for page tests. It keeps the
// rules the screens rely on (roles, no balance below Rp 0, zakat closed, terikat consent,
// reasons) and computes balances the way the Go usecase does: opening + live transactions.

/** The fake's clock: every summary is "today" on this date. */
export const FAKE_TODAY = '2026-11-11'
const STAMP = '2026-11-01T00:00:00Z'

type State = {
  accounts: Omit<Account, 'balance'>[]
  pockets: Omit<Pocket, 'balance' | 'progress'>[]
  categories: Omit<Category, 'transaction_count'>[]
  txs: Transaction[]
  closings: BookClosing[]
  revisions: TransactionRevision[]
  tv: TvPeriod
  nextId: number
}

const category = (
  id: number,
  name: string,
  type: CategoryType,
  scope: Category['scope'],
  code: string | null = null,
): State['categories'][number] => ({
  id,
  name,
  type,
  scope,
  code,
  archived_at: null,
  created_by: 0,
  created_at: STAMP,
  updated_at: STAMP,
})

const pocket = (
  id: number,
  kind: Pocket['kind'],
  name: string,
  over: Partial<Pocket> = {},
): State['pockets'][number] => ({
  id,
  kind,
  name,
  purpose: `Peruntukan ${name}`,
  target_amount: null,
  surplus_plan: null,
  show_on_tv: kind === 'umum',
  status: 'active',
  created_by: 0,
  created_at: STAMP,
  updated_at: STAMP,
  ...over,
})

/** Opening at the end of October: Tunai 2.000.000 in Kas Umum, BSI 1.500.000 in Zakat Mal. */
const OPENING: BookClosing = {
  id: 1,
  month: '2026-10',
  kind: 'opening',
  total_in: 0,
  total_out: 0,
  closing_balance: 3_500_000,
  accounts: [
    { account_id: 1, closing_balance: 2_000_000, counted_balance: null },
    { account_id: 2, closing_balance: 1_500_000, counted_balance: null },
  ],
  pockets: [
    { pocket_id: 1, closing_balance: 2_000_000 },
    { pocket_id: 2, closing_balance: 1_500_000 },
  ],
  closed_by: 3,
  closed_at: STAMP,
}

const seed = (): State => ({
  accounts: [
    {
      id: 1,
      name: 'Tunai',
      kind: 'cash',
      number: '',
      archived_at: null,
      created_by: 0,
      created_at: STAMP,
      updated_at: STAMP,
    },
    {
      id: 2,
      name: 'BSI Masjid',
      kind: 'bank',
      number: '7123',
      archived_at: null,
      created_by: 3,
      created_at: STAMP,
      updated_at: STAMP,
    },
  ],
  pockets: [
    pocket(1, 'umum', 'Kas Umum'),
    pocket(2, 'zakat_mal', 'Zakat Mal'),
    pocket(3, 'zakat_fitrah', 'Zakat Fitrah'),
    pocket(4, 'fidyah', 'Fidyah'),
    pocket(5, 'terikat', 'Renovasi Atap', {
      target_amount: 3_000_000,
      surplus_plan: 'to_umum',
      show_on_tv: true,
    }),
  ],
  categories: [
    category(10, 'Infaq Jumat', 'in', 'umum_terikat'),
    category(11, 'Listrik', 'out', 'umum_terikat'),
    category(12, 'Zakat mal', 'in', 'zakat', 'zakat_mal'),
    category(13, 'Fakir', 'out', 'zakat', 'asnaf_fakir'),
    category(14, 'Amil', 'out', 'zakat', 'asnaf_amil'),
    category(20, 'Pindah dana', 'out', 'all', 'pindah_out'),
    category(21, 'Pindah dana', 'in', 'all', 'pindah_in'),
  ],
  txs: [],
  closings: [OPENING],
  revisions: [],
  tv: 'month',
  nextId: 100,
})

let state = seed()

/** Read the fake's state, or start a fresh mosque that has not started pembukuan. */
export const fakeFinance = {
  get state() {
    return state
  },
  notStarted() {
    state.closings = []
  },
}

export function resetFakeFinance() {
  state = seed()
}

const fail = (status: number, code: string, message: string): never => {
  throw new ApiError(status, { code, message })
}

function requireRole(...roles: string[]) {
  const session = loadSession() ?? fail(401, 'UNAUTHENTICATED', 'silakan login terlebih dahulu')
  if (!roles.includes(session.role)) fail(403, 'FORBIDDEN', 'anda tidak memiliki akses')
}
const requireAny = () => requireRole('super_admin', 'takmir', 'bendahara')
const requireEditor = () => requireRole('super_admin', 'bendahara')

const live = () => state.txs.filter((t) => t.deleted_at === null)
const signed = (t: Transaction) => (t.kind === 'in' ? t.amount : -t.amount)
const isPindahId = (id: number) => id === 20 || id === 21

function balances() {
  const accounts = new Map<number, number>()
  const pockets = new Map<number, number>()
  const last = state.closings.at(-1)
  for (const a of last?.accounts ?? []) accounts.set(a.account_id, a.closing_balance)
  for (const p of last?.pockets ?? []) pockets.set(p.pocket_id, p.closing_balance)
  for (const t of live().filter((t) => !last || t.date.slice(0, 7) > last.month)) {
    accounts.set(t.account_id, (accounts.get(t.account_id) ?? 0) + signed(t))
    pockets.set(t.pocket_id, (pockets.get(t.pocket_id) ?? 0) + signed(t))
  }
  return { accounts, pockets }
}

function checkBalances(rows: Transaction[], removing: Transaction[] = []) {
  const b = balances()
  for (const r of removing) {
    b.accounts.set(r.account_id, (b.accounts.get(r.account_id) ?? 0) - signed(r))
    b.pockets.set(r.pocket_id, (b.pockets.get(r.pocket_id) ?? 0) - signed(r))
  }
  for (const r of rows) {
    b.accounts.set(r.account_id, (b.accounts.get(r.account_id) ?? 0) + signed(r))
    b.pockets.set(r.pocket_id, (b.pockets.get(r.pocket_id) ?? 0) + signed(r))
  }
  for (const [id, v] of b.pockets) {
    const name = state.pockets.find((p) => p.id === id)?.name
    if (v < 0) fail(422, 'BUSINESS_RULE', `saldo kas ${name} tidak cukup`)
  }
  for (const [id, v] of b.accounts) {
    const name = state.accounts.find((a) => a.id === id)?.name
    if (v < 0) fail(422, 'BUSINESS_RULE', `saldo akun ${name} tidak cukup`)
  }
}

function row(over: Partial<Transaction>): Transaction {
  const id = state.nextId++
  return {
    id,
    kind: 'in',
    amount: 0,
    date: FAKE_TODAY,
    account_id: 1,
    pocket_id: 1,
    category_id: 10,
    group_id: null,
    donor_consent: false,
    corrects_id: null,
    note: '',
    donor_note: '',
    has_photo: false,
    created_by: 3,
    updated_by: 3,
    created_at: STAMP,
    updated_at: STAMP,
    deleted_at: null,
    deleted_by: 0,
    ...over,
  }
}

const findPocket = (id: number) =>
  state.pockets.find((p) => p.id === id) ?? fail(404, 'NOT_FOUND', 'kas tidak ditemukan')
const findAccount = (id: number) =>
  state.accounts.find((a) => a.id === id) ?? fail(404, 'NOT_FOUND', 'akun tidak ditemukan')
const findCategory = (id: number) =>
  state.categories.find((c) => c.id === id) ?? fail(404, 'NOT_FOUND', 'kategori tidak ditemukan')

const withBalance = (a: State['accounts'][number]): Account => ({
  ...a,
  balance: balances().accounts.get(a.id) ?? 0,
})
function withPocketBalance(p: State['pockets'][number]): Pocket {
  const collected = live()
    .filter((t) => t.pocket_id === p.id && t.kind === 'in')
    .reduce((s, t) => s + t.amount, 0)
  return {
    ...p,
    balance: balances().pockets.get(p.id) ?? 0,
    progress:
      p.kind === 'terikat' && p.target_amount !== null
        ? { collected, target_reached: collected >= p.target_amount }
        : null,
  }
}

export const fakeFinanceService = {
  listAccounts: async () => {
    requireAny()
    return state.accounts.map(withBalance)
  },
  createAccount: async (body: { name: string; kind: AccountKind; number: string }) => {
    requireEditor()
    const account = {
      id: state.nextId++,
      ...body,
      archived_at: null,
      created_by: 3,
      created_at: STAMP,
      updated_at: STAMP,
    }
    state.accounts.push(account)
    return withBalance(account)
  },
  updateAccount: async (id: number, body: { name: string; number: string }) => {
    requireEditor()
    Object.assign(findAccount(id), body)
    return withBalance(findAccount(id))
  },
  archiveAccount: async (id: number) => {
    requireEditor()
    if ((balances().accounts.get(id) ?? 0) !== 0)
      fail(422, 'BUSINESS_RULE', 'akun hanya bisa diarsipkan saat saldonya Rp 0')
    findAccount(id).archived_at = STAMP
    return withBalance(findAccount(id))
  },
  restoreAccount: async (id: number) => {
    requireEditor()
    findAccount(id).archived_at = null
    return withBalance(findAccount(id))
  },

  listPockets: async () => {
    requireAny()
    return state.pockets.map(withPocketBalance)
  },
  createPocket: async (body: PocketFields) => {
    requireEditor()
    if (state.pockets.some((p) => p.name === body.name))
      fail(409, 'ALREADY_EXISTS', 'nama kas sudah dipakai')
    const created = pocket(state.nextId++, 'terikat', body.name, { ...body, created_by: 3 })
    state.pockets.push(created)
    return withPocketBalance(created)
  },
  updatePocket: async (id: number, body: PocketFields) => {
    requireEditor()
    const p = findPocket(id)
    if (p.kind !== 'terikat') p.show_on_tv = body.show_on_tv
    if (p.kind === 'terikat') Object.assign(p, body)
    return withPocketBalance(p)
  },
  archivePocket: async (id: number) => {
    requireEditor()
    const p = findPocket(id)
    if (p.kind !== 'terikat') fail(422, 'BUSINESS_RULE', 'kas bawaan tidak bisa diarsipkan')
    if ((balances().pockets.get(id) ?? 0) !== 0)
      fail(422, 'BUSINESS_RULE', 'kas hanya bisa diarsipkan saat saldonya Rp 0')
    p.status = 'archived'
    return withPocketBalance(p)
  },
  restorePocket: async (id: number) => {
    requireEditor()
    findPocket(id).status = 'active'
    return withPocketBalance(findPocket(id))
  },
  finishPocket: async (
    id: number,
    body: { account_id: number; donor_consent: boolean; note: string },
  ) => {
    requireEditor()
    const p = findPocket(id)
    const remainder = balances().pockets.get(id) ?? 0
    if (remainder > 0) {
      const group = `finish-${id}`
      state.txs.push(
        row({
          kind: 'out',
          amount: remainder,
          pocket_id: id,
          account_id: body.account_id,
          category_id: 20,
          group_id: group,
        }),
        row({
          kind: 'in',
          amount: remainder,
          pocket_id: 1,
          account_id: body.account_id,
          category_id: 21,
          group_id: group,
        }),
      )
    }
    p.status = 'archived'
    return withPocketBalance(p)
  },

  listCategories: async () => {
    requireAny()
    return state.categories.map((c) => ({
      ...c,
      transaction_count: live().filter((t) => t.category_id === c.id).length,
    }))
  },
  createCategory: async (body: { name: string; type: CategoryType }) => {
    requireEditor()
    const created = category(state.nextId++, body.name, body.type, 'umum_terikat')
    state.categories.push(created)
    return { ...created, transaction_count: 0 }
  },
  renameCategory: async (id: number, name: string) => {
    requireEditor()
    const c = findCategory(id)
    if (c.code) fail(422, 'BUSINESS_RULE', 'kategori sistem tidak bisa diubah namanya')
    c.name = name
    return { ...c, transaction_count: 0 }
  },
  deleteCategory: async (id: number) => {
    requireEditor()
    findCategory(id)
    state.categories = state.categories.filter((c) => c.id !== id)
    return { message: 'kategori dihapus' }
  },
  archiveCategory: async (id: number) => {
    requireEditor()
    findCategory(id).archived_at = STAMP
    return { ...findCategory(id), transaction_count: 0 }
  },
  restoreCategory: async (id: number) => {
    requireEditor()
    findCategory(id).archived_at = null
    return { ...findCategory(id), transaction_count: 0 }
  },

  getBookkeeping: async () => {
    requireAny()
    const opening = state.closings.find((c) => c.kind === 'opening') ?? null
    return { started: opening !== null, start_month: opening ? '2026-11' : '', opening }
  },
  startBookkeeping: async (body: {
    start_month: string
    accounts: OpeningBalance[]
    pockets: OpeningBalance[]
  }) => {
    requireEditor()
    if (state.closings.length) fail(422, 'BUSINESS_RULE', 'pembukuan sudah dimulai')
    const total = (l: OpeningBalance[]) => l.reduce((s, b) => s + b.amount, 0)
    if (total(body.accounts) !== total(body.pockets))
      fail(400, 'INVALID_INPUT', 'total saldo akun harus sama dengan total saldo kas')
    state.closings = [
      {
        ...OPENING,
        accounts: body.accounts.map((a) => ({
          account_id: a.id,
          closing_balance: a.amount,
          counted_balance: null,
        })),
        pockets: body.pockets.map((p) => ({ pocket_id: p.id, closing_balance: p.amount })),
      },
    ]
    return fakeFinanceService.getBookkeeping()
  },

  getSummary: async (month: string): Promise<FinanceSummary> => {
    requireAny()
    const m = month || FAKE_TODAY.slice(0, 7)
    const b = balances()
    const inMonth = live().filter((t) => t.date.startsWith(m) && !isPindahId(t.category_id))
    const today = inMonth.filter((t) => t.date === FAKE_TODAY)
    const sum = (l: Transaction[], kind: CategoryType) =>
      l.filter((t) => t.kind === kind).reduce((s, t) => s + t.amount, 0)
    return {
      month: m,
      today: FAKE_TODAY,
      closed: m <= (state.closings.at(-1)?.month ?? ''),
      total_in: sum(inMonth, 'in'),
      total_out: sum(inMonth, 'out'),
      today_in: sum(today, 'in'),
      today_out: sum(today, 'out'),
      balance: [...b.pockets.values()].reduce((s, v) => s + v, 0),
      pockets: state.pockets.map((p) => {
        const mine = inMonth.filter((t) => t.pocket_id === p.id)
        const mineToday = today.filter((t) => t.pocket_id === p.id)
        return {
          pocket_id: p.id,
          opening: 0,
          in: sum(mine, 'in'),
          out: sum(mine, 'out'),
          moved: 0,
          closing: b.pockets.get(p.id) ?? 0,
          today_in: sum(mineToday, 'in'),
          today_out: sum(mineToday, 'out'),
        }
      }),
      accounts: state.accounts.map((a) => ({
        account_id: a.id,
        opening: 0,
        in: 0,
        out: 0,
        moved: 0,
        closing: b.accounts.get(a.id) ?? 0,
      })),
      warnings: [],
    }
  },

  listTransactions: async (month: string, includeDeleted: boolean) => {
    requireAny()
    return state.txs.filter(
      (t) => t.date.startsWith(month) && (includeDeleted || t.deleted_at === null),
    )
  },
  createTransaction: async (body: SaveTransactionRequest) => {
    requireEditor()
    if (state.closings.length === 0) fail(422, 'BUSINESS_RULE', 'mulai pembukuan dulu di menu Akun')
    const group = body.lines.length > 1 ? `split-${state.nextId}` : null
    const rows = body.lines.map((l) =>
      row({
        kind: body.kind,
        amount: l.amount,
        date: body.date,
        account_id: body.account_id,
        pocket_id: l.pocket_id,
        category_id: l.category_id,
        group_id: group,
        note: body.note,
        donor_note: body.donor_note,
      }),
    )
    checkBalances(rows)
    state.txs.push(...rows)
    return rows
  },
  updateTransaction: async (id: number, body: SaveTransactionRequest) => {
    requireEditor()
    if (!body.reason) fail(400, 'INVALID_INPUT', 'alasan perubahan wajib diisi')
    const old =
      state.txs.find((t) => t.id === id) ?? fail(404, 'NOT_FOUND', 'transaksi tidak ditemukan')
    const line = body.lines[0] ?? fail(400, 'INVALID_INPUT', 'pilih kas dan nominal')
    const next = {
      ...old,
      amount: line.amount,
      category_id: line.category_id,
      pocket_id: line.pocket_id,
      date: body.date,
      note: body.note,
      updated_at: '2026-11-11T08:00:00Z',
    }
    checkBalances([next], [old])
    state.revisions.push({
      id: state.nextId++,
      transaction_id: id,
      action: 'edit',
      previous: old,
      reason: body.reason ?? '',
      user_id: 3,
      created_at: STAMP,
    })
    state.txs = state.txs.map((t) => (t.id === id ? next : t))
    return [next]
  },
  deleteTransaction: async (id: number, reason: string) => {
    requireEditor()
    if (!reason.trim()) fail(400, 'INVALID_INPUT', 'alasan penghapusan wajib diisi')
    const tx =
      state.txs.find((t) => t.id === id) ?? fail(404, 'NOT_FOUND', 'transaksi tidak ditemukan')
    const group = tx.group_id ? state.txs.filter((t) => t.group_id === tx.group_id) : [tx]
    checkBalances([], group)
    for (const t of group) {
      state.revisions.push({
        id: state.nextId++,
        transaction_id: t.id,
        action: 'delete',
        previous: t,
        reason,
        user_id: 3,
        created_at: STAMP,
      })
      t.deleted_at = '2026-11-11T08:00:00Z'
    }
    return { message: 'transaksi dihapus' }
  },
  listRevisions: async (id: number) => {
    requireAny()
    return state.revisions.filter((r) => r.transaction_id === id)
  },
  uploadReceipt: async (id: number) => {
    requireEditor()
    const tx =
      state.txs.find((t) => t.id === id) ?? fail(404, 'NOT_FOUND', 'transaksi tidak ditemukan')
    tx.has_photo = true
    return { message: 'foto bukti disimpan' }
  },
  getReceipt: async () => new Blob(['x'], { type: 'image/png' }),
  moveFunds: async (body: MoveFundsRequest) => {
    requireEditor()
    const from = findPocket(body.from_pocket_id)
    const to = findPocket(body.to_pocket_id)
    const movable = (k: Pocket['kind']) => k === 'umum' || k === 'terikat'
    if (from.id !== to.id && (!movable(from.kind) || !movable(to.kind)))
      fail(422, 'BUSINESS_RULE', 'dana zakat dan fidyah tidak bisa dipindah ke atau dari kas lain')
    if (from.id !== to.id && from.kind === 'terikat' && (!body.donor_consent || !body.note))
      fail(
        422,
        'BUSINESS_RULE',
        'dana kas terikat hanya bisa dipindah dengan alasan dan persetujuan donatur',
      )
    const group = `move-${state.nextId}`
    const rows = [
      row({
        kind: 'out',
        amount: body.amount,
        date: body.date,
        pocket_id: from.id,
        account_id: body.from_account_id,
        category_id: 20,
        group_id: group,
        note: body.note,
        donor_consent: body.donor_consent,
      }),
      row({
        kind: 'in',
        amount: body.amount,
        date: body.date,
        pocket_id: to.id,
        account_id: body.to_account_id,
        category_id: 21,
        group_id: group,
        note: body.note,
        donor_consent: body.donor_consent,
      }),
    ]
    checkBalances(rows)
    state.txs.push(...rows)
    return rows
  },
  updateMoveFunds: async () => fail(500, 'INTERNAL', 'not faked'),

  listClosings: async () => {
    requireAny()
    return state.closings
  },
  closeMonth: async (body: {
    month: string
    accounts: { account_id: number; counted: number }[]
  }) => {
    requireEditor()
    const closing: BookClosing = {
      ...OPENING,
      id: state.nextId++,
      kind: 'closing',
      month: body.month,
    }
    state.closings.push(closing)
    return closing
  },
  getTvSetting: async () => ({ period: state.tv }),
  updateTvSetting: async (period: TvPeriod) => {
    requireEditor()
    state.tv = period
    return { period }
  },
  getFinanceDisplay: async () => ({
    period: state.tv,
    from: '2026-11-01',
    to: FAKE_TODAY,
    pockets: [],
  }),
}
