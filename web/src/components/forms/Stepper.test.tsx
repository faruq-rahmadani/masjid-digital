import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'
import { Stepper } from './Stepper'

function Harness({ initial }: { initial: number }) {
  const [value, setValue] = useState(initial)
  return (
    <Stepper
      label="koreksi Subuh"
      min={-10}
      max={10}
      value={value}
      onChange={setValue}
      format={(v) => (v > 0 ? `+${v}` : String(v))}
    />
  )
}

describe('Stepper', () => {
  it('steps the value and formats it', async () => {
    render(<Harness initial={2} />)

    await userEvent.click(screen.getByRole('button', { name: 'Tambah koreksi Subuh' }))
    expect(screen.getByRole('status')).toHaveTextContent('+3')

    await userEvent.click(screen.getByRole('button', { name: 'Kurangi koreksi Subuh' }))
    await userEvent.click(screen.getByRole('button', { name: 'Kurangi koreksi Subuh' }))
    expect(screen.getByRole('status')).toHaveTextContent('+1')
  })

  it('disables the button at each bound (PRD: correction −10..+10)', () => {
    render(<Harness initial={10} />)

    expect(screen.getByRole('button', { name: 'Tambah koreksi Subuh' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Kurangi koreksi Subuh' })).toBeEnabled()
  })
})
