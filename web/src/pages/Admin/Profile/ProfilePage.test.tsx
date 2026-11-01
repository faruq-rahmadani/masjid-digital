import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/route'
import { fakeMosque } from '@/test/fakeApi'
import { bendahara, renderRoute, superAdmin, takmir } from '@/test/render'

// jsdom has no layout for Leaflet; a tap on this stand-in places the pin like a tap on the map.
vi.mock('./components/LocationMap', () => ({
  LocationMap: ({
    onChange,
  }: {
    onChange: (v: { latitude: number; longitude: number }) => void
  }) => (
    <button type="button" onClick={() => onChange({ latitude: -2.9761, longitude: 104.7754 })}>
      Ketuk peta
    </button>
  ),
}))

const nameInput = () => screen.findByLabelText('Nama masjid')
const save = () => userEvent.click(screen.getByRole('button', { name: 'Simpan profil' }))
const timezone = (label: string) =>
  within(screen.getByRole('group', { name: 'Zona waktu' })).getByRole('button', { name: label })

describe('ProfilePage', () => {
  it('opens with the saved profile', async () => {
    renderRoute(ROUTES.profile, { as: superAdmin })

    expect(await nameInput()).toHaveValue('Masjid Nur Husen')
    expect(screen.getByLabelText('Alamat (opsional)')).toHaveValue('Jl. Merdeka No. 1, Jakarta')
    expect(screen.getByLabelText('Kontak pengurus (opsional)')).toHaveValue('0812-3456-1000')
    expect(screen.getByText('-6.2088, 106.8456')).toBeInTheDocument()
    expect(timezone('WIB')).toHaveAttribute('aria-pressed', 'true')
  })

  it('lets a takmir save changes', async () => {
    renderRoute(ROUTES.profile, { as: takmir })

    const name = await nameInput()
    await userEvent.clear(name)
    await userEvent.type(name, 'Masjid Al-Ikhlas')
    await userEvent.clear(screen.getByLabelText('Kontak pengurus (opsional)'))
    await save()

    expect(await screen.findByText('Profil masjid disimpan')).toBeInTheDocument()
    expect(fakeMosque.profile).toMatchObject({ name: 'Masjid Al-Ikhlas', contact_phone: '' })
  })

  it('adds the dashes to the contact number while typing', async () => {
    renderRoute(ROUTES.profile, { as: superAdmin })

    const phone = await screen.findByLabelText('Kontak pengurus (opsional)')
    await userEvent.clear(phone)
    await userEvent.type(phone, '081298765432')
    expect(phone).toHaveValue('0812-9876-5432')

    await save()
    expect(await screen.findByText('Profil masjid disimpan')).toBeInTheDocument()
    expect(fakeMosque.profile?.contact_phone).toBe('+6281298765432')
  })

  it('validates like the API', async () => {
    renderRoute(ROUTES.profile, { as: superAdmin })

    await userEvent.clear(await nameInput())
    await userEvent.type(screen.getByLabelText('Kontak pengurus (opsional)'), '-555')
    await save()

    expect(screen.getByText('Nama masjid wajib diisi')).toBeInTheDocument()
    expect(screen.getByText(/Nomor HP tidak valid/)).toBeInTheDocument()
    expect(fakeMosque.profile?.name).toBe('Masjid Nur Husen')
  })

  it('asks for a location on a fresh install, then guesses the time zone from it', async () => {
    fakeMosque.profile = null
    renderRoute(ROUTES.profile, { as: superAdmin })

    expect(await screen.findByText('Profil belum diisi')).toBeInTheDocument()
    await userEvent.type(await nameInput(), 'Masjid Raya')
    await save()
    expect(screen.getByText('Tentukan lokasi masjid di peta')).toBeInTheDocument()

    await userEvent.type(
      screen.getByRole('searchbox', { name: 'Cari lokasi' }),
      'masjid raya{Enter}',
    )
    const results = await screen.findByRole('list', { name: 'Hasil pencarian lokasi' })
    await userEvent.click(within(results).getByRole('button', { name: /Masjid Raya Makassar/ }))

    expect(screen.getByText('-5.1336, 119.4183')).toBeInTheDocument()
    expect(timezone('WITA')).toHaveAttribute('aria-pressed', 'true')
    await save()
    expect(await screen.findByText('Profil masjid disimpan')).toBeInTheDocument()
    expect(fakeMosque.profile).toMatchObject({
      name: 'Masjid Raya',
      latitude: -5.1336,
      longitude: 119.4183,
      timezone: 'Asia/Makassar',
    })
  })

  it('moves the pin from the map and keeps a manual time zone choice', async () => {
    renderRoute(ROUTES.profile, { as: superAdmin })

    await userEvent.click(await screen.findByRole('button', { name: 'Ketuk peta' }))
    expect(screen.getByText('-2.9761, 104.7754')).toBeInTheDocument()

    await userEvent.click(timezone('WIT'))
    await save()
    expect(await screen.findByText('Profil masjid disimpan')).toBeInTheDocument()
    expect(fakeMosque.profile).toMatchObject({ latitude: -2.9761, timezone: 'Asia/Jayapura' })
  })

  it('says when a search finds nothing', async () => {
    renderRoute(ROUTES.profile, { as: superAdmin })

    await userEvent.type(
      await screen.findByRole('searchbox', { name: 'Cari lokasi' }),
      'tidak ada{Enter}',
    )
    expect(await screen.findByText(/Tidak ditemukan/)).toBeInTheDocument()
  })

  it('uploads, replaces and deletes the logo', async () => {
    renderRoute(ROUTES.profile, { as: superAdmin })

    const png = new File(['png'], 'logo.png', { type: 'image/png' })
    await userEvent.upload(await screen.findByLabelText('Unggah logo'), png)
    expect(await screen.findByRole('img', { name: 'Logo masjid' })).toHaveAttribute(
      'src',
      'data:image/png;base64,AAAA',
    )
    expect(screen.getByText('Logo masjid disimpan')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Hapus' }))
    await userEvent.click(screen.getByRole('button', { name: 'Hapus logo' }))
    expect(await screen.findByText('Logo masjid dihapus')).toBeInTheDocument()
    expect(fakeMosque.logo).toBeNull()
    expect(screen.getByLabelText('Unggah logo')).toBeInTheDocument()
  })

  it('rejects an SVG logo before uploading', async () => {
    renderRoute(ROUTES.profile, { as: superAdmin })

    const svg = new File(['<svg/>'], 'logo.svg', { type: 'image/svg+xml' })
    // The picker's `accept` would hide it; drop that filter to simulate a file chosen anyway.
    await userEvent.upload(await screen.findByLabelText('Unggah logo'), svg, { applyAccept: false })
    expect(await screen.findByText('Format logo harus PNG, JPG, atau WebP')).toBeInTheDocument()
    expect(fakeMosque.logo).toBeNull()
  })

  it('is hidden from the bendahara', async () => {
    renderRoute(ROUTES.profile, { as: bendahara })

    expect(await screen.findByText('Menu ini tidak tersedia untuk peran Anda')).toBeInTheDocument()
  })
})
