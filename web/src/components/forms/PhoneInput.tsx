import type { ComponentProps } from 'react'
import { Input } from '@/components/forms/Input'
import { formatPhoneInput } from '@/utils/format/phone'

type PhoneInputProps = Omit<ComponentProps<typeof Input>, 'type' | 'inputMode'>

/*
 * Phone field that adds the dashes as you type (0812-3456-1001). Works uncontrolled, so it can take
 * react-hook-form's register(): the value is rewritten before onChange runs. The API strips the
 * dashes (normalizePhone), so this is display only.
 */
export function PhoneInput({ onChange, ...props }: PhoneInputProps) {
  return (
    <Input
      type="tel"
      inputMode="tel"
      onChange={(event) => {
        const input = event.currentTarget
        const formatted = formatPhoneInput(input.value)
        if (formatted !== input.value) {
          // Keep the caret after the same digit, not thrown to the end when editing mid-number.
          const caret = input.selectionStart ?? input.value.length
          const kept = input.value.slice(0, caret).replace(/[^\d+]/g, '').length
          input.value = formatted
          const position = caretAfter(formatted, kept)
          input.setSelectionRange(position, position)
        }
        onChange?.(event)
      }}
      {...props}
    />
  )
}

/** Index just after the `count`-th digit (or +) in `text`. */
function caretAfter(text: string, count: number) {
  if (count === 0) return 0
  let seen = 0
  for (let i = 0; i < text.length; i++) {
    if (/[\d+]/.test(text.charAt(i)) && ++seen === count) return i + 1
  }
  return text.length
}
