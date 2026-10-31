import { X } from 'lucide-react'
import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'
import { toneIcons, type FeedbackTone } from './tone-icons'

const tones: Record<FeedbackTone, { box: string; icon: string }> = {
  info: { box: 'bg-track text-ink', icon: 'text-ink-soft' }, // mode lihat saja
  success: { box: 'bg-income-soft text-income', icon: 'text-income' },
  warning: { box: 'bg-notice text-notice-ink', icon: 'text-notice-ink' }, // offline, bulan terkunci
  danger: { box: 'bg-expense-soft text-expense-ink', icon: 'text-expense' },
}

type BannerProps = {
  tone?: FeedbackTone
  title?: ReactNode
  children?: ReactNode
  /** One link or button, e.g. "Coba lagi". */
  action?: ReactNode
  /** Shows a close button. Omit for banners that must stay while the condition lasts. */
  onDismiss?: () => void
  /** Announce when the banner appears after load (e.g. going offline). */
  live?: boolean
  className?: string
}

/*
 * Screen-wide status in a full-width strip: place it directly under the PageHeader, outside
 * PageBody. For a message about one card or form, use Callout; for a passing confirmation, a toast.
 */
export function Banner({
  tone = 'info',
  title,
  children,
  action,
  onDismiss,
  live = false,
  className,
}: BannerProps) {
  const Icon = toneIcons[tone]
  const style = tones[tone]

  return (
    <div
      aria-live={live ? (tone === 'danger' ? 'assertive' : 'polite') : undefined}
      className={cn('flex items-start gap-3 px-5 py-3', style.box, className)}
    >
      <Icon aria-hidden className={cn('mt-0.5 size-5 shrink-0', style.icon)} />
      <div className="min-w-0 flex-1 py-0.5 text-sm">
        {title && <p className="font-bold">{title}</p>}
        {children && <div className={cn(title && 'mt-0.5')}>{children}</div>}
        {action && <div className="mt-2 font-bold [&_a,&_button]:underline">{action}</div>}
      </div>
      {onDismiss && (
        <button
          type="button"
          aria-label="Tutup pemberitahuan"
          onClick={onDismiss}
          className="-mr-2 inline-flex size-8 shrink-0 items-center justify-center rounded-md opacity-70 hover:opacity-100"
        >
          <X aria-hidden className="size-4" />
        </button>
      )}
    </div>
  )
}
