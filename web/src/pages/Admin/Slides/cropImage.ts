import type { Area } from 'react-easy-crop'
import { SLIDE_ASPECT, SLIDE_PANEL } from '@/components'

/** Twice the TV panel, so a slide stays sharp on a 4K TV; smaller crops are kept as they are. */
export const MAX_OUTPUT_WIDTH = SLIDE_PANEL.width * 2

/** The uploaded size for a crop `areaWidth` px wide: never wider than MAX_OUTPUT_WIDTH. */
export function outputSize(areaWidth: number) {
  const width = Math.round(Math.min(areaWidth, MAX_OUTPUT_WIDTH))
  return { width, height: Math.round(width / SLIDE_ASPECT) }
}

/** Under transparent parts of a PNG: the caption bar's green, as on the TV. */
const BACKGROUND = '#0a2a22'

/**
 * Cut `area` (in the image's pixels, from the cropper) out of `file`, scaled down to at most
 * MAX_OUTPUT_WIDTH, as a JPEG ready to upload. The image is loaded the same way the cropper
 * shows it (an <img>, which applies the photo's EXIF rotation), so the cut matches what the admin
 * framed.
 */
export async function cropImage(file: File, area: Area): Promise<File> {
  const url = URL.createObjectURL(file)
  try {
    const image = new Image()
    image.src = url
    await image.decode()

    const { width, height } = outputSize(area.width)
    const canvas = document.createElement('canvas')
    canvas.width = width
    canvas.height = height
    const context = canvas.getContext('2d')
    if (!context) throw new Error('Browser tidak bisa memotong gambar')
    context.fillStyle = BACKGROUND
    context.fillRect(0, 0, width, height)
    context.imageSmoothingQuality = 'high'
    context.drawImage(image, area.x, area.y, area.width, area.height, 0, 0, width, height)

    const blob = await new Promise<Blob | null>((resolve) =>
      canvas.toBlob(resolve, 'image/jpeg', 0.9),
    )
    if (!blob) throw new Error('Gagal memotong gambar')
    return new File([blob], `${file.name.replace(/\.[^.]*$/, '')}.jpg`, { type: 'image/jpeg' })
  } finally {
    URL.revokeObjectURL(url)
  }
}
