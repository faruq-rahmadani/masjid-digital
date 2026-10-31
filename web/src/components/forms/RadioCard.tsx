import type { ComponentProps, ReactNode } from 'react'
import { cn } from '@/lib/cn'

type RadioCardProps = Omit<ComponentProps<'input'>, 'type'> & {
  title: ReactNode
  description?: ReactNode
}

// A radio choice with a description; the selected card gets a 2px brand border.
export function RadioCard({ title, description, className, ...props }: RadioCardProps) {
  return (
    <label
      className={cn(
        'flex cursor-pointer items-center gap-2.5 rounded-lg border border-line bg-card p-3 text-base',
        'has-checked:border-brand has-checked:outline-1 has-checked:-outline-offset-2 has-checked:outline-brand',
        'has-focus-visible:outline-3 has-focus-visible:outline-gold',
        className,
      )}
    >
      <input type="radio" className="size-4.5 shrink-0 accent-brand" {...props} />
      <span className="min-w-0">
        <span className="block font-semibold">{title}</span>
        {description && <span className="block text-xs text-ink-muted">{description}</span>}
      </span>
    </label>
  )
}
