import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { bendahara, renderRoute, superAdmin } from '@/test/render'

const kelola = () => screen.getByRole('navigation', { name: 'Kelola' })

describe('HomePage', () => {
  it('greets a super admin by name and shows every menu', async () => {
    renderRoute('/', { as: superAdmin })

    expect(
      await screen.findByText(/Assalamu'alaikum, Ahmad Fauzi · Super admin/),
    ).toBeInTheDocument()
    expect(within(kelola()).getAllByRole('link')).toHaveLength(8)
    // Status card plus menu tile.
    expect(screen.getByRole('link', { name: /1 online · 1 offline/ })).toBeInTheDocument()
  })

  it('shows the bendahara only what the role may open', async () => {
    renderRoute('/', { as: bendahara })

    await screen.findByText(/Assalamu'alaikum · Bendahara/)
    const links = within(kelola())
      .getAllByRole('link')
      .map((link) => link.textContent)
    expect(links).toEqual(['Keuangan'])
    expect(screen.queryByText('Layar TV')).not.toBeInTheDocument()
  })

  it('shows the next prayer with a countdown', async () => {
    renderRoute('/', { as: superAdmin })

    const card = (await screen.findByText('Sholat berikutnya')).closest('div.rounded-2xl')!
    expect(
      within(card as HTMLElement).getByText(/^(Subuh|Dzuhur|Ashar|Maghrib|Isya) \d{2}:\d{2}$/),
    ).toBeInTheDocument()
    expect(card.querySelector('time')).toHaveTextContent(/^\d{2}:\d{2}$/)
  })

  it('logs out from the account sheet', async () => {
    renderRoute('/', { as: superAdmin })

    await userEvent.click(await screen.findByRole('button', { name: 'Akun saya' }))
    await userEvent.click(screen.getByRole('button', { name: 'Keluar' }))

    expect(await screen.findByRole('heading', { name: 'Masuk' })).toBeInTheDocument()
  })
})
