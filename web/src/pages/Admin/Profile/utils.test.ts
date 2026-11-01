import { describe, expect, it } from 'vitest'
import { formatCoordinates, guessTimezone, logoProblem, roundCoordinates } from './utils'

describe('guessTimezone', () => {
  it('maps the main cities to their zone', () => {
    expect(guessTimezone(106.8456)).toBe('Asia/Jakarta') // Jakarta
    expect(guessTimezone(98.6722)).toBe('Asia/Jakarta') // Medan
    expect(guessTimezone(115.2126)).toBe('Asia/Makassar') // Denpasar
    expect(guessTimezone(119.4327)).toBe('Asia/Makassar') // Makassar
    expect(guessTimezone(128.1814)).toBe('Asia/Jayapura') // Ambon
    expect(guessTimezone(140.7181)).toBe('Asia/Jayapura') // Jayapura
  })
})

describe('coordinates', () => {
  it('rounds to 6 decimals and shows 4', () => {
    const point = roundCoordinates({ latitude: -6.20876543, longitude: 106.84561234 })
    expect(point).toEqual({ latitude: -6.208765, longitude: 106.845612 })
    expect(formatCoordinates(point)).toBe('-6.2088, 106.8456')
  })
})

describe('logoProblem', () => {
  const file = (type: string, size = 10) => new File([new Uint8Array(size)], 'logo', { type })

  it('accepts PNG, JPG and WebP up to 2 MB', () => {
    expect(logoProblem(file('image/png'))).toBeNull()
    expect(logoProblem(file('image/jpeg'))).toBeNull()
    expect(logoProblem(file('image/webp', 2 * 1024 * 1024))).toBeNull()
  })

  it('rejects SVG and files over 2 MB', () => {
    expect(logoProblem(file('image/svg+xml'))).toMatch(/PNG, JPG, atau WebP/)
    expect(logoProblem(file('image/png', 2 * 1024 * 1024 + 1))).toBe('Ukuran logo maksimal 2 MB')
  })
})
