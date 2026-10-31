import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'

// The green header that opens Beranda; extends under the status bar on notched phones.
export function HeroHeader({ className, children }: { className?: string; children: ReactNode }) {
  return (
    <header
      className={cn(
        'flex flex-col gap-4 bg-brand px-5 pt-[max(1.5rem,env(safe-area-inset-top))] pb-6 text-on-brand',
        className,
      )}
    >
      {children}
    </header>
  )
}
