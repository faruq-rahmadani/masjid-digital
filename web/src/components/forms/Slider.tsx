import type { ComponentProps } from 'react'
import { useFieldControl } from '@/components/forms/field-context'
import { cn } from '@/utils/cn'

type SliderProps = Omit<ComponentProps<'input'>, 'type'> & {
  /** Text at each end, e.g. "Pelan" / "Keras". */
  minLabel?: string
  maxLabel?: string
}

// Native range input (volume). Use inside a <Field> for its label.
export function Slider({ minLabel, maxLabel, className, ...props }: SliderProps) {
  return (
    <div className={cn('flex items-center gap-3', className)}>
      {minLabel && <span className="text-ink-muted text-xs">{minLabel}</span>}
      <input
        type="range"
        className="h-touch accent-brand min-w-0 flex-1 cursor-pointer"
        {...useFieldControl(props)}
      />
      {maxLabel && <span className="text-ink-muted text-xs">{maxLabel}</span>}
    </div>
  )
}
