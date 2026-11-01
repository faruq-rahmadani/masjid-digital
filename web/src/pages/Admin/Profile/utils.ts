import type { MosqueTimezone } from '@/types/mosque'

export type Coordinates = { latitude: number; longitude: number }

/*
 * Indonesia's zones follow province borders, not meridians, so this is only a first guess from the
 * longitude (Kalimantan Tengah vs Selatan, the Sula islands); the admin can still change it.
 */
export function guessTimezone(longitude: number): MosqueTimezone {
  if (longitude < 114.5) return 'Asia/Jakarta'
  if (longitude < 126) return 'Asia/Makassar'
  return 'Asia/Jayapura'
}

/** 6 decimals is about 10 cm, more than a pin on a phone can place. */
export function roundCoordinates({ latitude, longitude }: Coordinates): Coordinates {
  const round = (value: number) => Math.round(value * 1e6) / 1e6
  return { latitude: round(latitude), longitude: round(longitude) }
}

/** "-6.2088, 106.8456" */
export function formatCoordinates({ latitude, longitude }: Coordinates) {
  return `${latitude.toFixed(4)}, ${longitude.toFixed(4)}`
}

// Same rules as internal/domain/mosque_profile.go; the API also checks the bytes, not just the type.
export const MAX_LOGO_SIZE = 2 * 1024 * 1024
export const LOGO_TYPES = ['image/png', 'image/jpeg', 'image/webp'] as const

/** The reason a logo would be rejected, or null when it can be uploaded. */
export function logoProblem(file: File): string | null {
  if (!(LOGO_TYPES as readonly string[]).includes(file.type)) {
    return 'Format logo harus PNG, JPG, atau WebP'
  }
  if (file.size > MAX_LOGO_SIZE) return 'Ukuran logo maksimal 2 MB'
  return null
}
