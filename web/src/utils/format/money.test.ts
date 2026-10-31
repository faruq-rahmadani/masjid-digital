import { describe, expect, it } from 'vitest'
import { formatNumber, formatRupiah, formatSignedAmount } from './money'

// Intl separates "Rp" from the digits with a no-break space.
const nbsp = ' '

describe('money', () => {
  it('formats Rupiah without decimals', () => {
    expect(formatRupiah(92_058_114)).toBe(`Rp${nbsp}92.058.114`)
    expect(formatRupiah(0)).toBe(`Rp${nbsp}0`)
  })

  it('groups thousands with dots', () => {
    expect(formatNumber(1_250_000)).toBe('1.250.000')
  })

  it('signs amounts by cash flow using a true minus', () => {
    expect(formatSignedAmount(750_000, 'in')).toBe('+ 750.000')
    expect(formatSignedAmount(626_667, 'out')).toBe('− 626.667')
    expect(formatSignedAmount(-500, 'out', { withSymbol: true })).toBe(`− Rp${nbsp}500`)
  })
})
