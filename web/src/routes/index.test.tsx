import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/route'
import { renderRoute, superAdmin, takmir } from '@/test/render'

describe('routes', () => {
  it('sends a visitor without a session to the login page', async () => {
    renderRoute('/')

    expect(await screen.findByRole('heading', { name: 'Masuk' })).toBeInTheDocument()
  })

  it('remembers the page to return to after login', async () => {
    const { router } = renderRoute(ROUTES.users)

    await screen.findByRole('heading', { name: 'Masuk' })
    expect(router.state.location.search).toBe(`?next=${encodeURIComponent(ROUTES.users)}`)
  })

  it('shows a no-access page for a module outside the role', async () => {
    renderRoute(ROUTES.users, { as: takmir })

    expect(await screen.findByText('Menu ini tidak tersedia untuk peran Anda')).toBeInTheDocument()
  })

  it('shows "Segera hadir" for menu items not built yet', async () => {
    renderRoute(ROUTES.finance, { as: superAdmin })

    expect(await screen.findByRole('heading', { name: 'Keuangan' })).toBeInTheDocument()
    expect(screen.getByText('Segera hadir')).toBeInTheDocument()
  })

  it('renders the not-found page with a way back', async () => {
    renderRoute('/tidak-ada')

    expect(
      await screen.findByRole('heading', { name: 'Halaman tidak ditemukan' }),
    ).toBeInTheDocument()
  })

  it('serves the component gallery in development', async () => {
    renderRoute(ROUTES.kit)

    expect(await screen.findByRole('heading', { name: 'Design system' })).toBeInTheDocument()
  })
})
