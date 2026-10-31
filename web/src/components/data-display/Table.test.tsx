import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Table, TBody, Td, Th, THead, Tr } from './Table'

describe('Table', () => {
  it('renders a real table with column and row headers', () => {
    render(
      <Table caption="Jadwal sholat">
        <THead>
          <Tr>
            <Th>Tgl</Th>
            <Th>Sbh</Th>
          </Tr>
        </THead>
        <TBody>
          <Tr>
            <Th scope="row">Kam 29</Th>
            <Td>04:41</Td>
          </Tr>
        </TBody>
      </Table>,
    )

    const table = screen.getByRole('table', { name: 'Jadwal sholat' })
    expect(within(table).getAllByRole('columnheader')).toHaveLength(2)
    expect(within(table).getByRole('rowheader', { name: 'Kam 29' })).toBeInTheDocument()
  })
})
