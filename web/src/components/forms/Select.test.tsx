import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'
import { Field } from './Field'
import { Select } from './Select'

const OPTIONS = [
  { value: 'super_admin', label: 'Super admin' },
  { value: 'takmir', label: 'Takmir' },
  { value: 'bendahara', label: 'Bendahara' },
] as const

function Harness() {
  const [value, setValue] = useState<string>('takmir')
  return (
    <Field label="Peran">
      <Select options={OPTIONS} value={value} onChange={setValue} />
    </Field>
  )
}

describe('Select', () => {
  it('opens a labelled listbox and picks an option by click', async () => {
    render(<Harness />)
    const trigger = screen.getByRole('combobox', { name: 'Peran' })
    expect(trigger).toHaveTextContent('Takmir')
    expect(trigger).toHaveAttribute('aria-expanded', 'false')

    await userEvent.click(trigger)
    expect(trigger).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByRole('option', { name: 'Takmir' })).toHaveAttribute('aria-selected', 'true')

    await userEvent.click(screen.getByRole('option', { name: 'Bendahara' }))
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(trigger).toHaveTextContent('Bendahara')
    expect(trigger).toHaveFocus()
  })

  it('works with the keyboard', async () => {
    render(<Harness />)
    screen.getByRole('combobox', { name: 'Peran' }).focus()

    await userEvent.keyboard('{ArrowDown}')
    expect(screen.getByRole('listbox')).toHaveFocus()
    await userEvent.keyboard('{ArrowUp}{Enter}')

    expect(screen.getByRole('combobox', { name: 'Peran' })).toHaveTextContent('Super admin')
  })

  it('closes on Escape without picking', async () => {
    render(<Harness />)
    await userEvent.click(screen.getByRole('combobox', { name: 'Peran' }))

    await userEvent.keyboard('{ArrowDown}{Escape}')

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(screen.getByRole('combobox', { name: 'Peran' })).toHaveTextContent('Takmir')
  })

  it('opens right above the field when there is no room below', async () => {
    render(<Harness />)
    const trigger = screen.getByRole('combobox', { name: 'Peran' })
    // a field near the bottom of a 768px-high screen
    trigger.getBoundingClientRect = () => new DOMRect(16, 700, 300, 44)

    await userEvent.click(trigger)
    const list = screen.getByRole('listbox')
    // top must be reset, or the browser's `inset: 0` for [popover] pulls it to the top of the screen
    expect(list.style.top).toBe('auto')
    expect(list.style.bottom).toBe(`${window.innerHeight - 700 + 4}px`)
  })

  it('opens below the field when there is room', async () => {
    render(<Harness />)
    const trigger = screen.getByRole('combobox', { name: 'Peran' })
    trigger.getBoundingClientRect = () => new DOMRect(16, 100, 300, 44)

    await userEvent.click(trigger)
    const list = screen.getByRole('listbox')
    expect(list.style.top).toBe('148px')
    expect(list.style.bottom).toBe('auto')
  })
})
