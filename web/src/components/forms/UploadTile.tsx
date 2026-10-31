import { Plus, Upload } from 'lucide-react'
import type { ComponentProps, ReactNode } from 'react'
import { cn } from '@/utils/cn'

type UploadTileProps = Omit<ComponentProps<'input'>, 'type' | 'onChange'> & {
  onFiles: (files: File[]) => void
  /** `tile`: 72px square add button. `dropzone`: full-width with title and hint. */
  variant?: 'tile' | 'dropzone'
  label: string
  hint?: ReactNode
}

// A styled <label> around a visually hidden file input, so it works with keyboard and screen readers.
export function UploadTile({
  onFiles,
  variant = 'dropzone',
  label,
  hint,
  className,
  ...props
}: UploadTileProps) {
  return (
    <label
      className={cn(
        'border-dash bg-card text-ink flex cursor-pointer flex-col items-center justify-center border-2 border-dashed',
        'focus-within:outline-gold focus-within:outline-3 focus-within:outline-offset-2',
        variant === 'tile'
          ? 'text-ink-soft size-18 rounded-md'
          : 'gap-1.5 rounded-xl p-4.5 text-center',
        className,
      )}
    >
      <input
        type="file"
        className="sr-only"
        onChange={(event) => {
          onFiles(Array.from(event.currentTarget.files ?? []))
          event.currentTarget.value = ''
        }}
        {...props}
      />
      {variant === 'tile' ? (
        <>
          <Plus aria-hidden className="size-6" />
          <span className="sr-only">{label}</span>
        </>
      ) : (
        <>
          <Upload aria-hidden className="size-6" />
          <span className="text-base font-bold">{label}</span>
          {hint && <span className="text-ink-muted text-xs">{hint}</span>}
        </>
      )}
    </label>
  )
}
