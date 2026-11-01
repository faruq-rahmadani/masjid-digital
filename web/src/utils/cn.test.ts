import { describe, expect, it } from 'vitest'
import { cn } from './cn'

describe('cn', () => {
  it('keeps a TV size next to a text color', () => {
    expect(cn('text-tv-26', 'text-tv-tile-label')).toBe('text-tv-26 text-tv-tile-label')
    expect(cn('text-tv-28', 'text-income')).toBe('text-tv-28 text-income')
  })

  it('still lets a later size win', () => {
    expect(cn('text-tv-26', 'text-tv-30')).toBe('text-tv-30')
    expect(cn('rounded-tv', 'rounded-xl')).toBe('rounded-xl')
  })
})
