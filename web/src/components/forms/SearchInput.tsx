import { Search, X } from 'lucide-react'
import type { ComponentProps } from 'react'
import { cn } from '@/utils/cn'

type SearchInputProps = Omit<ComponentProps<'input'>, 'type' | 'aria-label'> & {
  /** Accessible name; the placeholder alone is not a label. */
  label: string
  /** Shows a clear (×) button while there is text. */
  onClear?: () => void
}

export function SearchInput({ label, onClear, className, ...props }: SearchInputProps) {
  const hasText = String(props.value ?? '') !== ''

  return (
    <div
      className={cn(
        'border-line bg-card text-ink-muted flex h-11 items-center gap-2 rounded-lg border pl-3',
        onClear ? 'pr-1' : 'pr-3',
        'focus-within:outline-gold focus-within:outline-3 focus-within:outline-offset-2',
        className,
      )}
    >
      <Search aria-hidden className="size-4.5 shrink-0" />
      <input
        type="search"
        aria-label={label}
        // The browser's own clear button is off-palette; ours below replaces it.
        className="text-ink placeholder:text-ink-muted min-w-0 flex-1 bg-transparent text-base outline-none [&::-webkit-search-cancel-button]:appearance-none"
        {...props}
      />
      {onClear && hasText && (
        <button
          type="button"
          aria-label="Hapus pencarian"
          onClick={onClear}
          className="inline-flex size-9 shrink-0 items-center justify-center rounded-md"
        >
          <X aria-hidden className="size-4" />
        </button>
      )}
    </div>
  )
}
