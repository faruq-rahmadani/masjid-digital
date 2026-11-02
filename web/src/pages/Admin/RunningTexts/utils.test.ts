import { describe, expect, it } from 'vitest'
import { mergeOrder } from './utils'

describe('mergeOrder', () => {
  it('reorders only the shown texts, keeping the hidden ones in place', () => {
    // 2 and 4 are hidden by the filter; 5 moved above 1
    expect(mergeOrder([1, 2, 3, 4, 5], [5, 1, 3])).toEqual([5, 2, 1, 4, 3])
  })

  it('is a plain reorder when nothing is hidden', () => {
    expect(mergeOrder([1, 2, 3], [3, 1, 2])).toEqual([3, 1, 2])
  })
})
