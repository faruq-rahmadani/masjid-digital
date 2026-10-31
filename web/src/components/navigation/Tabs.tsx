import { useId, useRef, type KeyboardEvent, type ReactNode } from 'react'
import { cn } from '@/lib/cn'

export type TabItem<T extends string> = { value: T; label: ReactNode; content: ReactNode }

type TabsProps<T extends string> = {
  /** Accessible name for the tab list. */
  label: string
  items: readonly TabItem<T>[]
  value: T
  onChange: (value: T) => void
  className?: string
}

/*
 * Switches between panels of content (Kategori: Pemasukan | Pengeluaran). Looks like the track
 * SegmentedControl but has tab semantics: arrow keys move between tabs, one tab stop.
 * Use SegmentedControl instead when the choice is a form value, not a view.
 */
export function Tabs<T extends string>({ label, items, value, onChange, className }: TabsProps<T>) {
  const baseId = useId()
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([])
  const activeIndex = Math.max(
    0,
    items.findIndex((item) => item.value === value),
  )

  const focusTab = (index: number) => {
    const next = (index + items.length) % items.length
    const item = items[next]
    if (!item) return
    onChange(item.value)
    tabRefs.current[next]?.focus()
  }

  const onKeyDown = (event: KeyboardEvent) => {
    const moves: Record<string, number> = {
      ArrowRight: activeIndex + 1,
      ArrowLeft: activeIndex - 1,
      Home: 0,
      End: items.length - 1,
    }
    const target = moves[event.key]
    if (target === undefined) return
    event.preventDefault()
    focusTab(target)
  }

  const active = items[activeIndex]

  return (
    <div className={cn('flex flex-col gap-3.5', className)}>
      <div
        role="tablist"
        aria-label={label}
        className="grid auto-cols-fr grid-flow-col gap-1.5 rounded-xl bg-track p-1"
      >
        {items.map((item, index) => {
          const selected = index === activeIndex
          return (
            <button
              key={item.value}
              ref={(node) => {
                tabRefs.current[index] = node
              }}
              type="button"
              role="tab"
              id={`${baseId}-tab-${index}`}
              aria-selected={selected}
              aria-controls={`${baseId}-panel`}
              tabIndex={selected ? 0 : -1}
              onClick={() => onChange(item.value)}
              onKeyDown={onKeyDown}
              className={cn(
                'h-11 rounded-md text-md font-semibold transition-colors',
                selected ? 'bg-brand font-bold text-on-brand' : 'text-ink-soft',
              )}
            >
              {item.label}
            </button>
          )
        })}
      </div>
      <div role="tabpanel" id={`${baseId}-panel`} aria-labelledby={`${baseId}-tab-${activeIndex}`}>
        {active?.content}
      </div>
    </div>
  )
}
