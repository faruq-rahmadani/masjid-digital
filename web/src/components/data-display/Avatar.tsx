import { initials } from '@/lib/format'
import { cn } from '@/lib/cn'

const tones = {
  brand: 'bg-brand-2 text-on-brand',
  success: 'bg-income-soft text-income',
  gold: 'bg-gold-soft text-gold-ink',
  neutral: 'bg-line-soft text-ink-muted',
} as const

type AvatarProps = {
  name: string
  tone?: keyof typeof tones
  size?: 'md' | 'lg'
  className?: string
}

// Decorative: the name is always shown next to it, so screen readers skip the initials.
export function Avatar({ name, tone = 'brand', size = 'md', className }: AvatarProps) {
  return (
    <span
      aria-hidden
      className={cn(
        'inline-flex shrink-0 items-center justify-center rounded-full font-bold',
        size === 'md' ? 'size-10 text-md' : 'size-touch text-base',
        tones[tone],
        className,
      )}
    >
      {initials(name)}
    </span>
  )
}
