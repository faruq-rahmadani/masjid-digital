import type { ComponentProps, ReactNode } from 'react'
import { cn } from '@/utils/cn'

type RadioCardProps = Omit<ComponentProps<'input'>, 'type'> & {
  title: ReactNode
  description?: ReactNode
}

// A radio choice with a description; the selected card gets a 2px brand border.
export function RadioCard({ title, description, className, ...props }: RadioCardProps) {
  return (
    <label
      className={cn(
        'border-line bg-card flex cursor-pointer items-center gap-2.5 rounded-lg border p-3 text-base',
        'has-checked:border-brand has-checked:outline-brand has-checked:outline-1 has-checked:-outline-offset-2',
        'has-focus-visible:outline-gold has-focus-visible:outline-3',
        className,
      )}
    >
      <input type="radio" className="accent-brand size-4.5 shrink-0" {...props} />
      <span className="min-w-0">
        <span className="block font-semibold">{title}</span>
        {description && <span className="text-ink-muted block text-xs">{description}</span>}
      </span>
    </label>
  )
}
