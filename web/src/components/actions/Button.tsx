import { LoaderCircle } from 'lucide-react'
import type { ComponentProps } from 'react'
import { type ButtonStyleProps, buttonStyles } from '@/components/actions/button-styles'
import { cn } from '@/utils/cn'

type ButtonProps = ComponentProps<'button'> &
  ButtonStyleProps & {
    loading?: boolean
  }

export function Button({
  variant,
  size,
  block,
  loading = false,
  disabled,
  type = 'button',
  className,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      className={cn(buttonStyles({ variant, size, block }), className)}
      {...props}
    >
      {loading && <LoaderCircle aria-hidden className="animate-spin" />}
      {children}
    </button>
  )
}
