import { X } from 'lucide-react'
import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { ToastContext, type ToastInput } from '@/components/feedback/toast-context'
import { type FeedbackTone, toneIcons } from '@/components/feedback/tone-icons'
import { cn } from '@/utils/cn'

type Toast = ToastInput & { id: number; tone: FeedbackTone }

const MAX_VISIBLE = 3

// Errors and warnings need more reading time (WCAG 2.2.1); hovering or focusing pauses them.
const DEFAULT_DURATION: Record<FeedbackTone, number> = {
  success: 4000,
  info: 4000,
  warning: 6000,
  danger: 8000,
}

// Toasts float on the page, so they use dark surfaces; the tone shows in the color and the icon.
const surfaces: Record<FeedbackTone, { box: string; icon: string }> = {
  success: { box: 'bg-brand text-on-brand', icon: 'text-gold' },
  info: { box: 'bg-on-gold text-on-brand', icon: 'text-on-brand-muted' },
  warning: { box: 'bg-notice-ink text-white', icon: 'text-gold-soft' },
  danger: { box: 'bg-expense-ink text-white', icon: 'text-expense-soft' },
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])
  const nextId = useRef(0)
  const timers = useRef(new Map<number, ReturnType<typeof setTimeout>>())

  const dismiss = useCallback((id: number) => {
    clearTimeout(timers.current.get(id))
    timers.current.delete(id)
    setToasts((current) => current.filter((t) => t.id !== id))
  }, [])

  const schedule = useCallback(
    (toast: Toast) => {
      const duration = toast.duration === undefined ? DEFAULT_DURATION[toast.tone] : toast.duration
      if (duration === null) return
      clearTimeout(timers.current.get(toast.id))
      timers.current.set(
        toast.id,
        setTimeout(() => dismiss(toast.id), duration),
      )
    },
    [dismiss],
  )

  const pause = useCallback((id: number) => {
    clearTimeout(timers.current.get(id))
    timers.current.delete(id)
  }, [])

  const show = useCallback(
    (input: ToastInput) => {
      const toast: Toast = { ...input, id: nextId.current++, tone: input.tone ?? 'success' }
      // Keep the newest few so the stack never covers the screen. A dropped toast's timer
      // later dismisses an id that is already gone, which is a no-op.
      setToasts((current) => [...current, toast].slice(-MAX_VISIBLE))
      schedule(toast)
      return toast.id
    },
    [schedule],
  )

  const value = useMemo(() => ({ show, dismiss }), [show, dismiss])

  useEffect(() => {
    const pending = timers.current
    return () => pending.forEach(clearTimeout)
  }, [])

  return (
    <ToastContext value={value}>
      {children}
      {/* Always mounted so screen readers register the live region before the first toast. */}
      <output
        aria-live="polite"
        className="max-w-app pointer-events-none fixed inset-x-0 bottom-[calc(5.5rem+env(safe-area-inset-bottom))] z-50 mx-auto flex flex-col gap-2 px-5"
      >
        {toasts.map((toast) => {
          const Icon = toneIcons[toast.tone]
          const surface = surfaces[toast.tone]
          return (
            // biome-ignore lint/a11y/noStaticElementInteractions: hover and focus only pause the timer (WCAG 2.2.1); the toast's buttons are the controls
            <div
              key={toast.id}
              role={toast.tone === 'danger' ? 'alert' : undefined}
              onPointerEnter={() => pause(toast.id)}
              onPointerLeave={() => schedule(toast)}
              onFocus={() => pause(toast.id)}
              onBlur={() => schedule(toast)}
              className={cn(
                'animate-toast-in shadow-toast pointer-events-auto flex items-start gap-3 rounded-xl py-3 pr-2 pl-4',
                surface.box,
              )}
            >
              <Icon aria-hidden className={cn('mt-0.5 size-5 shrink-0', surface.icon)} />
              <div className="min-w-0 flex-1 py-0.5">
                <p className="text-base font-semibold">{toast.message}</p>
                {toast.description && (
                  <p className="mt-0.5 text-sm opacity-85">{toast.description}</p>
                )}
              </div>
              {toast.action && (
                <button
                  type="button"
                  onClick={() => {
                    toast.action?.onClick()
                    dismiss(toast.id)
                  }}
                  className="h-8 shrink-0 rounded-md px-2 text-sm font-bold underline underline-offset-2"
                >
                  {toast.action.label}
                </button>
              )}
              <button
                type="button"
                aria-label="Tutup notifikasi"
                onClick={() => dismiss(toast.id)}
                className="inline-flex size-8 shrink-0 items-center justify-center rounded-md opacity-70 hover:opacity-100"
              >
                <X aria-hidden className="size-4" />
              </button>
            </div>
          )
        })}
      </output>
    </ToastContext>
  )
}
