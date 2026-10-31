import type { ComponentProps } from 'react'
import { cn } from '@/lib/cn'
import { controlBase, controlSizes, type ControlSize } from './control-styles'
import { useFieldControl } from './field-context'

type InputProps = Omit<ComponentProps<'input'>, 'size'> & { size?: ControlSize }

export function Input({ size = 'md', className, ...props }: InputProps) {
  return (
    <input className={cn(controlBase, controlSizes[size], className)} {...useFieldControl(props)} />
  )
}
