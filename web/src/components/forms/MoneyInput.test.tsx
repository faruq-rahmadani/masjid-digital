import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'
import { MoneyInput } from './MoneyInput'

function Harness({ initial = null }: { initial?: number | null }) {
  const [value, setValue] = useState<number | null>(initial)
  return (
    <>
      <MoneyInput value={value} onChange={setValue} />
      <output data-testid="value">{String(value)}</output>
    </>
  )
}

describe('MoneyInput', () => {
  it('keeps digits only and groups them with dots', async () => {
    render(<Harness />)
    const input = screen.getByLabelText('Nominal')

    await userEvent.type(input, '75a0.000')

    expect(input).toHaveValue('750.000')
    expect(screen.getByTestId('value')).toHaveTextContent('750000')
  })

  it('replaces the whole amount when retyped after focus (no appended digits)', async () => {
    render(<Harness initial={1_000_000} />)
    const input = screen.getByLabelText('Nominal')

    await userEvent.click(input)
    await userEvent.keyboard('15000')

    expect(screen.getByTestId('value')).toHaveTextContent('15000')
  })

  it('goes back to empty when cleared', async () => {
    render(<Harness initial={500} />)

    await userEvent.clear(screen.getByLabelText('Nominal'))

    expect(screen.getByTestId('value')).toHaveTextContent('null')
  })
})
