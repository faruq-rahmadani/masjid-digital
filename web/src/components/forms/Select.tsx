import { Check, ChevronDown } from 'lucide-react'
import {
  type KeyboardEvent,
  type ReactNode,
  useEffect,
  useId,
  useLayoutEffect,
  useRef,
  useState,
} from 'react'
import { type ControlSize, controlBase, controlSizes } from '@/components/forms/control-styles'
import { useFieldControl } from '@/components/forms/field-context'
import { cn } from '@/utils/cn'

export type SelectOption<T extends string> = { value: T; label: ReactNode }

type SelectProps<T extends string> = {
  options: readonly SelectOption<T>[]
  value: T | undefined
  onChange: (value: T) => void
  placeholder?: string
  size?: ControlSize | 'pill'
  disabled?: boolean
  name?: string
  id?: string
  'aria-label'?: string
  'aria-describedby'?: string
  onBlur?: () => void
  className?: string
}

const LIST_MAX_HEIGHT = 240

/*
 * A styled dropdown in place of the native <select>, whose desktop menu is small, unstyled and
 * placed by the browser. The list opens in the top layer (Popover API), so a sheet's scroll area
 * can't clip it; it sits under the field at full width and flips above when there's no room.
 * Keyboard: Enter/Space/arrows open, arrows/Home/End move, Enter picks, Escape closes, letters jump.
 */
export function Select<T extends string>({
  options,
  value,
  onChange,
  placeholder = 'Pilih',
  size = 'md',
  disabled,
  className,
  onBlur,
  ...rest
}: SelectProps<T>) {
  const listId = useId()
  const triggerRef = useRef<HTMLButtonElement>(null)
  const listRef = useRef<HTMLDivElement>(null)
  const [open, setOpen] = useState(false)
  const [active, setActive] = useState(0)
  const [position, setPosition] = useState<React.CSSProperties>({})
  const control = useFieldControl(rest)
  const selected = options.find((o) => o.value === value)

  const place = () => {
    const rect = triggerRef.current?.getBoundingClientRect()
    if (!rect) return
    const below = window.innerHeight - rect.bottom
    const above = rect.top
    const flip = below < Math.min(LIST_MAX_HEIGHT, options.length * 44 + 8) && above > below
    setPosition({
      left: rect.left,
      width: rect.width,
      ...(flip ? { bottom: window.innerHeight - rect.top + 4 } : { top: rect.bottom + 4 }),
    })
  }

  const openList = () => {
    if (disabled) return
    setActive(
      Math.max(
        0,
        options.findIndex((o) => o.value === value),
      ),
    )
    place()
    setOpen(true)
  }

  const close = (refocus = true) => {
    setOpen(false)
    if (refocus) triggerRef.current?.focus()
  }

  const pick = (index: number) => {
    const option = options[index]
    if (option) onChange(option.value)
    close()
  }

  // Show in the top layer once rendered, then move focus into the list.
  useLayoutEffect(() => {
    const list = listRef.current
    if (!open || !list) return
    if (typeof list.showPopover === 'function' && !list.matches(':popover-open')) list.showPopover()
    list.focus()
  }, [open])

  // Follow the field while the page or sheet scrolls; close if it leaves the viewport.
  useEffect(() => {
    if (!open) return
    const update = () => place()
    window.addEventListener('resize', update)
    window.addEventListener('scroll', update, true)
    return () => {
      window.removeEventListener('resize', update)
      window.removeEventListener('scroll', update, true)
    }
  })

  const onListKeyDown = (event: KeyboardEvent) => {
    const last = options.length - 1
    const moves: Record<string, number> = {
      ArrowDown: Math.min(last, active + 1),
      ArrowUp: Math.max(0, active - 1),
      Home: 0,
      End: last,
    }
    if (event.key in moves) {
      event.preventDefault()
      setActive(moves[event.key] as number)
    } else if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      pick(active)
    } else if (event.key === 'Escape') {
      // Close only the list, not the sheet around it.
      event.preventDefault()
      event.stopPropagation()
      close()
    } else if (event.key === 'Tab') {
      close(false)
    } else if (event.key.length === 1) {
      const letter = event.key.toLowerCase()
      const match = options.findIndex(
        (o, i) => i > active && String(o.label).toLowerCase().startsWith(letter),
      )
      const wrap = options.findIndex((o) => String(o.label).toLowerCase().startsWith(letter))
      if (match >= 0 || wrap >= 0) setActive(match >= 0 ? match : wrap)
    }
  }

  const onTriggerKeyDown = (event: KeyboardEvent) => {
    if (['ArrowDown', 'ArrowUp', 'Enter', ' '].includes(event.key)) {
      event.preventDefault()
      openList()
    }
  }

  return (
    <>
      <button
        ref={triggerRef}
        type="button"
        role="combobox"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={open ? listId : undefined}
        disabled={disabled}
        onClick={() => (open ? close() : openList())}
        onKeyDown={onTriggerKeyDown}
        onBlur={onBlur}
        {...control}
        className={cn(
          controlBase,
          'flex items-center justify-between gap-2 text-left',
          size === 'pill'
            ? 'inline-flex h-9 w-auto rounded-full pr-2.5 pl-3 text-sm'
            : cn(controlSizes[size], 'pr-3'),
          open && 'border-brand',
          className,
        )}
      >
        <span className={cn('truncate', !selected && 'text-ink-muted')}>
          {selected?.label ?? placeholder}
        </span>
        <ChevronDown
          aria-hidden
          className={cn(
            'text-ink-muted size-4 shrink-0 transition-transform',
            open && 'rotate-180',
          )}
        />
      </button>

      {open && (
        <div
          ref={listRef}
          id={listId}
          role="listbox"
          tabIndex={-1}
          popover="auto"
          aria-activedescendant={`${listId}-${active}`}
          onKeyDown={onListKeyDown}
          onToggle={(event) => event.newState === 'closed' && setOpen(false)}
          style={{ position: 'fixed', maxHeight: LIST_MAX_HEIGHT, ...position }}
          className="border-line bg-card text-ink shadow-overlay m-0 overflow-y-auto rounded-md border p-1 outline-none"
        >
          {options.map((option, index) => (
            <div
              key={option.value}
              id={`${listId}-${index}`}
              role="option"
              tabIndex={-1}
              aria-selected={option.value === value}
              onPointerMove={() => setActive(index)}
              onClick={() => pick(index)}
              onKeyDown={() => {}}
              className={cn(
                'min-h-touch flex cursor-pointer items-center justify-between gap-2 rounded-sm px-3 text-base',
                index === active && 'bg-highlight',
                option.value === value && 'font-semibold',
              )}
            >
              <span className="truncate">{option.label}</span>
              {option.value === value && (
                <Check aria-hidden className="text-brand size-4 shrink-0" />
              )}
            </div>
          ))}
        </div>
      )}
    </>
  )
}
