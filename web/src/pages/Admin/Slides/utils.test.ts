import { describe, expect, it } from 'vitest'
import { slideProblem } from './utils'

const file = (type: string, size = 1024) => new File([new Uint8Array(size)], 'poster', { type })

describe('slideProblem', () => {
  it('accepts PNG, JPG and WebP up to 20 MB', () => {
    expect(slideProblem(file('image/png'))).toBeNull()
    expect(slideProblem(file('image/jpeg'))).toBeNull()
    expect(slideProblem(file('image/webp', 20 * 1024 * 1024))).toBeNull()
  })

  it('rejects other formats and files over 20 MB', () => {
    expect(slideProblem(file('image/gif'))).toBe('Format gambar harus PNG, JPG, atau WebP')
    expect(slideProblem(file('image/png', 20 * 1024 * 1024 + 1))).toBe(
      'Ukuran gambar maksimal 20 MB',
    )
  })
})
