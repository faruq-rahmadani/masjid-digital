import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'

// The green header that opens Home; extends under the status bar on notched phones.
export function HeroHeader({ className, children }: { className?: string; children: ReactNode }) {
  return (
    <header
      className={cn('bg-brand text-on-brand pt-safe-6 flex flex-col gap-4 px-5 pb-6', className)}
    >
      {children}
    </header>
  )
}
