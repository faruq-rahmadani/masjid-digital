import { cn } from '@/utils/cn'

// Variants map 1:1 to buttons in docs/design/mockups; see design-system.md#button.
const variants = {
  primary: 'bg-brand text-on-brand font-bold',
  secondary: 'border border-line bg-card text-ink font-semibold',
  outline: 'border-2 border-brand bg-transparent text-brand font-bold',
  ghost: 'bg-transparent text-ink font-semibold',
  danger: 'border border-expense-line bg-card text-expense font-semibold',
  income: 'bg-income text-white font-bold',
  expense: 'border-2 border-expense bg-card text-expense font-bold',
  gold: 'bg-gold text-on-gold font-bold',
  notice: 'bg-notice-ink text-white font-bold',
} as const

const sizes = {
  sm: 'h-9 rounded-md px-3 text-sm', // 36px: chips, inline row actions
  md: 'h-10 rounded-md px-3.5 text-sm', // 40px: card actions
  lg: 'h-12 rounded-lg px-4 text-md', // 48px: section actions
  xl: 'h-13 rounded-xl px-5 text-lg', // 52px: the screen's primary action
} as const

export type ButtonVariant = keyof typeof variants
export type ButtonSize = keyof typeof sizes

export type ButtonStyleProps = {
  variant?: ButtonVariant
  size?: ButtonSize
  block?: boolean
}

// Exposed separately so a router <Link> can look like a button without wrapping it in one.
export function buttonStyles({
  variant = 'primary',
  size = 'lg',
  block = false,
}: ButtonStyleProps = {}) {
  return cn(
    'inline-flex shrink-0 items-center justify-center gap-2 whitespace-nowrap select-none',
    'transition-opacity active:opacity-80 disabled:pointer-events-none disabled:opacity-50',
    '[&_svg]:size-[1.15em] [&_svg]:shrink-0',
    variants[variant],
    sizes[size],
    block && 'w-full',
  )
}
