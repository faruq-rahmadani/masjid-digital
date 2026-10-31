import type { ComponentProps, ReactNode } from 'react'
import { cn } from '@/lib/cn'

const variants = {
  card: 'size-touch rounded-full bg-card text-ink', // header back / settings
  brand: 'size-touch rounded-full bg-brand-2 text-on-brand', // on the brand header
  ghost: 'h-touch w-9 rounded-md bg-transparent text-ink-muted', // ⋯ row menu
} as const

type IconButtonProps = Omit<ComponentProps<'button'>, 'aria-label' | 'children'> & {
  /** Accessible name; icon-only buttons have no visible text. */
  label: string
  icon: ReactNode
  variant?: keyof typeof variants
}

export function IconButton({
  label,
  icon,
  variant = 'card',
  type = 'button',
  className,
  ...props
}: IconButtonProps) {
  return (
    <button
      type={type}
      aria-label={label}
      className={cn(
        'inline-flex shrink-0 items-center justify-center transition-opacity active:opacity-70',
        'disabled:pointer-events-none disabled:opacity-50 [&_svg]:size-5',
        variants[variant],
        className,
      )}
      {...props}
    >
      {icon}
    </button>
  )
}
