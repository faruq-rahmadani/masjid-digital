import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/route'
import { fakeMosque } from '@/test/fakeApi'
import { renderRoute, superAdmin } from '@/test/render'

vi.mock('@/pages/Admin/Profile/components/LocationMap', () => ({ LocationMap: () => null }))

const favicon = () => document.querySelector<HTMLLinkElement>('link[rel="icon"]')

afterEach(() => {
  document.title = ''
  favicon()?.remove()
})

describe('useDocumentBranding', () => {
  it('names the tab after the mosque and uses its logo as the favicon, even before login', async () => {
    fakeMosque.logo = 'data:image/png;base64,LOGO'
    renderRoute(ROUTES.login)

    await waitFor(() => expect(document.title).toBe('Masjid Nur Husen'))
    await waitFor(() => expect(favicon()?.href).toBe('data:image/png;base64,LOGO'))
  })

  it('falls back to Masjid Digital and no favicon before the profile is saved', async () => {
    fakeMosque.profile = null
    renderRoute(ROUTES.login)

    await waitFor(() => expect(document.title).toBe('Masjid Digital'))
    expect(favicon()).toBeNull()
  })

  it('follows changes saved on the profile screen', async () => {
    renderRoute(ROUTES.profile, { as: superAdmin })

    const name = await screen.findByLabelText('Nama masjid')
    await userEvent.clear(name)
    await userEvent.type(name, 'Masjid Al-Ikhlas')
    await userEvent.click(screen.getByRole('button', { name: 'Simpan profil' }))
    await waitFor(() => expect(document.title).toBe('Masjid Al-Ikhlas'))

    await userEvent.upload(
      screen.getByLabelText('Unggah logo'),
      new File(['png'], 'logo.png', { type: 'image/png' }),
    )
    await waitFor(() => expect(favicon()?.href).toBe('data:image/png;base64,AAAA'))

    await userEvent.click(screen.getByRole('button', { name: 'Hapus' }))
    await userEvent.click(screen.getByRole('button', { name: 'Hapus logo' }))
    await waitFor(() => expect(favicon()).toBeNull())
  })

  it('brands the TV too', async () => {
    renderRoute(ROUTES.tv)

    await waitFor(() => expect(document.title).toBe('Masjid Nur Husen'))
  })
})
