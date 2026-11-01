import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { Field } from './Field'
import { PasswordInput } from './PasswordInput'

describe('PasswordInput', () => {
  it('is masked by default and toggles visibility with the eye button', async () => {
    render(
      <Field label="Password baru">
        <PasswordInput defaultValue="rahasia-kuat" />
      </Field>,
    )
    const input = screen.getByLabelText('Password baru')
    expect(input).toHaveAttribute('type', 'password')

    await userEvent.click(screen.getByRole('button', { name: 'Tampilkan password' }))
    expect(input).toHaveAttribute('type', 'text')
    expect(screen.getByRole('button', { name: 'Sembunyikan password' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )

    await userEvent.click(screen.getByRole('button', { name: 'Sembunyikan password' }))
    expect(input).toHaveAttribute('type', 'password')
  })
})
