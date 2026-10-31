import type { ComponentProps } from 'react'
import { Link } from 'react-router'
import { cn } from '@/lib/cn'

type TextLinkProps = ComponentProps<typeof Link> & { size?: 'xs' | 'sm' }

// Inline text link ("Lihat semua", "Ubah", "Kelola kategori"). Underlined so it reads as a link.
export function TextLink({ size = 'xs', className, ...props }: TextLinkProps) {
  return (
    <Link
      className={cn(
        'font-semibold text-ink underline decoration-line underline-offset-2 hover:decoration-ink',
        size === 'xs' ? 'text-xs' : 'text-sm',
        className,
      )}
      {...props}
    />
  )
}
