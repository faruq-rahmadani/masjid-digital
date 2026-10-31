import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ActionMenu } from '@/components/overlays/ActionMenu'

describe('ActionMenu', () => {
  it('opens the row actions in a sheet and closes after choosing one', async () => {
    const deactivate = vi.fn()
    render(
      <ActionMenu
        label="[Nama Takmir]"
        actions={[
          { label: 'Ubah peran', onSelect: () => {} },
          { label: 'Nonaktifkan akun', onSelect: deactivate, danger: true },
        ]}
      />,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Aksi untuk [Nama Takmir]' }))
    const sheet = screen.getByRole('dialog', { name: '[Nama Takmir]' })
    const action = screen.getByRole('button', { name: 'Nonaktifkan akun' })
    expect(sheet).toBeVisible()
    expect(action).toHaveClass('text-expense')

    await userEvent.click(action)

    expect(deactivate).toHaveBeenCalledOnce()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})
