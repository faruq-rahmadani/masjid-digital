import type { ComponentProps } from 'react'
import { cn } from '@/lib/cn'

const tones = {
  default: 'rounded-xl bg-card text-ink',
  brand: 'rounded-3xl bg-brand text-on-brand', // saldo card
  'brand-2': 'rounded-2xl bg-brand-2 text-on-brand', // cards sitting on the brand header
} as const

const paddings = {
  none: '',
  list: 'px-4 py-0.5', // wraps a <List>; rows bring their own vertical padding
  md: 'px-4 py-3.5',
  lg: 'p-4',
  xl: 'px-5 py-4.5',
} as const

type CardProps = ComponentProps<'div'> & {
  tone?: keyof typeof tones
  padding?: keyof typeof paddings
}

export function Card({ tone = 'default', padding = 'lg', className, ...props }: CardProps) {
  return <div className={cn(tones[tone], paddings[padding], className)} {...props} />
}
