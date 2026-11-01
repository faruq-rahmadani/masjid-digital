import { useEffect, useRef } from 'react'
import { useBlocker } from 'react-router'
import { ConfirmDialog } from '@/components/overlays/ConfirmDialog'

type LeaveGuardProps = {
  when: boolean
}

export function LeaveGuard({ when }: LeaveGuardProps) {
  const blocker = useBlocker(({ currentLocation, nextLocation }) => {
    return when && currentLocation.pathname !== nextLocation.pathname
  })
  // ConfirmDialog closes after confirming too; resetting then would cancel the navigation
  const leaving = useRef(false)

  useEffect(() => {
    if (!when) return
    const warn = (event: BeforeUnloadEvent) => event.preventDefault()
    window.addEventListener('beforeunload', warn)
    return () => window.removeEventListener('beforeunload', warn)
  }, [when])

  return (
    <ConfirmDialog
      open={blocker.state === 'blocked'}
      onClose={() => {
        if (!leaving.current) blocker.reset?.()
        leaving.current = false
      }}
      title="Tinggalkan halaman?"
      description="Perubahan yang belum disimpan akan hilang."
      confirmLabel="Tinggalkan"
      danger
      onConfirm={() => {
        leaving.current = true
        blocker.proceed?.()
      }}
    />
  )
}
