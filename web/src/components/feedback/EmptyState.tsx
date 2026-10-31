import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'

type EmptyStateProps = {
  title: ReactNode
  description?: ReactNode
  action?: ReactNode
  className?: string
}

export function EmptyState({ title, description, action, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        'border-line flex flex-col items-center gap-1 rounded-xl border border-dashed px-6 py-8 text-center',
        className,
      )}
    >
      <p className="text-base font-bold">{title}</p>
      {description && <p className="text-ink-muted text-xs text-pretty">{description}</p>}
      {action && <div className="mt-3">{action}</div>}
    </div>
  )
}
