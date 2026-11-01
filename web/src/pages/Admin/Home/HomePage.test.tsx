import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/client'
import { fakeMosqueProfileService } from '@/test/fakeApi'
import { bendahara, renderRoute, superAdmin } from '@/test/render'

const kelola = () => screen.getByRole('navigation', { name: 'Kelola' })

describe('HomePage', () => {
  it('greets a super admin by name and shows every menu', async () => {
    renderRoute('/', { as: superAdmin })

    expect(
      await screen.findByText(/Assalamu'alaikum, Ahmad Fauzi · Super admin/),
    ).toBeInTheDocument()
    expect(await screen.findByRole('heading', { name: 'Masjid Nur Husen' })).toBeInTheDocument()
    expect(within(kelola()).getAllByRole('link')).toHaveLength(8)
    // Status card plus menu tile; screens and activity come from their services.
    expect(
      await screen.findByRole('link', { name: /1 online · 1 offline sejak 2 jam lalu/ }),
    ).toBeInTheDocument()
    expect(await screen.findByText('menambah pemasukan Rp 750.000')).toBeInTheDocument()
  })

  it('shows the bendahara only what the role may open', async () => {
    renderRoute('/', { as: bendahara })

    await screen.findByText(/Assalamu'alaikum, Siti Aminah · Bendahara/)
    const links = within(kelola())
      .getAllByRole('link')
      .map((link) => link.textContent)
    expect(links).toEqual(['Keuangan'])
    expect(screen.queryByText('Layar TV')).not.toBeInTheDocument()
  })

  it('shows the next prayer with a countdown', async () => {
    renderRoute('/', { as: superAdmin })

    // Times load from the prayer time service, so wait for them.
    await screen.findByText(/^(Subuh|Dzuhur|Ashar|Maghrib|Isya) \d{2}:\d{2}$/)
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

  it('shows the product name until the mosque profile is saved', async () => {
    vi.spyOn(fakeMosqueProfileService, 'getMosqueProfile').mockRejectedValueOnce(
      new ApiError(404, { code: 'NOT_FOUND', message: 'profil masjid belum diisi' }),
    )
    renderRoute('/', { as: superAdmin })

    expect(await screen.findByRole('heading', { name: 'Masjid Digital' })).toBeInTheDocument()
  })
})
