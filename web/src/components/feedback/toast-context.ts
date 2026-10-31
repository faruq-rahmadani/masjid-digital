import { createContext, useContext, useMemo } from 'react'
import type { FeedbackTone } from '@/components/feedback/tone-icons'

export type ToastInput = {
  message: string
  /** Second line with detail, e.g. which item changed. */
  description?: string
  /** `danger` toasts stay longer and interrupt screen readers. Defaults to success. */
  tone?: FeedbackTone
  /** One short action, typically "Batalkan" (undo). */
  action?: { label: string; onClick: () => void }
  /** Milliseconds on screen; null keeps it until dismissed. Defaults by tone. */
  duration?: number | null
}

export type Toaster = ((toast: ToastInput) => number) & {
  dismiss: (id: number) => void
}

type ToastContextValue = {
  show: (toast: ToastInput) => number
  dismiss: (id: number) => void
}

export const ToastContext = createContext<ToastContextValue | null>(null)

/**
 * const toast = useToast()
 * toast({ message: 'Tersimpan dan dikirim ke layar' })
 * const id = toast({ message: 'Transaksi dihapus', action: { label: 'Batalkan', onClick: undo } })
 * toast.dismiss(id)
 */
export function useToast() {
  const context = useContext(ToastContext)
  if (!context) throw new Error('useToast must be used inside <ToastProvider>')
  const { show, dismiss } = context
  return useMemo<Toaster>(
    () => Object.assign((toast: ToastInput) => show(toast), { dismiss }),
    [show, dismiss],
  )
}
