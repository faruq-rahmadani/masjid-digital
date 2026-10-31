import { ChevronDown } from 'lucide-react'
import type { ComponentProps } from 'react'
import { type ControlSize, controlBase, controlSizes } from '@/components/forms/control-styles'
import { useFieldControl } from '@/components/forms/field-context'
import { cn } from '@/utils/cn'

type SelectProps = Omit<ComponentProps<'select'>, 'size'> & { size?: ControlSize | 'pill' }

// Native <select> so phones get their own picker; only the closed state is styled.
export function Select({ size = 'md', className, children, ...props }: SelectProps) {
  return (
    <div className={cn('relative', size === 'pill' && 'inline-flex', className)}>
      <select
        className={cn(
          controlBase,
          'appearance-none pr-9',
          size === 'pill' ? 'h-9 rounded-full pl-3 text-sm' : controlSizes[size],
        )}
        {...useFieldControl(props)}
      >
        {children}
      </select>
      <ChevronDown
        aria-hidden
        className="text-ink-muted pointer-events-none absolute top-1/2 right-3 size-4 -translate-y-1/2"
      />
    </div>
  )
}
