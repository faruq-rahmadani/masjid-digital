import { X } from 'lucide-react'
import { useEffect, useId, useRef, type ReactNode } from 'react'
import { cn } from '@/lib/cn'
import { IconButton } from '@/components/actions/IconButton'

type SheetProps = {
  open: boolean
  onClose: () => void
  title: ReactNode
  /** Sticky area under the content, usually the save button. */
  footer?: ReactNode
  className?: string
  children: ReactNode
}

/*
 * Bottom sheet on a native <dialog>: focus trap, Escape and inert background come from the browser.
 * Rendered only while open so its form state resets between uses.
 */
export function Sheet({ open, onClose, title, footer, className, children }: SheetProps) {
  const ref = useRef<HTMLDialogElement>(null)
  const titleId = useId()

  useEffect(() => {
    const dialog = ref.current
    if (open && dialog && !dialog.open) dialog.showModal()
  }, [open])

  if (!open) return null

  return (
    // Backdrop click closes; keyboard users get Escape, which <dialog> handles natively.
    // oxlint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-noninteractive-element-interactions
    <dialog
      ref={ref}
      aria-labelledby={titleId}
      onClose={onClose}
      // A click on the dialog element itself (not its content) is a click on the backdrop.
      onClick={(event) => event.target === ref.current && onClose()}
      className={cn(
        'fixed inset-x-0 top-auto bottom-0 m-0 mx-auto animate-sheet-in bg-canvas text-ink',
        'flex max-h-[90dvh] w-full max-w-app flex-col rounded-t-sheet p-0',
        'shadow-overlay backdrop:bg-black/40',
        className,
      )}
    >
      <div className="flex items-center gap-3 px-5 pt-5 pb-3">
        <h2 id={titleId} className="min-w-0 flex-1 text-xl font-extrabold">
          {title}
        </h2>
        <IconButton label="Tutup" icon={<X />} onClick={onClose} />
      </div>
      <div className="flex-1 overflow-y-auto px-5 pb-5">{children}</div>
      {footer && (
        <div className="px-5 pt-3 pb-[max(1.5rem,env(safe-area-inset-bottom))]">{footer}</div>
      )}
    </dialog>
  )
}
