import { ChevronLeft, X } from 'lucide-react'
import type { ReactNode } from 'react'
import { useNavigate } from 'react-router'
import { IconButton } from '@/components/actions/IconButton'

type PageHeaderProps = {
  title: string
  /** Where the back button goes; defaults to the previous page, falling back to Home. */
  backTo?: string
  /** `close` (×) for full-screen forms such as Tambah Transaksi. */
  dismiss?: 'back' | 'close'
  /** Small text under the title, e.g. "4 akun". */
  meta?: ReactNode
  /** Right-side icon buttons, e.g. settings. */
  actions?: ReactNode
}

export function PageHeader({ title, backTo, dismiss = 'back', meta, actions }: PageHeaderProps) {
  const navigate = useNavigate()
  const goBack = () => {
    if (backTo) navigate(backTo)
    else if (window.history.state?.idx > 0) navigate(-1)
    else navigate('/')
  }

  return (
    <header className="pt-safe-5 flex items-center gap-3 px-5 pb-5">
      <IconButton
        label={dismiss === 'back' ? 'Kembali' : 'Tutup'}
        icon={dismiss === 'back' ? <ChevronLeft /> : <X />}
        onClick={goBack}
      />
      <div className="min-w-0 flex-1">
        <h1 className="truncate text-xl font-extrabold">{title}</h1>
        {meta && <p className="text-ink-muted text-xs">{meta}</p>}
      </div>
      {actions}
    </header>
  )
}
