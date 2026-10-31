import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'

const tones = {
  notice: 'bg-notice text-notice-ink', // saldo preview, tutup buku
  danger: 'bg-expense-soft text-expense-ink', // screen offline
} as const

type CalloutProps = {
  tone?: keyof typeof tones
  title?: ReactNode
  /** Button placed to the right of the text. */
  action?: ReactNode
  className?: string
  children?: ReactNode
}

export function Callout({ tone = 'notice', title, action, className, children }: CalloutProps) {
  return (
    <div className={cn('flex items-center gap-3 rounded-xl px-4 py-3.5', tones[tone], className)}>
      <div className="min-w-0 flex-1 text-xs leading-normal">
        {title && <p className="text-base font-bold">{title}</p>}
        {children && <div className={cn(title && 'mt-0.5')}>{children}</div>}
      </div>
      {action}
    </div>
  )
}
