// Money is integer Rupiah end to end (docs/product/prd.md#data-model).

const rupiah = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
})
const grouped = new Intl.NumberFormat('id-ID', { maximumFractionDigits: 0 })

/** 92058114 → "Rp 92.058.114" */
export function formatRupiah(amount: number) {
  // Intl puts a no-break space after "Rp"; keep it so the amount never wraps from its symbol.
  return rupiah.format(amount)
}

/** 750000 → "750.000" */
export function formatNumber(amount: number) {
  return grouped.format(amount)
}

export type CashFlow = 'in' | 'out'

/** (750000, 'in') → "+ 750.000", (626667, 'out') → "− 626.667" */
export function formatSignedAmount(amount: number, flow: CashFlow, { withSymbol = false } = {}) {
  const value = withSymbol ? formatRupiah(Math.abs(amount)) : formatNumber(Math.abs(amount))
  return `${flow === 'in' ? '+' : '−'} ${value}`
}

/** "[Ketua DKM]" → "KD", "Nama Bendahara" → "NB" */
export function initials(name: string) {
  return name
    .replace(/[^\p{L}\s]/gu, ' ')
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((word) => word.charAt(0).toUpperCase())
    .join('')
}

const longDate = new Intl.DateTimeFormat('id-ID', {
  weekday: 'long',
  day: 'numeric',
  month: 'short',
  year: 'numeric',
})
const shortDate = new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short' })

/** → "Rabu, 28 Okt 2026" */
export function formatDate(date: Date) {
  return longDate.format(date)
}

/** → "28 Okt" */
export function formatShortDate(date: Date) {
  return shortDate.format(date)
}

/** → "18:15". id-ID uses "18.15"; the mockup and TV use a colon. */
export function formatTime(date: Date) {
  const hh = String(date.getHours()).padStart(2, '0')
  const mm = String(date.getMinutes()).padStart(2, '0')
  return `${hh}:${mm}`
}
