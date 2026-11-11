import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/route'
import { fakeFinance } from '@/test/fakeFinance'
import { bendahara, renderRoute } from '@/test/render'

const actionsOf = async (name: string) => {
  await userEvent.click(await screen.findByRole('button', { name: `Aksi untuk ${name}` }))
  const menu = await screen.findByRole('dialog', { name })
  return within(menu)
    .getAllByRole('button')
    .map((item) => item.textContent)
    .filter((text) => text !== '')
}

describe('CategoriesPage', () => {
  it('adds an own category after confirming', async () => {
    renderRoute(ROUTES.categories, { as: bendahara })

    await userEvent.type(await screen.findByLabelText('Kategori pemasukan baru'), 'Infaq Idul Adha')
    await userEvent.click(screen.getByRole('button', { name: 'Tambah' }))
    const dialog = await screen.findByRole('dialog', { name: 'Tambah kategori "Infaq Idul Adha"?' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Tambah' }))

    expect(await screen.findByText('Kategori ditambahkan')).toBeInTheDocument()
    expect(fakeFinance.state.categories.at(-1)).toMatchObject({
      name: 'Infaq Idul Adha',
      type: 'in',
      scope: 'umum_terikat',
    })
  })

  it('lets an asnaf be hidden but never renamed or deleted', async () => {
    renderRoute(ROUTES.categories, { as: bendahara })

    await userEvent.click(await screen.findByRole('tab', { name: 'Pengeluaran' }))
    expect(await actionsOf('Fakir')).toEqual(['Arsipkan'])
  })

  it('offers rename and delete for an unused own category', async () => {
    renderRoute(ROUTES.categories, { as: bendahara })

    await userEvent.click(await screen.findByRole('tab', { name: 'Pengeluaran' }))
    expect(await actionsOf('Listrik')).toEqual(['Ganti nama', 'Hapus'])
  })
})
