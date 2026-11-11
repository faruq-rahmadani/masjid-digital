import type { Category, CategoryScope, Pocket, PocketKind } from '@/types/finance'

// Kas rules mirrored from internal/domain/finance.go, so the forms only offer what the API takes.

export const POCKET_KIND_LABEL: Record<PocketKind, string> = {
  umum: 'Umum',
  zakat_mal: 'Zakat',
  zakat_fitrah: 'Zakat',
  fidyah: 'Fidyah',
  terikat: 'Terikat',
}

export const POCKET_KIND_TONE = {
  umum: 'success',
  zakat_mal: 'gold',
  zakat_fitrah: 'gold',
  fidyah: 'gold',
  terikat: 'notice',
} as const satisfies Record<PocketKind, string>

const SCOPE: Record<PocketKind, CategoryScope> = {
  umum: 'umum_terikat',
  terikat: 'umum_terikat',
  zakat_mal: 'zakat',
  zakat_fitrah: 'zakat',
  fidyah: 'fidyah',
}

/** Zakat mal only into Zakat Mal, zakat fitrah only into Zakat Fitrah. */
const CODE_POCKET: Record<string, PocketKind> = {
  zakat_mal: 'zakat_mal',
  zakat_fitrah: 'zakat_fitrah',
}

export const isBuiltIn = (kind: PocketKind) => kind !== 'terikat'
export const isZakat = (kind: PocketKind) =>
  kind === 'zakat_mal' || kind === 'zakat_fitrah' || kind === 'fidyah'
/** Pindah dana into or out of this kind of kas (setor tunai within one kas is always fine). */
export const canMoveFunds = (kind: PocketKind) => kind === 'umum' || kind === 'terikat'
export const isPindah = (c: Category) => c.code === 'pindah_in' || c.code === 'pindah_out'
export const isOpen = (p: Pocket) => p.status === 'active' || p.status === 'target_reached'

export function categoryFits(category: Category, kind: PocketKind) {
  if (category.scope === 'all') return true
  if (category.scope !== SCOPE[kind]) return false
  const only = category.code ? CODE_POCKET[category.code] : undefined
  return !only || only === kind
}

/** The categories the transaction form offers for a kas: live, of the type, fitting the kas. */
export function formCategories(categories: Category[], kind: PocketKind, type: Category['type']) {
  return categories.filter(
    (c) => c.archived_at === null && c.type === type && !isPindah(c) && categoryFits(c, kind),
  )
}

/** Categories may be hidden when own, or an asnaf / fidyah recipient. */
export const canArchiveCategory = (c: Category) => !c.code || (c.type === 'out' && !isPindah(c))

const monthLong = new Intl.DateTimeFormat('id-ID', {
  month: 'long',
  year: 'numeric',
  timeZone: 'UTC',
})
const monthShort = new Intl.DateTimeFormat('id-ID', {
  month: 'short',
  year: 'numeric',
  timeZone: 'UTC',
})
const dayShort = new Intl.DateTimeFormat('id-ID', {
  day: 'numeric',
  month: 'short',
  timeZone: 'UTC',
})

/** "2026-11" → "November 2026" */
export const formatMonth = (month: string) => monthLong.format(new Date(`${month}-01T00:00:00Z`))
/** "2026-11" → "Nov 2026" */
export const formatMonthShort = (month: string) =>
  monthShort.format(new Date(`${month}-01T00:00:00Z`))
/** "2026-11-06" → "6 Nov" */
export const formatDayShort = (date: string) => dayShort.format(new Date(`${date}T00:00:00Z`))

/** "2026-11" ± n months */
export function addMonths(month: string, n: number) {
  const [y = 0, m = 1] = month.split('-').map(Number)
  const d = new Date(Date.UTC(y, m - 1 + n, 1))
  return `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, '0')}`
}

/** Every month from `first` to `last`, newest first. */
export function monthsBetween(first: string, last: string) {
  const months: string[] = []
  for (let m = last; m >= first && months.length < 120; m = addMonths(m, -1)) months.push(m)
  return months
}

/** The amount the form asks to confirm above (a mistyped zero costs a lot). */
export const CONFIRM_AMOUNT_ABOVE = 10_000_000
export const MAX_AMOUNT = 10_000_000_000
export const MAX_RECEIPT_BYTES = 1 << 20

/**
 * Shrinks a photo to a JPEG of at most `maxBytes`, halving the quality and then the size until it
 * fits. Receipts are stored in the database, so they stay small.
 */
export async function resizeImage(file: File, maxBytes = MAX_RECEIPT_BYTES): Promise<Blob> {
  if (file.size <= maxBytes && file.type === 'image/jpeg') return file
  const bitmap = await createImageBitmap(file)
  let scale = Math.min(1, 1600 / Math.max(bitmap.width, bitmap.height))
  for (let i = 0; i < 6; i++) {
    const canvas = document.createElement('canvas')
    canvas.width = Math.round(bitmap.width * scale)
    canvas.height = Math.round(bitmap.height * scale)
    canvas.getContext('2d')?.drawImage(bitmap, 0, 0, canvas.width, canvas.height)
    const blob = await new Promise<Blob | null>((resolve) =>
      canvas.toBlob(resolve, 'image/jpeg', 0.8),
    )
    if (blob && blob.size <= maxBytes) return blob
    scale *= 0.7
  }
  throw new Error('Foto terlalu besar, pilih foto lain')
}
