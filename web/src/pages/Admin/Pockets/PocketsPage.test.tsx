import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/route'
import { fakeFinance } from '@/test/fakeFinance'
import { bendahara, renderRoute, takmir } from '@/test/render'

const confirm = async (title: string, button: string) => {
  const dialog = await screen.findByRole('dialog', { name: title })
  await userEvent.click(within(dialog).getByRole('button', { name: button }))
}

describe('PocketsPage', () => {
  it('adds a terikat kas with its purpose, target and surplus plan', async () => {
    renderRoute(ROUTES.pockets, { as: bendahara })

    await userEvent.click(await screen.findByRole('button', { name: 'Tambah kas' }))
    const form = await screen.findByRole('dialog', { name: 'Kas baru' })
    await userEvent.type(within(form).getByLabelText('Nama kas'), 'Santunan Yatim')
    await userEvent.type(within(form).getByLabelText('Peruntukan'), 'Santunan bulanan anak yatim')
    await userEvent.type(within(form).getByLabelText(/Target/), '5000000')
    await userEvent.click(within(form).getByLabelText(/Minta persetujuan donatur dulu/))
    await userEvent.click(within(form).getByRole('button', { name: 'Buat kas' }))
    await confirm('Buat kas baru?', 'Simpan')

    expect(await screen.findByText('Kas dibuat')).toBeInTheDocument()
    expect(fakeFinance.state.pockets.at(-1)).toMatchObject({
      kind: 'terikat',
      name: 'Santunan Yatim',
      target_amount: 5_000_000,
      surplus_plan: 'ask_donor',
    })
  })

  it('keeps a built-in kas locked except for the TV', async () => {
    renderRoute(ROUTES.pockets, { as: bendahara })

    await userEvent.click(await screen.findByText('Zakat Mal'))
    const form = await screen.findByRole('dialog', { name: 'Zakat Mal' })
    expect(
      within(form).getByText(/Kas bawaan beserta aturannya tidak bisa diubah/),
    ).toBeInTheDocument()
    expect(within(form).queryByLabelText('Nama kas')).not.toBeInTheDocument()
    expect(within(form).queryByRole('button', { name: 'Arsipkan kas' })).not.toBeInTheDocument()

    await userEvent.click(within(form).getByLabelText('Tampil di layar TV'))
    await userEvent.click(within(form).getByRole('button', { name: 'Simpan' }))
    await confirm('Simpan perubahan kas?', 'Simpan')
    await screen.findByText('Kas disimpan')
    expect(fakeFinance.state.pockets[1]).toMatchObject({ name: 'Zakat Mal', show_on_tv: true })
  })

  it('shows takmir the kas without a way to add one', async () => {
    renderRoute(ROUTES.pockets, { as: takmir })

    expect(await screen.findByText('Renovasi Atap')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Tambah kas' })).not.toBeInTheDocument()
  })
})
