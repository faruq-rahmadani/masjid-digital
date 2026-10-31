import { cn } from '@/lib/cn'

export function Divider({ label, className }: { label?: string; className?: string }) {
  if (!label) return <hr className={cn('border-line', className)} />

  return (
    <div className={cn('flex items-center gap-3 text-xs text-ink-muted', className)}>
      <span className="h-px flex-1 bg-line" />
      {label}
      <span className="h-px flex-1 bg-line" />
    </div>
  )
}
