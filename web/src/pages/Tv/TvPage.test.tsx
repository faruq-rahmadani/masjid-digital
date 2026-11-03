import { act, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/route'
import {
  fakeFridayDuties,
  fakeIqamah,
  fakeMosque,
  fakeRunningText,
  fakeServerClock,
} from '@/test/fakeApi'
import { FakeEventSource } from '@/test/fakeEventSource'
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

  it('scrolls the active running texts at the saved speed', async () => {
    fakeRunningText.texts = fakeRunningText.texts.map((t) => ({ ...t, active: t.id !== 1 }))
    at('2026-10-28T10:27:15Z')
    renderRoute(ROUTES.tv)

    // the moving strip next to the INFO badge
    const strip = (await screen.findByText('INFO')).nextElementSibling!.firstElementChild!
    expect(strip).toHaveTextContent("Kajian Tafsir setiap Kamis ba'da Maghrib")
    expect(strip).toHaveTextContent('Infaq via QRIS')
    expect(strip).not.toHaveTextContent('Mohon matikan nada dering')
    // one loop: the joined line plus the ۞ mark's spacing, at medium speed (8 characters a second)
    const line = fakeRunningText.texts
      .filter((t) => t.active)
      .map((t) => t.text)
      .join('   •   ')
    expect(strip).toHaveStyle({ animationDuration: `${(line.length + 7) / 8}s` })
    expect(strip).toHaveTextContent('۞')
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
    // no imam card: petugas per waktu come with the daily roster (E12-02)
    expect(screen.queryByText(/^Imam:/)).not.toBeInTheDocument()
    second.unmount()

    at('2026-10-28T11:25:00Z') // 18:25 WIB
    renderRoute(ROUTES.tv)
    expect(await screen.findByText('Sholat Maghrib sedang berlangsung')).toBeInTheDocument()
  })

  it('goes fully black during the prayer when set to', async () => {
    fakeIqamah.settings = { ...fakeIqamah.settings, sholat_display: 'black' }
    at('2026-10-28T11:25:00Z') // 18:25 WIB, Maghrib prayer on
    renderRoute(ROUTES.tv)

    const black = await screen.findByRole('img', { name: 'Sholat Maghrib sedang berlangsung' })
    expect(black).toHaveClass('bg-black')
    expect(screen.queryByText('Mohon tenang dan jaga kekhusyukan')).not.toBeInTheDocument()
  })

  it('shows the Jumat screen without petugas when none were filled in', async () => {
    fakeFridayDuties.list = []
    at('2026-10-30T04:51:48Z') // Friday 11:51:48 WIB
    renderRoute(ROUTES.tv)

    expect(await screen.findByText('SHOLAT JUMAT')).toBeInTheDocument()
    expect(screen.queryByText('KHATIB')).not.toBeInTheDocument()
  })

  it('shows the Jumat screen before Friday Dzuhur', async () => {
    at('2026-10-30T04:51:48Z') // Friday 11:51:48 WIB
    renderRoute(ROUTES.tv)

    expect(await screen.findByText('SHOLAT JUMAT')).toBeInTheDocument()
    expect(screen.getByText('Adzan dalam 00:17:12')).toBeInTheDocument()
    expect(await screen.findByText('Ustadz Abdullah')).toBeInTheDocument()
  })

  describe('live connection (E9)', () => {
    it('registers with the ?layar= name and app version, and keeps its device ID', async () => {
      at('2026-10-28T10:27:15Z')
      renderRoute(`${ROUTES.tv}?layar=teras`)
      await screen.findByText('Masjid Nur Husen')

      const url = new URL(FakeEventSource.latest!.url, 'http://tv')
      expect(url.pathname).toMatch(/^\/fake\/screens\/[0-9a-f-]{32,36}\/events$/)
      expect(localStorage.getItem('masjid-digital:device-id')).toBe(url.pathname.split('/')[3])
    })

    it('goes dark when an admin forces it, until forced_until', async () => {
      at('2026-10-28T10:27:15Z')
      renderRoute(ROUTES.tv)
      await screen.findByText('Menuju Maghrib')

      act(() => {
        FakeEventSource.latest!.open()
        FakeEventSource.latest!.emit('state', {
          name: 'TV Teras',
          forced_mode: 'dark',
          forced_until: '2026-10-28T17:00:00Z',
        })
      })
      expect(await screen.findByRole('img', { name: 'Layar gelap' })).toBeInTheDocument()
      expect(screen.queryByText('Menuju Maghrib')).not.toBeInTheDocument()

      // a forced mode that already ended is ignored
      act(() =>
        FakeEventSource.latest!.emit('state', {
          name: 'TV Teras',
          forced_mode: 'dark',
          forced_until: '2026-10-28T10:00:00Z',
        }),
      )
      expect(await screen.findByText('Menuju Maghrib')).toBeInTheDocument()
    })

    it('shows the Jumat screen when forced, on any day', async () => {
      at('2026-10-28T03:00:00Z') // Wednesday 10:00 WIB
      renderRoute(ROUTES.tv)
      await screen.findByText('Masjid Nur Husen')

      act(() =>
        FakeEventSource.latest!.emit('state', {
          name: '',
          forced_mode: 'jumat',
          forced_until: '2026-10-28T17:00:00Z',
        }),
      )
      expect(await screen.findByText('SHOLAT JUMAT')).toBeInTheDocument()
    })

    it('shows when it is blocked, and comes back once unblocked', async () => {
      at('2026-10-28T10:27:15Z')
      renderRoute(ROUTES.tv)
      await screen.findByText('Menuju Maghrib')

      act(() => FakeEventSource.latest!.emit('blocked'))
      expect(await screen.findByText('Layar ini diblokir')).toBeInTheDocument()

      act(() =>
        FakeEventSource.latest!.emit('state', {
          name: '',
          forced_mode: 'auto',
          forced_until: null,
        }),
      )
      expect(await screen.findByText('Menuju Maghrib')).toBeInTheDocument()
    })

    it('says it is offline after 2 minutes without the connection', async () => {
      at('2026-10-28T10:27:15Z')
      renderRoute(ROUTES.tv)
      await screen.findByText('Menuju Maghrib')

      act(() => {
        FakeEventSource.latest!.open()
        FakeEventSource.latest!.drop()
      })
      expect(screen.queryByText(/Offline/)).not.toBeInTheDocument()

      act(() => vi.setSystemTime(new Date('2026-10-28T10:29:20Z')))
      expect(await screen.findByText('Offline · menampilkan data tersimpan')).toBeInTheDocument()

      act(() => FakeEventSource.latest!.open())
      expect(screen.queryByText(/Offline/)).not.toBeInTheDocument()
    })
  })
})
