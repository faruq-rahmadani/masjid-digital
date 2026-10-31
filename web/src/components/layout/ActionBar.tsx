import type { ReactNode } from 'react'

/*
 * Sticky footer holding a screen's primary action, e.g. "Simpan dan kirim ke layar".
 * Stays above the home indicator and keeps the button in thumb reach on long forms.
 */
export function ActionBar({ children }: { children: ReactNode }) {
  return (
    <div className="sticky bottom-0 z-10 bg-canvas/95 px-5 pt-4 pb-[max(1.5rem,env(safe-area-inset-bottom))] backdrop-blur-sm">
      {children}
    </div>
  )
}
