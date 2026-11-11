import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/route'
import { fakeFinance } from '@/test/fakeFinance'
import { bendahara, renderRoute } from '@/test/render'

describe('AccountsPage', () => {
  it('starts pembukuan only once account and kas totals match', async () => {
    fakeFinance.notStarted()
    renderRoute(ROUTES.accounts, { as: bendahara })

    await userEvent.click(await screen.findByRole('button', { name: 'Mulai pembukuan' }))
    const form = await screen.findByRole('dialog', { name: 'Mulai pembukuan' })
    await userEvent.type(within(form).getByLabelText('Tunai'), '2000000')
    await userEvent.type(within(form).getByLabelText('Kas Umum'), '1500000')
    const start = within(form).getByRole('button', { name: 'Mulai pembukuan' })
    expect(start).toBeDisabled()
    expect(within(form).getByText(/Selisih/)).toHaveTextContent(/500\.000/)

    await userEvent.click(within(form).getByLabelText('Kas Umum'))
    await userEvent.type(within(form).getByLabelText('Kas Umum'), '2000000')
    expect(start).toBeEnabled()
    await userEvent.click(start)
    const dialog = await screen.findByRole('dialog', { name: 'Mulai pembukuan?' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Mulai' }))

    expect(await screen.findByText('Pembukuan dimulai')).toBeInTheDocument()
    expect(fakeFinance.state.closings[0]?.accounts).toEqual([
      { account_id: 1, closing_balance: 2_000_000, counted_balance: null },
    ])
  })

  it('adds a bank account', async () => {
    renderRoute(ROUTES.accounts, { as: bendahara })

    await userEvent.click(await screen.findByRole('button', { name: 'Tambah akun' }))
    const form = await screen.findByRole('dialog', { name: 'Tambah akun' })
    await userEvent.type(within(form).getByLabelText('Nama akun'), 'BRI Masjid')
    await userEvent.click(within(form).getByRole('button', { name: 'Simpan' }))
    const dialog = await screen.findByRole('dialog', { name: 'Tambah akun ini?' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Simpan' }))

    expect(await screen.findByText('Akun ditambahkan')).toBeInTheDocument()
    expect(fakeFinance.state.accounts.at(-1)).toMatchObject({ name: 'BRI Masjid', kind: 'bank' })
  })
})
