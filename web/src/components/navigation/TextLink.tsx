import type { ComponentProps } from 'react'
import { Link } from 'react-router'
import { cn } from '@/utils/cn'

type TextLinkProps = ComponentProps<typeof Link> & { size?: 'xs' | 'sm' }

// Inline text link ("Lihat semua", "Ubah", "Kelola kategori"). Underlined so it reads as a link.
export function TextLink({ size = 'xs', className, ...props }: TextLinkProps) {
  return (
    <Link
      className={cn(
        'text-ink decoration-line hover:decoration-ink font-semibold underline underline-offset-2',
        size === 'xs' ? 'text-xs' : 'text-sm',
        className,
      )}
      {...props}
    />
  )
}
