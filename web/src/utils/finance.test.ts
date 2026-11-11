import { describe, expect, it } from 'vitest'
import type { Category } from '@/types/finance'
import {
  addMonths,
  canArchiveCategory,
  categoryFits,
  formatMonth,
  formCategories,
  monthsBetween,
} from './finance'

const category = (over: Partial<Category>): Category => ({
  id: 1,
  name: 'Infaq',
  type: 'in',
  scope: 'umum_terikat',
  code: null,
  archived_at: null,
  created_by: 0,
  created_at: '',
  updated_at: '',
  transaction_count: 0,
  ...over,
})

describe('categoryFits', () => {
  it('follows the kas kind', () => {
    expect(categoryFits(category({}), 'umum')).toBe(true)
    expect(categoryFits(category({}), 'zakat_mal')).toBe(false)
    const zakatMal = category({ scope: 'zakat', code: 'zakat_mal' })
    expect(categoryFits(zakatMal, 'zakat_mal')).toBe(true)
    expect(categoryFits(zakatMal, 'zakat_fitrah')).toBe(false)
    expect(categoryFits(category({ scope: 'zakat', code: 'asnaf_fakir' }), 'zakat_fitrah')).toBe(
      true,
    )
  })
})

describe('formCategories', () => {
  it('drops archived, Pindah dana and the other type', () => {
    const list = [
      category({ id: 1 }),
      category({ id: 2, archived_at: '2026-11-01T00:00:00Z' }),
      category({ id: 3, scope: 'all', code: 'pindah_in' }),
      category({ id: 4, type: 'out' }),
    ]
    expect(formCategories(list, 'umum', 'in').map((c) => c.id)).toEqual([1])
  })
})

describe('canArchiveCategory', () => {
  it('lets asnaf go but keeps zakat income and Pindah dana', () => {
    expect(canArchiveCategory(category({ type: 'out', scope: 'zakat', code: 'asnaf_amil' }))).toBe(
      true,
    )
    expect(canArchiveCategory(category({ scope: 'zakat', code: 'zakat_mal' }))).toBe(false)
    expect(canArchiveCategory(category({ type: 'out', scope: 'all', code: 'pindah_out' }))).toBe(
      false,
    )
  })
})

describe('months', () => {
  it('steps over years and lists newest first', () => {
    expect(addMonths('2026-01', -1)).toBe('2025-12')
    expect(monthsBetween('2026-10', '2026-12')).toEqual(['2026-12', '2026-11', '2026-10'])
    expect(formatMonth('2026-11')).toBe('November 2026')
  })
})
