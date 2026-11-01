import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/route'
import { fakeIqamah } from '@/test/fakeApi'
import { renderRoute, takmir } from '@/test/render'

// Only Date is faked. The fakes pin Wednesday 2026-10-28 Maghrib at 18:15 (a date override),
// Friday Dzuhur at 12:09 (a weekly rule), Maghrib iqamah 7 min, a 1-min adzan screen and 10 min
// of mode sholat.
function at(iso: string) {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date(iso))
}

afterEach(() => {
  vi.useRealTimers()
})

const confirm = async (title: string, button: string) => {
  const dialog = await screen.findByRole('dialog', { name: title })
  await userEvent.click(within(dialog).getByRole('button', { name: button }))
}

const flowSteps = async (name: string) =>
  within(await screen.findByRole('list', { name: `Alur ${name} di layar` }))
    .getAllByRole('listitem')
    .map((li) => li.textContent)

describe('IqamahPage', () => {
  it('shows the next prayer as the TV will run it, live as the steppers change', async () => {
    at('2026-10-28T10:27:00Z') // Wednesday 17:27 WIB
    renderRoute(ROUTES.iqamah, { as: takmir })

    expect(await screen.findByText('Maghrib 18:15')).toBeInTheDocument()
    expect(await flowSteps('Maghrib')).toEqual([
      'Layar adzan18:15–18:16',
      'Hitung mundur iqamah18:16–18:22',
      'Mode sholat18:22–18:32',
    ])

    await userEvent.click(screen.getByRole('button', { name: 'Tambah Iqamah Maghrib' }))
    expect(await flowSteps('Maghrib')).toEqual([
      'Layar adzan18:15–18:16',
      'Hitung mundur iqamah18:16–18:23',
      'Mode sholat18:23–18:33',
    ])
    expect(screen.getByRole('row', { name: 'Maghrib, belum disimpan' })).toHaveClass('bg-highlight')
    expect(screen.getByText('Ada perubahan belum disimpan')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Simpan dan kirim ke layar' }))
    expect(fakeIqamah.settings.iqamah_min.maghrib).toBe(7) // not before confirming
    await confirm('Simpan pengaturan iqamah?', 'Simpan')
    expect(await screen.findByText('Pengaturan iqamah disimpan')).toBeInTheDocument()
    expect(fakeIqamah.settings.iqamah_min.maghrib).toBe(8)
  })

  it('shows Friday Dzuhur as Jumat', async () => {
    at('2026-10-30T03:00:00Z') // Friday 10:00 WIB
    renderRoute(ROUTES.iqamah, { as: takmir })

    expect(await screen.findByText('Jumat 12:09')).toBeInTheDocument()
    expect(await flowSteps('Jumat')).toEqual([
      'Layar Jumat11:09–12:09',
      'Layar adzan12:09–12:10',
      'Khutbah12:10–12:40',
      'Mode sholat12:40–12:50',
    ])

    await userEvent.click(screen.getByRole('button', { name: 'Kurangi Durasi khutbah' }))
    expect(screen.getByRole('group', { name: 'Khutbah, belum disimpan' })).toBeInTheDocument()
    expect((await flowSteps('Jumat'))[2]).toBe('Khutbah12:10–12:35')
  })

  it('links to Petugas Jumat, its own menu', async () => {
    at('2026-10-28T03:00:00Z')
    renderRoute(ROUTES.iqamah, { as: takmir })

    expect(await screen.findByRole('link', { name: /Petugas Jumat/ })).toHaveAttribute(
      'href',
      ROUTES.fridayDuties,
    )
  })

  it('hides the alarm details while it is off, and saves the screen settings', async () => {
    at('2026-10-28T03:00:00Z')
    renderRoute(ROUTES.iqamah, { as: takmir })

    await screen.findByText(/^Dzuhur \d{2}:\d{2}$/)
    expect(screen.getByLabelText('Saat masuk waktu adzan')).toBeInTheDocument()
    await userEvent.click(screen.getByLabelText('Aktif'))
    expect(screen.queryByLabelText('Saat masuk waktu adzan')).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('radio', { name: /Gelap total/ }))
    await userEvent.click(screen.getByRole('button', { name: 'Simpan dan kirim ke layar' }))
    await confirm('Simpan pengaturan iqamah?', 'Simpan')
    await screen.findByText('Pengaturan iqamah disimpan')
    expect(fakeIqamah.settings.alarm.enabled).toBe(false)
    expect(fakeIqamah.settings.sholat_display).toBe('black')
  })

  it('resets the table to the defaults after confirming', async () => {
    at('2026-10-28T03:00:00Z')
    renderRoute(ROUTES.iqamah, { as: takmir })

    await screen.findByText(/^Dzuhur \d{2}:\d{2}$/)
    await userEvent.click(screen.getByRole('button', { name: 'Atur ulang' }))
    await confirm('Atur ulang iqamah & mode sholat?', 'Atur ulang')
    // the fake's seed differs from the defaults: Subuh iqamah 10 → 15
    expect(screen.getByRole('group', { name: 'Iqamah Subuh' })).toHaveTextContent('15 mnt')
    expect(screen.getByRole('button', { name: 'Atur ulang' })).toBeDisabled()
  })
})
