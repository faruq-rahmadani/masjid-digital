import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'
import { SegmentedControl } from '@/components/forms/SegmentedControl'

function Harness() {
  const [value, setValue] = useState<'in' | 'out'>('in')
  return (
    <SegmentedControl
      label="Jenis transaksi"
      variant="track"
      value={value}
      onChange={setValue}
      tone={(v) => (v === 'in' ? 'income' : 'expense')}
      options={[
        { value: 'in', label: 'Pemasukan' },
        { value: 'out', label: 'Pengeluaran' },
      ]}
    />
  )
}

describe('SegmentedControl', () => {
  it('exposes a named group with exactly one pressed option', async () => {
    render(<Harness />)

    expect(screen.getByRole('group', { name: 'Jenis transaksi' })).toBeInTheDocument()
    const income = screen.getByRole('button', { name: 'Pemasukan' })
    const expense = screen.getByRole('button', { name: 'Pengeluaran' })
    expect(income).toHaveAttribute('aria-pressed', 'true')

    await userEvent.click(expense)

    expect(expense).toHaveAttribute('aria-pressed', 'true')
    expect(income).toHaveAttribute('aria-pressed', 'false')
    expect(expense).toHaveClass('bg-expense')
  })
})
