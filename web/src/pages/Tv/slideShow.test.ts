import { describe, expect, it } from 'vitest'
import { slideIndexAt } from './slideShow'

describe('slideIndexAt', () => {
  it('keeps each slide on for its own duration, then loops', () => {
    const durations = [10, 5, 15] // 30 seconds in all
    expect(slideIndexAt(durations, 0)).toBe(0)
    expect(slideIndexAt(durations, 9.9)).toBe(0)
    expect(slideIndexAt(durations, 10)).toBe(1)
    expect(slideIndexAt(durations, 14)).toBe(1)
    expect(slideIndexAt(durations, 15)).toBe(2)
    expect(slideIndexAt(durations, 29)).toBe(2)
    expect(slideIndexAt(durations, 30)).toBe(0)
    expect(slideIndexAt(durations, 1_793_596_812)).toBe(slideIndexAt(durations, 1_793_596_812 % 30))
  })

  it('is -1 without slides', () => {
    expect(slideIndexAt([], 42)).toBe(-1)
  })
})
