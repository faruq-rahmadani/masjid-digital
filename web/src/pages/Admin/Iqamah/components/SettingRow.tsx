import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'

type SettingRowProps = {
  title: string
  description?: string
  /** Changed since the last save: highlighted like a changed table row. */
  changed?: boolean
  children: ReactNode
}

/** A label on the left and its control (a stepper) on the right. */
export function SettingRow({ title, description, changed = false, children }: SettingRowProps) {
  return (
    <fieldset
      aria-label={changed ? `${title}, belum disimpan` : title}
      className={cn(
        'border-line-soft m-0 -mx-4 flex min-w-0 items-center justify-between gap-3 border-x-0 border-t border-b-0 px-4 py-2.5 first:border-t-0',
        changed && 'bg-highlight [&_output]:text-gold-ink',
      )}
    >
      <div className="min-w-0">
        <p className="text-sm font-semibold">{title}</p>
        {description && <p className="text-ink-muted text-xs">{description}</p>}
      </div>
      {children}
    </fieldset>
  )
}
