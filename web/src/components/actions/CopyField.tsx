import { Check } from 'lucide-react'
import { useState } from 'react'
import { buttonStyles } from '@/components/actions/button-styles'
import { cn } from '@/utils/cn'

type CopyFieldProps = {
  value: string
  /** `brand` sits on a green card (Layar TV "Pasang di TV"). */
  tone?: 'default' | 'brand'
  className?: string
}

// A read-only value with a "Salin" button, e.g. the TV URL.
export function CopyField({ value, tone = 'default', className }: CopyFieldProps) {
  const [copied, setCopied] = useState(false)

  return (
    <div
      className={cn(
        'flex items-center gap-2 rounded-lg p-1.5 pl-3',
        tone === 'brand' ? 'bg-brand-2 text-on-brand' : 'border-line bg-card border',
        className,
      )}
    >
      <span className="min-w-0 flex-1 truncate text-base font-bold">{value}</span>
      <button
        type="button"
        onClick={async () => {
          await navigator.clipboard.writeText(value)
          setCopied(true)
          setTimeout(() => setCopied(false), 2000)
        }}
        className={buttonStyles({ variant: 'gold', size: 'md' })}
      >
        {copied ? (
          <>
            <Check aria-hidden /> Disalin
          </>
        ) : (
          'Salin'
        )}
      </button>
      <span aria-live="polite" className="sr-only">
        {copied ? 'Alamat disalin' : ''}
      </span>
    </div>
  )
}
