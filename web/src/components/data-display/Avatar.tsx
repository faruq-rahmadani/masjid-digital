import { cn } from '@/utils/cn'
import { initials } from '@/utils/format/text'

const tones = {
  brand: 'bg-brand text-on-brand', // super admin on white
  'brand-2': 'bg-brand-2 text-on-brand', // on the green hero
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
        size === 'md' ? 'text-md size-10' : 'size-touch text-base',
        tones[tone],
        className,
      )}
    >
      {initials(name)}
    </span>
  )
}
