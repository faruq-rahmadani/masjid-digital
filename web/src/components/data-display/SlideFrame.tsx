import { Image } from 'lucide-react'
import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'

/**
 * The TV's slide panel in normal mode, in stage px: 1920 wide minus the padding (2×32), the clock
 * column (520) and the gap (24); 1080 high minus the padding, header (120), schedule (190), INFO
 * bar (52) and three gaps (24). New slides are cropped to this, so they fit it exactly. Without
 * running text the INFO bar hides and the panel grows to 658 high: the image then loses a little
 * at the sides (object-cover).
 */
export const SLIDE_PANEL = { width: 1312, height: 582 } as const
export const SLIDE_ASPECT = SLIDE_PANEL.width / SLIDE_PANEL.height

type SlideSize = 'tv' | 'preview'

type SlideFrameProps = {
  /** The image; empty shows a placeholder (no slides yet, or a preview still loading). */
  src?: string
  title: string
  caption: string
  /** Off: no caption bar, the image gets the whole frame. */
  showCaption: boolean
  /** `tv` on the 1920×1080 stage; `preview` the same panel scaled down for the admin sheets. */
  size?: SlideSize
  /** Next to the caption, e.g. the TV's slide dots. */
  aside?: ReactNode
  className?: string
}

// The preview is the TV panel scaled to its width: sizes in cqw are the TV's px / 1312.
const sizes = {
  tv: {
    frame: 'rounded-tv-lg',
    bar: 'h-35 gap-6 px-10',
    title: 'text-tv-34',
    caption: 'text-tv-22 mt-1.5',
    icon: 'size-24',
  },
  preview: {
    frame: '@container aspect-[1312/582] rounded-md',
    bar: 'h-[24%] gap-[1.8cqw] px-[3cqw]', // 140 / 582 high
    title: 'text-[max(2.6cqw,11px)]',
    caption: 'text-[max(1.7cqw,9px)] mt-[0.4cqw]',
    icon: 'size-10',
  },
} as const

/*
 * A slide as the TV shows it: the image fills the frame (object-cover) and the title and caption
 * sit on a frosted-glass bar over its bottom. The bar may cover text printed on a poster; the
 * admin sees that in the preview and can turn the caption off. Used by the TV and the admin
 * preview, so what the admin sees is what the TV shows.
 */
export function SlideFrame({
  src,
  title,
  caption,
  showCaption,
  size = 'tv',
  aside,
  className,
}: SlideFrameProps) {
  const s = sizes[size]

  return (
    <div className={cn('bg-tv-media text-on-brand relative overflow-hidden', s.frame, className)}>
      {src ? (
        <img src={src} alt="" className="absolute inset-0 size-full object-cover" />
      ) : (
        <div className="text-on-brand-muted absolute inset-0 flex items-center justify-center">
          <Image aria-hidden className={s.icon} strokeWidth={1.5} />
        </div>
      )}
      {showCaption && <SlideCaptionBar title={title} caption={caption} size={size} aside={aside} />}
    </div>
  )
}

type SlideCaptionBarProps = {
  title: string
  caption: string
  size?: SlideSize
  aside?: ReactNode
  className?: string
}

/*
 * The frosted-glass caption bar, on its own so the crop step can lay it over the cropper. The
 * photo behind it is blurred and dimmed so white text stays readable on a bright one; TVs without
 * backdrop-filter get a denser tint instead. Nothing when there is no text.
 */
export function SlideCaptionBar({
  title,
  caption,
  size = 'tv',
  aside,
  className,
}: SlideCaptionBarProps) {
  const s = sizes[size]
  if (title === '' && caption === '') return null

  return (
    <div
      className={cn(
        'text-on-brand absolute inset-x-0 bottom-0 flex items-center justify-between',
        'bg-tv-caption/80 border-t border-white/15 [text-shadow:0_1px_2px_rgb(0_0_0/0.5),0_2px_12px_rgb(0_0_0/0.45)]',
        'supports-[backdrop-filter:blur(0)]:bg-tv-caption/40 supports-[backdrop-filter:blur(0)]:backdrop-blur-xl supports-[backdrop-filter:blur(0)]:backdrop-brightness-75 supports-[backdrop-filter:blur(0)]:backdrop-saturate-150',
        s.bar,
        className,
      )}
    >
      <div className="min-w-0">
        <div className={cn('truncate font-bold', s.title)}>{title}</div>
        {caption && <div className={cn('text-on-brand-muted truncate', s.caption)}>{caption}</div>}
      </div>
      {aside}
    </div>
  )
}
