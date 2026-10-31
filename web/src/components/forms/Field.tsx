import { useId, type ReactNode } from 'react'
import { cn } from '@/lib/cn'
import { FieldContext } from './field-context'

type FieldProps = {
  label: ReactNode
  /** Help text under the control, e.g. limits from the PRD. */
  hint?: ReactNode
  /** Validation message; also marks the control aria-invalid. */
  error?: ReactNode
  className?: string
  children: ReactNode
}

export function Field({ label, hint, error, className, children }: FieldProps) {
  const id = useId()
  const hintId = hint ? `${id}-hint` : undefined
  const errorId = error ? `${id}-error` : undefined
  const describedBy = [errorId, hintId].filter(Boolean).join(' ') || undefined

  return (
    <FieldContext value={{ id, describedBy, invalid: Boolean(error) }}>
      <div className={cn('flex flex-col gap-1.5', className)}>
        <label htmlFor={id} className="text-xs font-semibold text-ink-soft">
          {label}
        </label>
        {children}
        {error && (
          <p id={errorId} className="text-xs font-semibold text-expense">
            {error}
          </p>
        )}
        {hint && (
          <p id={hintId} className="text-xs text-ink-muted">
            {hint}
          </p>
        )}
      </div>
    </FieldContext>
  )
}
