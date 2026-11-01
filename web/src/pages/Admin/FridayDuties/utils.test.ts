import { describe, expect, it } from 'vitest'
import { nextFridays } from './utils'

describe('nextFridays', () => {
  it('starts from the coming Friday, or today on a Friday', () => {
    expect(nextFridays('2026-11-02', 3)).toEqual(['2026-11-06', '2026-11-13', '2026-11-20'])
    expect(nextFridays('2026-11-06', 2)).toEqual(['2026-11-06', '2026-11-13'])
    expect(nextFridays('2026-11-28', 1)).toEqual(['2026-12-04'])
  })
})
