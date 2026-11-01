import { screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/route'
import { fakeMosque, fakeServerClock } from '@/test/fakeApi'
import { renderRoute } from '@/test/render'

// Only Date is faked, so the page's own 1 s timer and React Query still run normally.
function at(iso: string) {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date(iso))
}

afterEach(() => {
  vi.useRealTimers()
})

describe('TvPage', () => {
  // No session anywhere in this file: the TV is public.
  it('shows the live mosque name and address with the mocked schedule', async () => {
    at('2026-10-28T10:27:15Z') // Wednesday 17:27:15 WIB
    renderRoute(ROUTES.tv)

    expect(await screen.findByText('Masjid Nur Husen')).toBeInTheDocument()
    expect(screen.getByText('Jl. Merdeka No. 1, Jakarta')).toBeInTheDocument()
    expect(screen.getByText('Rabu, 28 Oktober 2026')).toBeInTheDocument()
    expect(screen.getByText('17:27')).toBeInTheDocument()
    expect(screen.getByText('Menuju Maghrib')).toBeInTheDocument()
    expect(screen.getByText('00:47:45')).toBeInTheDocument()
    expect(screen.getByText('BERIKUTNYA').parentElement).toHaveTextContent('Maghrib18:15')
    expect(await screen.findByText(/92\.058\.114/)).toBeInTheDocument()
  })

  it('shows the uploaded logo', async () => {
    fakeMosque.logo = 'data:image/png;base64,AAAA'
    at('2026-10-28T10:27:15Z')
    const { container } = renderRoute(ROUTES.tv)

    await screen.findByText('Masjid Nur Husen')
    await vi.waitFor(() =>
      expect(container.querySelector('img[src="data:image/png;base64,AAAA"]')).not.toBeNull(),
    )
  })

  it("runs on the server's clock when the TV's is wrong", async () => {
    fakeServerClock.aheadMs = 48 * 60_000 // the TV is 48 minutes behind
    at('2026-10-28T10:27:15Z') // TV thinks 17:27, server says 18:15:15: adzan Maghrib
    renderRoute(ROUTES.tv)

    expect(await screen.findByText('TELAH MASUK WAKTU')).toBeInTheDocument()
    expect(localStorage.getItem('masjid-digital:clock-offset')).not.toBeNull()
  })

  it("keeps the mosque's time zone", async () => {
    fakeMosque.profile = { ...fakeMosque.profile!, timezone: 'Asia/Makassar' }
    at('2026-10-28T10:27:15Z') // 18:27 WITA, so Maghrib (18:15) has just passed: iqamah over
    renderRoute(ROUTES.tv)

    expect(await screen.findByText('Sholat Maghrib sedang berlangsung')).toBeInTheDocument()
    expect(screen.getByText('18:27')).toBeInTheDocument()
  })

  it('switches to adzan, then the iqamah countdown, then the dim prayer screen', async () => {
    at('2026-10-28T11:15:30Z') // 18:15:30 WIB
    const first = renderRoute(ROUTES.tv)
    expect(await screen.findByText('TELAH MASUK WAKTU')).toBeInTheDocument()
    expect(screen.getByText('Iqamah dalam 7 menit')).toBeInTheDocument()
    first.unmount()

    at('2026-10-28T11:16:48Z') // 18:16:48 WIB
    const second = renderRoute(ROUTES.tv)
    expect(await screen.findByText('05:12')).toBeInTheDocument()
    expect(screen.getByRole('progressbar', { name: 'Sisa waktu iqamah' })).toHaveAttribute(
      'aria-valuenow',
      '74',
    )
    expect(await screen.findByText('Imam: Ustadz Hasan')).toBeInTheDocument()
    second.unmount()

    at('2026-10-28T11:25:00Z') // 18:25 WIB
    renderRoute(ROUTES.tv)
    expect(await screen.findByText('Sholat Maghrib sedang berlangsung')).toBeInTheDocument()
  })

  it('shows the Jumat screen before Friday Dzuhur', async () => {
    at('2026-10-30T04:51:48Z') // Friday 11:51:48 WIB
    renderRoute(ROUTES.tv)

    expect(await screen.findByText('SHOLAT JUMAT')).toBeInTheDocument()
    expect(screen.getByText('Adzan dalam 00:17:12')).toBeInTheDocument()
    expect(await screen.findByText('Ustadz Abdullah')).toBeInTheDocument()
  })
})
