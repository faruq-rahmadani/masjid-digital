import { useMemo } from 'react'
import { encode } from 'uqr'
import { cn } from '@/utils/cn'

type QrCodeProps = {
  value: string
  label: string
  className?: string
}

// Draws the code as one SVG path in the current text color, so it follows the theme tokens.
export function QrCode({ value, label, className }: QrCodeProps) {
  const { size, path } = useMemo(() => {
    const { data, size } = encode(value, { border: 0 })
    const cells = data.flatMap((row, y) =>
      row.flatMap((dark, x) => (dark ? [`M${x} ${y}h1v1h-1z`] : [])),
    )
    return { size, path: cells.join('') }
  }, [value])

  return (
    <svg
      viewBox={`0 0 ${size} ${size}`}
      role="img"
      aria-label={label}
      shapeRendering="crispEdges"
      className={cn('fill-current', className)}
    >
      <path d={path} />
    </svg>
  )
}
