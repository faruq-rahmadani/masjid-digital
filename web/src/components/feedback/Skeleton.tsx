import { cn } from '@/utils/cn'

export function Skeleton({ className }: { className?: string }) {
  return <div aria-hidden className={cn('bg-line-soft animate-pulse rounded-md', className)} />
}
