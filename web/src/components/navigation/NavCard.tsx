import { ChevronRight } from 'lucide-react'
import type { ReactNode } from 'react'
import { Link } from 'react-router'
import { cn } from '@/lib/cn'

type NavCardProps = {
  to: string
  title: ReactNode
  description?: ReactNode
  /** Status shown before the chevron, e.g. StatusDots. */
  trailing?: ReactNode
  className?: string
}

// A whole card that links somewhere (Beranda: "Layar TV · 1 online · 1 offline").
export function NavCard({ to, title, description, trailing, className }: NavCardProps) {
  return (
    <Link
      to={to}
      className={cn(
        'flex items-center gap-3 rounded-xl bg-card px-4 py-3 text-ink active:opacity-80',
        className,
      )}
    >
      <span className="min-w-0 flex-1">
        <span className="block font-bold">{title}</span>
        {description && <span className="block text-xs text-ink-muted">{description}</span>}
      </span>
      {trailing}
      <ChevronRight aria-hidden className="size-4.5 shrink-0" />
    </Link>
  )
}
