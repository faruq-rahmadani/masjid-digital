import type { ComponentProps } from 'react'
import { cn } from '@/lib/cn'
import { useFieldControl } from './field-context'

type SliderProps = Omit<ComponentProps<'input'>, 'type'> & {
  /** Text at each end, e.g. "Pelan" / "Keras". */
  minLabel?: string
  maxLabel?: string
}

// Native range input (volume). Use inside a <Field> for its label.
export function Slider({ minLabel, maxLabel, className, ...props }: SliderProps) {
  return (
    <div className={cn('flex items-center gap-3', className)}>
      {minLabel && <span className="text-xs text-ink-muted">{minLabel}</span>}
      <input
        type="range"
        className="h-touch min-w-0 flex-1 cursor-pointer accent-brand"
        {...useFieldControl(props)}
      />
      {maxLabel && <span className="text-xs text-ink-muted">{maxLabel}</span>}
    </div>
  )
}
