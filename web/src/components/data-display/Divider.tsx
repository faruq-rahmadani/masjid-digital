import { cn } from '@/utils/cn'

export function Divider({ label, className }: { label?: string; className?: string }) {
  if (!label) return <hr className={cn('border-line', className)} />

  return (
    <div className={cn('text-ink-muted flex items-center gap-3 text-xs', className)}>
      <span className="bg-line h-px flex-1" />
      {label}
      <span className="bg-line h-px flex-1" />
    </div>
  )
}
