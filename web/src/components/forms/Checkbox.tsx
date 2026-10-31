import type { ComponentProps, ReactNode } from 'react'
import { cn } from '@/utils/cn'

type CheckboxProps = Omit<ComponentProps<'input'>, 'type' | 'size'> & {
  label: ReactNode
  size?: 'md' | 'lg'
}

// The mockup's on/off settings ("Aktif", "Tandai penting") are labelled checkboxes, not switches.
export function Checkbox({ label, size = 'md', className, ...props }: CheckboxProps) {
  return (
    <label
      className={cn(
        'min-h-touch text-ink inline-flex cursor-pointer items-center gap-2 text-sm font-semibold',
        className,
      )}
    >
      <input
        type="checkbox"
        className={cn('accent-brand shrink-0', size === 'md' ? 'size-5' : 'size-5.5')}
        {...props}
      />
      {label}
    </label>
  )
}
