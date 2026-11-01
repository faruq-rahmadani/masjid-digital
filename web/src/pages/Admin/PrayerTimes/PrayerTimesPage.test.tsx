import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/route'
import { fakeMosque, fakePrayer } from '@/test/fakeApi'
import { renderRoute, takmir } from '@/test/render'
import { minutesOf } from '@/utils/prayer/calendar'

// Only Date is faked: Monday 2026-11-02 08:00 WIB, so "today" and the 7-day preview are fixed.
beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date('2026-11-02T01:00:00Z'))
})

afterEach(() => {
  vi.useRealTimers()
})

const todayTable = () => screen.findByRole('table', { name: 'Koreksi waktu dan jam hari ini' })
// Confirm the bottom-sheet question that every save, delete and leave asks
const confirm = async (title: string, button: string) => {
  const dialog = await screen.findByRole('dialog', { name: title })
  await userEvent.click(within(dialog).getByRole('button', { name: button }))
}

const timeOf = (table: HTMLElement, name: string) =>
  within(within(table).getByRole('rowheader', { name }).closest('tr')!).getAllByRole('cell')[0]!
    .textContent

describe('PrayerTimesPage', () => {
  it('previews a correction live, then saves it', async () => {
    renderRoute(ROUTES.prayerTimes, { as: takmir })

    const table = await todayTable()
    const before = timeOf(table, 'Maghrib')!
    expect(before).toMatch(/^1[78]:\d{2}$/)
    const save = screen.getByRole('button', { name: 'Simpan dan kirim ke layar' })
    expect(save).toBeDisabled()

    await userEvent.click(screen.getByRole('button', { name: 'Tambah Koreksi Maghrib' }))
    await userEvent.click(screen.getByRole('button', { name: 'Tambah Koreksi Maghrib' }))
    expect(minutesOf(timeOf(table, 'Maghrib')!)).toBe(minutesOf(before) + 2)
    expect(screen.getByText('Ada perubahan belum disimpan')).toBeInTheDocument()
    expect(screen.getByRole('row', { name: 'Maghrib, belum disimpan' })).toHaveClass('bg-highlight')
    expect(screen.getByRole('rowheader', { name: 'Isya' }).closest('tr')).not.toHaveClass(
      'bg-highlight',
    )

    await userEvent.click(save)
    expect(fakePrayer.settings.corrections.maghrib).toBe(0) // not before confirming
    await confirm('Simpan pengaturan waktu sholat?', 'Simpan')
    expect(await screen.findByText('Waktu sholat disimpan')).toBeInTheDocument()
    expect(screen.queryByText('Ada perubahan belum disimpan')).not.toBeInTheDocument()
    expect(fakePrayer.settings.corrections.maghrib).toBe(2)
    // saved: no longer marked as changed
    expect(screen.queryByRole('row', { name: 'Maghrib, belum disimpan' })).not.toBeInTheDocument()
  })

  it('adds a weekly rule that shows in the preview', async () => {
    renderRoute(ROUTES.prayerTimes, { as: takmir })
    await todayTable()

    await userEvent.click(
      within(
        screen.getByRole('heading', { name: 'Jadwal mingguan' }).closest('div.rounded-xl')!,
      ).getByRole('button', { name: 'Tambah' }),
    )
    const drawer = screen.getByRole('dialog', { name: 'Tambah jadwal mingguan' })
    await userEvent.click(within(drawer).getAllByRole('combobox')[0]!)
    await userEvent.click(screen.getByRole('option', { name: 'Senin' }))
    await userEvent.type(within(drawer).getByLabelText('Jam'), '12:30')
    await userEvent.click(within(drawer).getByRole('button', { name: 'Tambahkan' }))

    expect(screen.getByText('Setiap Senin: Dzuhur 12:30')).toBeInTheDocument()
    // today is Monday
    expect(timeOf(await todayTable(), 'Dzuhur')).toContain('12:30')
    const preview = screen.getByRole('table', { name: 'Pratinjau waktu sholat 7 hari' })
    expect(within(preview).getByText('12:30').closest('td')).toHaveClass('font-bold') // replaced

    await userEvent.click(screen.getByRole('button', { name: 'Simpan dan kirim ke layar' }))
    await confirm('Simpan pengaturan waktu sholat?', 'Simpan')
    await screen.findByText('Waktu sholat disimpan')
    expect(fakePrayer.settings.weekly_overrides).toContainEqual({
      weekday: 'monday',
      prayer: 'dzuhur',
      time: '12:30',
      remarks: '',
    })
  })

  it('adds a date override right away, rejects a duplicate, and deletes it', async () => {
    renderRoute(ROUTES.prayerTimes, { as: takmir })
    await todayTable()
    const card = () =>
      screen
        .getByRole('heading', { name: 'Jadwal khusus tanggal tertentu' })
        .closest('div.rounded-xl')!

    const add = async (time: string) => {
      await userEvent.click(within(card() as HTMLElement).getByRole('button', { name: 'Tambah' }))
      const drawer = screen.getByRole('dialog', { name: 'Tambah jadwal khusus' })
      const date = within(drawer).getByLabelText('Tanggal')
      await userEvent.clear(date)
      await userEvent.type(date, '2026-11-06')
      await userEvent.type(within(drawer).getByLabelText('Jam'), time)
      await userEvent.type(within(drawer).getByLabelText('Keterangan (opsional)'), 'Khatib tamu')
      await userEvent.click(within(drawer).getByRole('button', { name: 'Simpan jadwal khusus' }))
      await confirm('Simpan jadwal khusus?', 'Simpan')
    }

    await add('11:58')
    expect(
      await within(card() as HTMLElement).findByText('Jumat, 6 Nov: Dzuhur 11:58'),
    ).toBeInTheDocument()
    expect(fakePrayer.overrides).toHaveLength(2)

    await add('12:05')
    expect(
      await screen.findByText(/sudah punya jadwal khusus di tanggal tersebut/i),
    ).toBeInTheDocument()
    await userEvent.keyboard('{Escape}')

    await userEvent.click(
      screen.getByRole('button', { name: 'Hapus jadwal khusus Dzuhur Jumat, 6 Nov' }),
    )
    await confirm('Hapus jadwal khusus?', 'Hapus')
    expect(await screen.findByText('Jadwal khusus dihapus')).toBeInTheDocument()
    expect(
      within(card() as HTMLElement).queryByText('Jumat, 6 Nov: Dzuhur 11:58'),
    ).not.toBeInTheDocument()
    expect(fakePrayer.overrides).toHaveLength(1)
  })

  it('resets the corrections after confirming', async () => {
    renderRoute(ROUTES.prayerTimes, { as: takmir })
    const table = await todayTable()
    const reset = screen.getByRole('button', { name: 'Atur ulang' })
    expect(reset).toBeDisabled() // already the defaults
    const before = timeOf(table, 'Isya')!

    await userEvent.click(screen.getByRole('button', { name: 'Tambah Koreksi Isya' }))
    await userEvent.click(screen.getByRole('button', { name: 'Tambah Imsak sebelum Subuh' }))
    expect(reset).toBeEnabled()

    await userEvent.click(reset)
    await confirm('Atur ulang koreksi waktu?', 'Atur ulang')
    expect(timeOf(table, 'Isya')).toBe(before)
    expect(screen.getByRole('group', { name: 'Imsak sebelum Subuh' })).toHaveTextContent('10 mnt')
    expect(reset).toBeDisabled()
    // back to what is saved, so nothing left to save
    expect(screen.queryByText('Ada perubahan belum disimpan')).not.toBeInTheDocument()
  })

  it('removes a weekly rule only after confirming', async () => {
    renderRoute(ROUTES.prayerTimes, { as: takmir })
    await todayTable()
    const remove = () => screen.getByRole('button', { name: 'Hapus jadwal Dzuhur setiap Jumat' })

    await userEvent.click(remove())
    await userEvent.click(
      within(await screen.findByRole('dialog', { name: 'Hapus jadwal mingguan?' })).getByRole(
        'button',
        { name: 'Batal' },
      ),
    )
    expect(screen.getByText('Setiap Jumat: Dzuhur 12:09')).toBeInTheDocument()

    await userEvent.click(remove())
    await confirm('Hapus jadwal mingguan?', 'Hapus')
    expect(screen.queryByText('Setiap Jumat: Dzuhur 12:09')).not.toBeInTheDocument()
    expect(fakePrayer.settings.weekly_overrides).toHaveLength(1) // a draft until saved
  })

  it('asks before leaving with unsaved changes', async () => {
    renderRoute(ROUTES.prayerTimes, { as: takmir })
    await todayTable()

    await userEvent.click(screen.getByRole('button', { name: 'Tambah Koreksi Subuh' }))
    await userEvent.click(screen.getByRole('button', { name: 'Kembali' }))
    await confirm('Tinggalkan halaman?', 'Batal')
    expect(screen.getByRole('heading', { name: 'Waktu Sholat' })).toBeInTheDocument()
    expect(screen.getByText('Ada perubahan belum disimpan')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Kembali' }))
    await confirm('Tinggalkan halaman?', 'Tinggalkan')
    expect(await screen.findByText(/Assalamu'alaikum/)).toBeInTheDocument()
  })

  it('leaves without asking when nothing changed', async () => {
    renderRoute(ROUTES.prayerTimes, { as: takmir })
    await todayTable()

    await userEvent.click(screen.getByRole('button', { name: 'Kembali' }))
    expect(await screen.findByText(/Assalamu'alaikum/)).toBeInTheDocument()
    expect(screen.queryByRole('dialog', { name: 'Tinggalkan halaman?' })).not.toBeInTheDocument()
  })

  it('highlights today in the 7-day preview', async () => {
    renderRoute(ROUTES.prayerTimes, { as: takmir })
    const preview = await screen.findByRole('table', { name: 'Pratinjau waktu sholat 7 hari' })

    const rows = within(preview).getAllByRole('row').slice(1)
    expect(rows).toHaveLength(7)
    expect(rows[0]).toHaveAttribute('aria-current', 'date')
    expect(rows[0]).toHaveTextContent('Hari ini')
    expect(rows[0]).toHaveClass('bg-brand')
    expect(rows[1]).not.toHaveAttribute('aria-current')
  })

  it('shows the Hijri date with its correction', async () => {
    renderRoute(ROUTES.prayerTimes, { as: takmir })
    await todayTable()

    expect(screen.getByText('22 Jumadilawal 1448 H')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: '-1' }))
    expect(screen.getByText('21 Jumadilawal 1448 H')).toBeInTheDocument()
  })

  it('asks for the location before computing anything', async () => {
    fakeMosque.profile = null
    renderRoute(ROUTES.prayerTimes, { as: takmir })

    expect(await screen.findByText('Lokasi masjid belum diisi')).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
    expect(screen.getByText('Belum diisi')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Isi lokasi' })).toHaveAttribute('href', ROUTES.profile)
  })
})
