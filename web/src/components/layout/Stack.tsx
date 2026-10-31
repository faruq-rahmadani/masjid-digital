import type { ComponentProps } from 'react'
import { cn } from '@/utils/cn'

const gaps = {
  xs: 'gap-1.5', // label → control
  sm: 'gap-2.5', // fields inside a card, buttons in a row
  md: 'gap-3.5', // cards on a page (default)
  lg: 'gap-5', // page sections
} as const

type StackProps = ComponentProps<'div'> & { gap?: keyof typeof gaps }

// Vertical rhythm from the spacing scale, so screens don't invent their own gaps.
export function Stack({ gap = 'md', className, ...props }: StackProps) {
  return <div className={cn('flex flex-col', gaps[gap], className)} {...props} />
}
