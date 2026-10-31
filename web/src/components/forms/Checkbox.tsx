import type { ComponentProps, ReactNode } from 'react'
import { cn } from '@/lib/cn'

type CheckboxProps = Omit<ComponentProps<'input'>, 'type' | 'size'> & {
  label: ReactNode
  size?: 'md' | 'lg'
}

// The mockup's on/off settings ("Aktif", "Tandai penting") are labelled checkboxes, not switches.
export function Checkbox({ label, size = 'md', className, ...props }: CheckboxProps) {
  return (
    <label
      className={cn(
        'inline-flex min-h-touch cursor-pointer items-center gap-2 text-sm font-semibold text-ink',
        className,
      )}
    >
      <input
        type="checkbox"
        className={cn('shrink-0 accent-brand', size === 'md' ? 'size-5' : 'size-5.5')}
        {...props}
      />
      {label}
    </label>
  )
}
