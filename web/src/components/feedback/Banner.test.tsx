import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Banner } from './Banner'

describe('Banner', () => {
  it('renders title and body, with no close button unless dismissible', () => {
    render(
      <Banner tone="warning" title="Anda sedang offline">
        Perubahan belum bisa disimpan.
      </Banner>,
    )

    expect(screen.getByText('Anda sedang offline')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Tutup pemberitahuan' })).not.toBeInTheDocument()
  })

  it('calls onDismiss from its close button', async () => {
    const onDismiss = vi.fn()
    render(<Banner onDismiss={onDismiss}>Semua layar menerima jadwal terbaru.</Banner>)

    await userEvent.click(screen.getByRole('button', { name: 'Tutup pemberitahuan' }))

    expect(onDismiss).toHaveBeenCalledOnce()
  })
})
