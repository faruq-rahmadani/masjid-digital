import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ConfirmDialog } from '@/components/overlays/ConfirmDialog'

describe('ConfirmDialog', () => {
  it('requires a reason before a destructive confirm (PRD: edits need a reason)', async () => {
    const onConfirm = vi.fn()
    render(
      <ConfirmDialog
        open
        onClose={() => {}}
        title="Hapus transaksi?"
        confirmLabel="Hapus"
        danger
        requireReason
        onConfirm={onConfirm}
      />,
    )
    const confirm = screen.getByRole('button', { name: 'Hapus' })
    expect(confirm).toBeDisabled()

    await userEvent.type(screen.getByLabelText('Alasan (wajib)'), '  Salah input nominal  ')
    await userEvent.click(confirm)

    expect(onConfirm).toHaveBeenCalledWith('Salah input nominal')
  })
})
