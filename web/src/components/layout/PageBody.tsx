import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'

// Standard content column under a PageHeader: 20px gutters, 14px between cards.
export function PageBody({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={cn('flex flex-1 flex-col gap-3.5 px-5 pb-6', className)}>{children}</div>
}
