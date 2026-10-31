import type { ComponentProps } from 'react'
import { cn } from '@/utils/cn'

/*
 * Compact data table, styled after the Waktu sholat 7-day preview: small caps header, rows split
 * by a soft line, no zebra striping. Wide tables scroll sideways inside the card, never the page.
 */
type TableProps = ComponentProps<'table'> & {
  /** `compact` (12px) for dense grids such as the 7-day preview; `default` (14px) otherwise. */
  density?: 'default' | 'compact'
  /** Visible or screen-reader caption describing the table. */
  caption?: string
}

export function Table({ density = 'default', caption, className, children, ...props }: TableProps) {
  return (
    <div className="-mx-1 overflow-x-auto px-1">
      <table
        className={cn(
          'w-full border-collapse text-left tabular-nums',
          density === 'compact' ? 'text-xs' : 'text-base',
          className,
        )}
        {...props}
      >
        {caption && <caption className="sr-only">{caption}</caption>}
        {children}
      </table>
    </div>
  )
}

export function THead(props: ComponentProps<'thead'>) {
  return <thead {...props} />
}

export function TBody(props: ComponentProps<'tbody'>) {
  return <tbody {...props} />
}

export function Tr({ className, ...props }: ComponentProps<'tr'>) {
  return <tr className={cn('border-line-soft [tbody_&]:border-t', className)} {...props} />
}

export function Th({ className, scope = 'col', ...props }: ComponentProps<'th'>) {
  return (
    <th
      scope={scope}
      className={cn(
        'text-2xs text-ink-muted py-2 pr-2 font-bold tracking-wide whitespace-nowrap uppercase last:pr-0',
        // A row header (scope="row") reads as the row's label, not a column title.
        scope === 'row' && 'text-ink font-bold tracking-normal text-[length:inherit] normal-case',
        className,
      )}
      {...props}
    />
  )
}

export function Td({ className, ...props }: ComponentProps<'td'>) {
  return <td className={cn('py-2 pr-2 align-middle last:pr-0', className)} {...props} />
}
