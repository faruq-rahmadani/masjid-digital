import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/client'
import {
  FAKE_PASSWORD,
  fakeMosque,
  fakeMosqueProfileService,
  fakeServerClock,
} from '@/test/fakeApi'
import { bendahara, renderRoute, superAdmin } from '@/test/render'

const kelola = () => screen.getByRole('navigation', { name: 'Kelola' })

// Only Date is faked, so timers and React Query run normally. The fake seeds a Maghrib override of
// 18:15 on 2026-10-28, which pins the countdown.
function at(iso: string) {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date(iso))
}

const nextPrayerCard = async () =>
  (await screen.findByText('Sholat berikutnya')).closest('div.rounded-2xl') as HTMLElement

afterEach(() => {
  vi.useRealTimers()
})

describe('HomePage', () => {
  it('greets a super admin by name and shows every menu', async () => {
    renderRoute('/', { as: superAdmin })

    expect(
      await screen.findByText(/Assalamu'alaikum, Ahmad Fauzi · Super admin/),
    ).toBeInTheDocument()
    expect(await screen.findByRole('heading', { name: 'Masjid Nur Husen' })).toBeInTheDocument()
    expect(within(kelola()).getAllByRole('link')).toHaveLength(9)
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

  it("counts down on the server's clock when the phone's is wrong", async () => {
    at('2026-10-28T10:27:00Z') // the phone says 17:27 WIB
    fakeServerClock.aheadMs = 30 * 60_000 // the server says 17:57
    renderRoute('/', { as: superAdmin })

    const card = await nextPrayerCard()
    expect(await within(card).findByText('Maghrib 18:15')).toBeInTheDocument()
    await vi.waitFor(() => expect(card.querySelector('time')).toHaveTextContent('00:18'))
  })

  it("counts down on the mosque's clock, not the phone's zone", async () => {
    fakeMosque.profile = { ...fakeMosque.profile!, timezone: 'Asia/Makassar' }
    at('2026-10-28T10:27:00Z') // 17:27 WIB on a Jakarta phone, but 18:27 WITA at the mosque
    renderRoute('/', { as: superAdmin })

    const card = await nextPrayerCard()
    // Maghrib (18:15) has passed in Makassar
    expect(await within(card).findByText(/^Isya \d{2}:\d{2}$/)).toBeInTheDocument()
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

  it('changes the own password from Akun saya', async () => {
    renderRoute('/', { as: bendahara })

    await userEvent.click(await screen.findByRole('button', { name: 'Akun saya' }))
    await userEvent.click(screen.getByRole('button', { name: 'Ganti password' }))
    expect(screen.getByRole('dialog', { name: 'Ganti password' })).toBeInTheDocument()

    await userEvent.type(screen.getByLabelText('Password lama'), 'salah-sekali')
    await userEvent.type(screen.getByLabelText('Password baru'), 'password-baru-1')
    await userEvent.click(screen.getByRole('button', { name: 'Simpan password' }))
    expect(await screen.findByText('Password lama salah')).toBeInTheDocument()

    await userEvent.clear(screen.getByLabelText('Password lama'))
    await userEvent.type(screen.getByLabelText('Password lama'), FAKE_PASSWORD)
    await userEvent.click(screen.getByRole('button', { name: 'Simpan password' }))

    expect(await screen.findByText('Password diganti')).toBeInTheDocument()
    expect(screen.queryByRole('dialog', { name: 'Ganti password' })).not.toBeInTheDocument()
  })
})
