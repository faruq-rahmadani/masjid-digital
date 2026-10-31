import { type ComponentProps, useState } from 'react'
import { controlBase } from '@/components/forms/control-styles'
import { useFieldControl } from '@/components/forms/field-context'
import { cn } from '@/utils/cn'

type TextareaProps = ComponentProps<'textarea'> & {
  /** Show "57 / 200" under the box. Requires `maxLength`. */
  showCount?: boolean
}

export function Textarea({
  rows = 2,
  showCount = false,
  maxLength,
  className,
  onChange,
  ...props
}: TextareaProps) {
  const initial = String(props.value ?? props.defaultValue ?? '')
  const [length, setLength] = useState(initial.length)
  const count = props.value !== undefined ? String(props.value).length : length
  const control = (
    <textarea
      rows={rows}
      maxLength={maxLength}
      className={cn(controlBase, 'resize-none px-3 py-2.5', className)}
      onChange={(event) => {
        setLength(event.currentTarget.value.length)
        onChange?.(event)
      }}
      {...useFieldControl(props)}
    />
  )

  if (!showCount || maxLength === undefined) return control

  return (
    <div className="flex flex-col gap-1">
      {control}
      <span aria-live="polite" className="text-ink-muted self-end text-xs tabular-nums">
        {count} / {maxLength}
      </span>
    </div>
  )
}
