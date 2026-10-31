import { LoaderCircle } from 'lucide-react'
import { cn } from '@/lib/cn'

// Inline loading indicator. Prefer <Skeleton> for content that has a known shape.
export function Spinner({ label = 'Memuat', className }: { label?: string; className?: string }) {
  return (
    <output className={cn('inline-flex items-center gap-2 text-ink-muted', className)}>
      <LoaderCircle aria-hidden className="size-5 animate-spin" />
      <span className="sr-only">{label}</span>
    </output>
  )
}
