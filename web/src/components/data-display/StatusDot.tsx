import { cn } from '@/utils/cn'

const tones = {
  success: 'bg-income',
  danger: 'bg-expense',
  warning: 'bg-gold',
  neutral: 'bg-line',
} as const

type StatusDotProps = {
  tone: keyof typeof tones
  /** Announced to screen readers; omit when the status is already written next to the dot. */
  label?: string
  className?: string
}

// 10px dot for compact status (Home TV card). Never the only signal: pair it with text.
export function StatusDot({ tone, label, className }: StatusDotProps) {
  return (
    <span className={cn('inline-flex shrink-0', className)}>
      <span aria-hidden className={cn('size-2.5 rounded-full', tones[tone])} />
      {label && <span className="sr-only">{label}</span>}
    </span>
  )
}
