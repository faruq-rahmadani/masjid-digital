import type { ComponentProps } from 'react'
import { type ControlSize, controlBase, controlSizes } from '@/components/forms/control-styles'
import { useFieldControl } from '@/components/forms/field-context'
import { cn } from '@/utils/cn'

type InputProps = Omit<ComponentProps<'input'>, 'size'> & { size?: ControlSize }

export function Input({ size = 'md', className, ...props }: InputProps) {
  return (
    <input className={cn(controlBase, controlSizes[size], className)} {...useFieldControl(props)} />
  )
}
