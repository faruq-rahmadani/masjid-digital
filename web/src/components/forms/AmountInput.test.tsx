import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { AmountInput } from './AmountInput'

describe('AmountInput', () => {
  it('keeps digits only and groups them', () => {
    const onChange = vi.fn()
    render(<AmountInput aria-label="Saldo" value={1500000} onChange={onChange} />)
    const input = screen.getByLabelText('Saldo')
    expect(input).toHaveValue('1.500.000')
    fireEvent.change(input, { target: { value: '2.000a0' } })
    expect(onChange).toHaveBeenLastCalledWith(20000)
    input.focus()
    fireEvent.click(input)
    expect((input as HTMLInputElement).selectionEnd).toBe('1.500.000'.length)
    expect((input as HTMLInputElement).selectionStart).toBe(0)
    fireEvent.change(input, { target: { value: '' } })
    expect(onChange).toHaveBeenLastCalledWith(null)
  })
})
