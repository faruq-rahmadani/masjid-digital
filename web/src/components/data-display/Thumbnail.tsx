import { ImageIcon } from 'lucide-react'
import { cn } from '@/lib/cn'

type ThumbnailProps = {
  src?: string
  /** Empty for decorative thumbnails whose title is shown next to them. */
  alt?: string
  /** `slide` 72×40 (16:9 list rows), `square` 72×72 (foto bukti, logo). */
  shape?: 'slide' | 'square'
  /** Dims the image for inactive items (nonaktif slide). */
  muted?: boolean
  className?: string
}

export function Thumbnail({ src, alt = '', shape = 'slide', muted, className }: ThumbnailProps) {
  const box = cn(
    'bg-media text-ink-soft shrink-0 overflow-hidden',
    shape === 'slide' ? 'h-10 w-18 rounded-sm' : 'size-18 rounded-md',
    muted && 'opacity-50 grayscale',
    className,
  )

  if (!src) {
    return (
      <span aria-hidden className={cn(box, 'inline-flex items-center justify-center')}>
        <ImageIcon className="size-5" />
      </span>
    )
  }

  return <img src={src} alt={alt} loading="lazy" className={cn(box, 'object-cover')} />
}
