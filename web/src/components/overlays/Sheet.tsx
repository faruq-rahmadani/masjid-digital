import { X } from 'lucide-react'
import { type ReactNode, useEffect, useId, useRef } from 'react'
import { IconButton } from '@/components/actions/IconButton'
import { cn } from '@/utils/cn'

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
    // biome-ignore lint/a11y/useKeyWithClickEvents: backdrop click closes; keyboard users get Escape, which <dialog> handles natively
    <dialog
      ref={ref}
      aria-labelledby={titleId}
      onClose={onClose}
      // A click on the dialog element itself (not its content) is a click on the backdrop.
      onClick={(event) => event.target === ref.current && onClose()}
      className={cn(
        'animate-sheet-in bg-canvas text-ink fixed inset-x-0 top-auto bottom-0 m-0 mx-auto',
        'max-w-app rounded-t-sheet flex max-h-[90dvh] w-full flex-col p-0',
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
      {footer && <div className="pb-safe-6 px-5 pt-3">{footer}</div>}
    </dialog>
  )
}
