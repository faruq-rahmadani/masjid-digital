import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { Textarea } from './Textarea'

describe('Textarea', () => {
  it('counts characters against maxLength', async () => {
    render(<Textarea aria-label="Teks baru" defaultValue="Kajian" maxLength={200} showCount />)
    expect(screen.getByText('6 / 200')).toBeInTheDocument()

    await userEvent.type(screen.getByRole('textbox', { name: 'Teks baru' }), ' Ahad')

    expect(screen.getByText('11 / 200')).toBeInTheDocument()
  })
})
