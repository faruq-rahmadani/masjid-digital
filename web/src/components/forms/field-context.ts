import { createContext, useContext } from 'react'

export type FieldContextValue = {
  id: string
  describedBy: string | undefined
  invalid: boolean
}

export const FieldContext = createContext<FieldContextValue | null>(null)

// Controls inside a <Field> pick up its id and ARIA wiring; outside one they are left alone.
export function useFieldControl<P extends { id?: string; 'aria-describedby'?: string }>(props: P) {
  const field = useContext(FieldContext)
  if (!field) return props
  return {
    ...props,
    id: props.id ?? field.id,
    'aria-describedby': cn(field.describedBy, props['aria-describedby']),
    'aria-invalid': field.invalid || undefined,
  }
}

function cn(...ids: (string | undefined)[]) {
  return ids.filter(Boolean).join(' ') || undefined
}
