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
