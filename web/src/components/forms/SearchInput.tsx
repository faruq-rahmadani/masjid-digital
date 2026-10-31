import { Search } from 'lucide-react'
import type { ComponentProps } from 'react'
import { cn } from '@/lib/cn'

type SearchInputProps = Omit<ComponentProps<'input'>, 'type' | 'aria-label'> & {
  /** Accessible name; the placeholder alone is not a label. */
  label: string
}

export function SearchInput({ label, className, ...props }: SearchInputProps) {
  return (
    <label
      className={cn(
        'flex h-11 items-center gap-2 rounded-lg border border-line bg-card px-3 text-ink-muted',
        'focus-within:outline-3 focus-within:outline-offset-2 focus-within:outline-gold',
        className,
      )}
    >
      <Search aria-hidden className="size-4.5 shrink-0" />
      <input
        type="search"
        aria-label={label}
        className="min-w-0 flex-1 bg-transparent text-base text-ink outline-none placeholder:text-ink-muted"
        {...props}
      />
    </label>
  )
}
