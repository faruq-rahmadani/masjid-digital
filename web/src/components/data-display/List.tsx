import type { ComponentProps, ReactNode } from 'react'
import { cn } from '@/lib/cn'

export function List({ className, ...props }: ComponentProps<'ul'>) {
  return <ul className={cn('divide-y divide-line-soft', className)} {...props} />
}

type ListRowProps = Omit<ComponentProps<'li'>, 'title'> & {
  title: ReactNode
  /** Second line: date, category, counts. Separate parts with " · ". */
  description?: ReactNode
  leading?: ReactNode
  trailing?: ReactNode
}

export function ListRow({
  title,
  description,
  leading,
  trailing,
  className,
  ...props
}: ListRowProps) {
  return (
    <li className={cn('flex items-center gap-3 py-3', className)} {...props}>
      {leading}
      <div className="min-w-0 flex-1">
        <div className="text-base font-semibold">{title}</div>
        {description && <div className="text-xs text-ink-muted">{description}</div>}
      </div>
      {trailing}
    </li>
  )
}
