import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'

const tones = {
  brand: 'bg-brand text-on-brand', // Super admin
  success: 'bg-income-soft text-income', // Online, Takmir
  danger: 'bg-expense-soft text-expense', // Offline
  gold: 'bg-gold-soft text-gold-ink', // BARU, PENTING, Bendahara
  notice: 'bg-notice text-notice-ink', // Langsung tersimpan: acts right away
  neutral: 'bg-line-soft text-ink-muted', // Nonaktif, Super admin
} as const

const sizes = {
  sm: 'rounded-sm px-2 py-1 text-2xs', // role and flag tags
  md: 'rounded-md px-2.5 py-1 text-xs', // status pills
} as const

type BadgeProps = {
  tone?: keyof typeof tones
  size?: keyof typeof sizes
  /** Uppercase flag style used for BARU and PENTING. */
  caps?: boolean
  className?: string
  children: ReactNode
}

export function Badge({
  tone = 'neutral',
  size = 'sm',
  caps = false,
  className,
  children,
}: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center font-bold whitespace-nowrap',
        tones[tone],
        sizes[size],
        caps && 'font-extrabold tracking-wider uppercase',
        className,
      )}
    >
      {children}
    </span>
  )
}
