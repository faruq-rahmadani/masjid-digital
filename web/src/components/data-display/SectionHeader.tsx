import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'

type SectionHeaderProps = {
  title: ReactNode
  description?: ReactNode
  /** Right-aligned link or count, e.g. "Lihat semua". */
  action?: ReactNode
  as?: 'h2' | 'h3'
  className?: string
}

export function SectionHeader({
  title,
  description,
  action,
  as: Heading = 'h2',
  className,
}: SectionHeaderProps) {
  return (
    <div className={cn('flex items-center justify-between gap-3', className)}>
      <div className="min-w-0">
        <Heading className="text-md font-bold">{title}</Heading>
        {description && <p className="text-xs text-ink-muted">{description}</p>}
      </div>
      {action && <div className="shrink-0 text-xs font-semibold">{action}</div>}
    </div>
  )
}
