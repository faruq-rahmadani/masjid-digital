import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ToastProvider } from '@/components/feedback/ToastProvider'
import type { ToastInput } from '@/components/feedback/toast-context'
import { useToast } from '@/components/feedback/toast-context'

function Trigger({ input }: { input: ToastInput }) {
  const toast = useToast()
  return (
    <button type="button" onClick={() => toast(input)}>
      Picu
    </button>
  )
}

function renderToaster(input: ToastInput) {
  return render(
    <ToastProvider>
      <Trigger input={input} />
    </ToastProvider>,
  )
}

describe('ToastProvider', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows a toast and removes it after its duration', () => {
    vi.useFakeTimers()
    renderToaster({ message: 'Tersimpan dan dikirim ke layar' })

    act(() => screen.getByRole('button', { name: 'Picu' }).click())
    expect(screen.getByText('Tersimpan dan dikirim ke layar')).toBeInTheDocument()

    act(() => vi.advanceTimersByTime(4000))
    expect(screen.queryByText('Tersimpan dan dikirim ke layar')).not.toBeInTheDocument()
  })

  it('runs the action once and closes the toast', async () => {
    const undo = vi.fn()
    renderToaster({ message: 'Slide dinonaktifkan', action: { label: 'Batalkan', onClick: undo } })

    await userEvent.click(screen.getByRole('button', { name: 'Picu' }))
    await userEvent.click(screen.getByRole('button', { name: 'Batalkan' }))

    expect(undo).toHaveBeenCalledOnce()
    expect(screen.queryByText('Slide dinonaktifkan')).not.toBeInTheDocument()
  })

  it('announces danger toasts as alerts and can be closed', async () => {
    renderToaster({ tone: 'danger', message: 'Gagal menyimpan', duration: null })

    await userEvent.click(screen.getByRole('button', { name: 'Picu' }))
    expect(screen.getByRole('alert')).toHaveTextContent('Gagal menyimpan')

    await userEvent.click(screen.getByRole('button', { name: 'Tutup notifikasi' }))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('keeps at most three toasts on screen', async () => {
    renderToaster({ message: 'Tersimpan', duration: null })
    const trigger = screen.getByRole('button', { name: 'Picu' })

    for (let i = 0; i < 5; i++) await userEvent.click(trigger)

    expect(screen.getAllByText('Tersimpan')).toHaveLength(3)
  })
})
