import { Eye, EyeOff } from 'lucide-react'
import { type ComponentProps, useState } from 'react'
import { cn } from '@/utils/cn'
import { Input } from './Input'

type PasswordInputProps = Omit<ComponentProps<typeof Input>, 'type'>

/** Masked by default, with an eye button to show or hide what was typed. Works inside a Field. */
export function PasswordInput({ className, ...props }: PasswordInputProps) {
  const [visible, setVisible] = useState(false)

  return (
    <div className="relative">
      <Input
        type={visible ? 'text' : 'password'}
        autoCapitalize="none"
        spellCheck={false}
        className={cn('pr-12', className)}
        {...props}
      />
      <button
        type="button"
        aria-label={visible ? 'Sembunyikan password' : 'Tampilkan password'}
        aria-pressed={visible}
        onClick={() => setVisible((v) => !v)}
        className="text-ink-muted hover:text-ink absolute top-1/2 right-1 inline-flex size-10 -translate-y-1/2 items-center justify-center rounded-md"
      >
        {visible ? (
          <EyeOff aria-hidden className="size-5" />
        ) : (
          <Eye aria-hidden className="size-5" />
        )}
      </button>
    </div>
  )
}
