import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { renderRoute } from '@/test/render'

describe('routes', () => {
  it('renders Beranda at the root', async () => {
    renderRoute('/')

    expect(await screen.findByRole('heading', { name: 'Panel Pengurus' })).toBeInTheDocument()
  })

  it('renders the not-found page with a way back', async () => {
    renderRoute('/tidak-ada')

    expect(
      await screen.findByRole('heading', { name: 'Halaman tidak ditemukan' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Ke Beranda' })).toHaveAttribute('href', '/')
  })

  it('serves the component gallery in development', async () => {
    renderRoute('/kit')

    expect(await screen.findByRole('heading', { name: 'Design system' })).toBeInTheDocument()
  })
})
