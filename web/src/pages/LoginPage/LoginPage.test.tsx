import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/route'
import { FAKE_PASSWORD, fakeMosque } from '@/test/fakeApi'
import { renderRoute } from '@/test/render'

async function logIn(identifier: string, password: string) {
  await userEvent.type(await screen.findByLabelText('Email atau nomor HP'), identifier)
  await userEvent.type(screen.getByLabelText('Password'), password)
  await userEvent.click(screen.getByRole('button', { name: 'Masuk' }))
}

describe('LoginPage', () => {
  it('validates required fields before calling the API', async () => {
    renderRoute(ROUTES.login)

    await userEvent.click(await screen.findByRole('button', { name: 'Masuk' }))

    expect(screen.getByText('Email atau nomor HP wajib diisi')).toBeInTheDocument()
    expect(screen.getByText('Password wajib diisi')).toBeInTheDocument()
  })

  it('shows the API message for wrong credentials', async () => {
    renderRoute(ROUTES.login)

    await logIn('ketua@masjid.id', 'salah-password')

    expect(await screen.findByRole('alert')).toHaveTextContent('Nomor HP/email atau password salah')
  })

  it('refuses a disabled account', async () => {
    renderRoute(ROUTES.login)

    await logIn('0812-3456-1003', FAKE_PASSWORD)

    expect(await screen.findByRole('alert')).toHaveTextContent('Akun dinonaktifkan')
  })

  it('logs in with a phone number and opens Home', async () => {
    renderRoute(ROUTES.login)

    await logIn('0812-3456-1001', FAKE_PASSWORD)

    expect(await screen.findByText(/Assalamu'alaikum, Hasan Basri · Takmir/)).toBeInTheDocument()
  })

  it('returns to the page that asked for login', async () => {
    renderRoute(ROUTES.users)

    await logIn('ketua@masjid.id', FAKE_PASSWORD)

    expect(await screen.findByRole('heading', { name: 'Pengguna' })).toBeInTheDocument()
  })

  it('ignores a ?next= that leaves the app', async () => {
    const { router } = renderRoute(`${ROUTES.login}?next=//evil.example`)

    await logIn('ketua@masjid.id', FAKE_PASSWORD)

    await screen.findByText(/Assalamu'alaikum/)
    expect(router.state.location.pathname).toBe(ROUTES.home)
  })

  it('shows the mosque name and logo from the public profile', async () => {
    fakeMosque.logo = 'data:image/png;base64,AAAA'
    renderRoute(ROUTES.login)

    const name = fakeMosque.profile?.name ?? ''
    expect(await screen.findByText(name)).toBeInTheDocument()
    expect(await screen.findByRole('img', { name: `Logo ${name}` })).toHaveAttribute(
      'src',
      'data:image/png;base64,AAAA',
    )
  })

  it('falls back to the product name and placeholder mark before the profile is saved', async () => {
    fakeMosque.profile = null
    renderRoute(ROUTES.login)

    expect(await screen.findByText('Masjid Digital')).toBeInTheDocument()
    expect(screen.getByRole('img', { name: 'Logo masjid' })).toBeInTheDocument()
  })
})
