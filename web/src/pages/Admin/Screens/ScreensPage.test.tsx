import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/route'
import { fakeCalls, fakeScreens } from '@/test/fakeApi'
import { bendahara, renderRoute, takmir } from '@/test/render'

// The fake seeds TV Ruang Utama (online), TV Teras (offline), a new screen from ?layar=aula and a
// blocked TV Lama.

const confirm = async (title: string, button: string) => {
  const dialog = await screen.findByRole('dialog', { name: title })
  await userEvent.click(within(dialog).getByRole('button', { name: button }))
}

const card = (name: string) =>
  screen.getByRole('heading', { name }).closest('div.flex-col') as HTMLElement

describe('ScreensPage', () => {
  it('lists the valid screens and keeps the blocked one apart', async () => {
    renderRoute(ROUTES.screens, { as: takmir })

    expect(await screen.findByText('Layar terdaftar (3)')).toBeInTheDocument()
    expect(screen.getByText(`${window.location.origin}/tv`)).toBeInTheDocument()
    expect(screen.getByRole('img', { name: 'QR code alamat layar TV' })).toBeInTheDocument()

    expect(within(card('TV Ruang Utama')).getByText('Online')).toBeInTheDocument()
    expect(within(card('TV Teras')).getByText('Offline')).toBeInTheDocument()
    expect(within(card('TV Teras')).getByText(/konten baru belum masuk/)).toBeInTheDocument()
    expect(within(card('Layar baru · aula')).getByText('BARU')).toBeInTheDocument()
    expect(screen.queryByText('TV Lama')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Lihat layar diblokir (1)' })).toBeInTheDocument()
  })

  it('names a new screen from its ?layar= hint, after confirming', async () => {
    renderRoute(ROUTES.screens, { as: takmir })

    await userEvent.click(
      within(
        await screen
          .findByText('Layar baru · aula')
          .then((h) => h.closest('div.flex-col') as HTMLElement),
      ).getByRole('button', { name: 'Beri nama' }),
    )
    const name = screen.getByLabelText('Nama layar')
    expect(name).toHaveValue('aula')
    await userEvent.clear(name)
    await userEvent.type(name, 'TV Aula')
    await userEvent.click(screen.getByRole('button', { name: 'Simpan nama' }))
    expect(fakeScreens.list.find((s) => s.id === 3)?.name).toBe('') // not before confirming
    await confirm('Simpan nama layar?', 'Simpan')

    expect(await screen.findByRole('heading', { name: 'TV Aula' })).toBeInTheDocument()
    expect(fakeScreens.list.find((s) => s.id === 3)?.name).toBe('TV Aula')
  })

  it('forces a mode until midnight and reloads a screen, each after confirming', async () => {
    renderRoute(ROUTES.screens, { as: takmir })

    await screen.findByText('Layar terdaftar (3)')
    await userEvent.click(within(card('TV Ruang Utama')).getByRole('combobox'))
    await userEvent.click(screen.getByRole('option', { name: 'Layar gelap' }))
    await confirm('Ubah mode TV Ruang Utama?', 'Ubah mode')

    expect(
      await within(card('TV Ruang Utama')).findByText(/lalu kembali ikut jadwal/),
    ).toBeInTheDocument()
    expect(fakeScreens.list[0]?.forced_mode).toBe('dark')

    await userEvent.click(within(card('TV Ruang Utama')).getByRole('button', { name: 'Reload' }))
    await confirm('Muat ulang TV Ruang Utama?', 'Muat ulang')
    expect(await screen.findByText('Layar dimuat ulang')).toBeInTheDocument()
    expect(fakeCalls.screenCommands).toEqual(['reload:1'])
  })

  it('blocks a screen and lets it be unblocked from the blocked list', async () => {
    renderRoute(ROUTES.screens, { as: takmir })

    await screen.findByText('Layar terdaftar (3)')
    await userEvent.click(within(card('TV Teras')).getByRole('button', { name: 'Blokir' }))
    await confirm('Blokir TV Teras?', 'Blokir')

    expect(await screen.findByText('Layar terdaftar (2)')).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'TV Teras' })).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Lihat layar diblokir (2)' }))
    const blocked = screen.getByRole('dialog', { name: 'Layar diblokir (2)' })
    const row = within(blocked).getByText('TV Teras').closest('li') as HTMLElement
    await userEvent.click(within(row).getByRole('button', { name: 'Buka blokir' }))
    await confirm('Buka blokir TV Teras?', 'Buka blokir')

    expect(await screen.findByText('Layar terdaftar (3)')).toBeInTheDocument()
    expect(fakeScreens.list.find((s) => s.id === 2)?.blocked_at).toBeNull()
  })

  it('is not open to the bendahara', async () => {
    renderRoute(ROUTES.screens, { as: bendahara })

    expect(await screen.findByRole('heading', { name: 'Tidak ada akses' })).toBeInTheDocument()
  })
})
