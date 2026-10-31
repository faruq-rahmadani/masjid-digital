import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'
import { Tabs } from './Tabs'

function Harness() {
  const [value, setValue] = useState<'in' | 'out'>('in')
  return (
    <Tabs
      label="Jenis kategori"
      value={value}
      onChange={setValue}
      items={[
        { value: 'in', label: 'Pemasukan', content: 'Kotak amal harian' },
        { value: 'out', label: 'Pengeluaran', content: 'Listrik' },
      ]}
    />
  )
}

describe('Tabs', () => {
  it('shows the selected panel and moves with arrow keys', async () => {
    render(<Harness />)
    const income = screen.getByRole('tab', { name: 'Pemasukan' })

    expect(screen.getByRole('tablist', { name: 'Jenis kategori' })).toBeInTheDocument()
    expect(income).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('tabpanel')).toHaveTextContent('Kotak amal harian')

    income.focus()
    await userEvent.keyboard('{ArrowRight}')

    const expense = screen.getByRole('tab', { name: 'Pengeluaran' })
    expect(expense).toHaveFocus()
    expect(expense).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('tabpanel', { name: 'Pengeluaran' })).toHaveTextContent('Listrik')
  })

  it('keeps a single tab stop', () => {
    render(<Harness />)

    expect(screen.getByRole('tab', { name: 'Pemasukan' })).toHaveAttribute('tabindex', '0')
    expect(screen.getByRole('tab', { name: 'Pengeluaran' })).toHaveAttribute('tabindex', '-1')
  })
})
