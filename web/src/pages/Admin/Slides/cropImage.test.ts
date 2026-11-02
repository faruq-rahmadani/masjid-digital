import { describe, expect, it } from 'vitest'
import { MAX_OUTPUT_WIDTH, outputSize } from './cropImage'

describe('outputSize', () => {
  it('scales a large crop (a phone photo) down to twice the TV panel', () => {
    expect(MAX_OUTPUT_WIDTH).toBe(2624)
    expect(outputSize(4032)).toEqual({ width: 2624, height: 1164 })
  })

  it("keeps a small crop as it is, in the panel's shape", () => {
    expect(outputSize(1312)).toEqual({ width: 1312, height: 582 })
    expect(outputSize(800.4)).toEqual({ width: 800, height: 355 })
  })
})
