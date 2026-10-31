import { describe, expect, it } from 'vitest'
import { initials } from './text'

describe('text', () => {
  it('builds initials from names, ignoring placeholder brackets', () => {
    expect(initials('[Ketua DKM]')).toBe('KD')
    expect(initials('Siti')).toBe('S')
    expect(initials('Ahmad bin Abdullah')).toBe('AB')
  })
})
