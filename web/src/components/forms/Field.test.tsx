import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Field } from '@/components/forms/Field'
import { Input } from '@/components/forms/Input'
import { Select } from '@/components/forms/Select'

describe('Field', () => {
  it('labels its control and links the hint', () => {
    render(
      <Field label="Nama masjid" hint="Tampil di header layar TV">
        <Input />
      </Field>,
    )

    const input = screen.getByRole('textbox', { name: 'Nama masjid' })
    expect(input).toHaveAccessibleDescription('Tampil di header layar TV')
    expect(input).not.toHaveAttribute('aria-invalid')
  })

  it('marks the control invalid and announces the error first', () => {
    render(
      <Field label="Kategori" hint="Pilih satu" error="Kategori wajib diisi">
        <Select>
          <option>Infaq Jumat</option>
        </Select>
      </Field>,
    )

    const select = screen.getByRole('combobox', { name: 'Kategori' })
    expect(select).toHaveAttribute('aria-invalid', 'true')
    expect(select).toHaveAccessibleDescription('Kategori wajib diisi Pilih satu')
  })
})
