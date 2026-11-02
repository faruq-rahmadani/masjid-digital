import type { SelectOption } from '@/components'

// Same formats as internal/domain/slide.go. The picked file may be larger than the API's 5 MB:
// it is cropped and scaled down here before upload (cropImage), which brings phone photos well
// under. The limit only keeps a huge file from stalling the phone.
export const MAX_SLIDE_SIZE = 20 * 1024 * 1024
export const SLIDE_TYPES = ['image/png', 'image/jpeg', 'image/webp'] as const
export const MAX_TITLE_LENGTH = 100
export const MAX_CAPTION_LENGTH = 200

/** The reason an image would be rejected, or null when it can be uploaded. */
export function slideProblem(file: File): string | null {
  if (!(SLIDE_TYPES as readonly string[]).includes(file.type)) {
    return 'Format gambar harus PNG, JPG, atau WebP'
  }
  if (file.size > MAX_SLIDE_SIZE) return 'Ukuran gambar maksimal 20 MB'
  return null
}

/** The default title for a new slide: the file name without its extension, as the API does. */
export const titleFromFileName = (name: string) =>
  [...name.replace(/\.[^.]*$/, '').trim()].slice(0, MAX_TITLE_LENGTH).join('')

export const DURATION_OPTIONS: readonly SelectOption<string>[] = [5, 10, 15, 20, 30].map((s) => ({
  value: String(s),
  label: `${s} detik`,
}))

/** The TV polls every 5 minutes (usePublicSlides); push updates come with Layar TV (E9). */
export const TV_DELAY = 'Tampil di TV dalam ±5 menit'
