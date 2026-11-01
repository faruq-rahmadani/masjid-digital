import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/route'
import { fakeRunningText } from '@/test/fakeApi'
import { bendahara, renderRoute, takmir } from '@/test/render'

const list = () => screen.findByRole('list')
const texts = () => fakeRunningText.texts.map((t) => t.text)
const rowTexts = async () =>
  within(await list())
    .getAllByRole('listitem')
    .map(
      (row) =>
        within(row).getByRole('button', { name: /^(?!Urutkan|Matikan|Tampilkan)/ }).firstChild
          ?.textContent,
    )
const showFilter = async (name: RegExp) =>
  userEvent.click(
    within(screen.getByRole('group', { name: 'Tampilkan' })).getByRole('button', { name }),
  )

describe('RunningTextsPage', () => {
  it('opens on the active texts, with who made them', async () => {
    renderRoute(ROUTES.runningTexts, { as: takmir })

    expect(await rowTexts()).toEqual(texts().slice(0, 3))
    expect(screen.getByRole('button', { name: /^Matikan: Kajian Tafsir/ })).toBeInTheDocument()
    expect(screen.getAllByText('Ahmad Fauzi · Sedang tampil')).toHaveLength(3)
    const filter = screen.getByRole('group', { name: 'Tampilkan' })
    expect(within(filter).getByRole('button', { name: 'Aktif (3)' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
    expect(within(filter).getByRole('button', { name: 'Nonaktif (1)' })).toBeInTheDocument()

    await showFilter(/^Nonaktif/)
    expect(await rowTexts()).toEqual(['Pendaftaran qurban dibuka sampai 20 Mei'])
    expect(screen.getByText(/^Ahmad Fauzi · Terakhir tampil \d+ hari lalu$/)).toBeInTheDocument()

    await showFilter(/^Semua/)
    expect(await rowTexts()).toEqual(texts())
  })

  it('adds a text at the bottom', async () => {
    renderRoute(ROUTES.runningTexts, { as: takmir })
    await list()

    const add = screen.getByRole('button', { name: 'Tambah' })
    expect(add).toBeDisabled()
    await userEvent.type(screen.getByLabelText('Teks baru'), '  Kerja bakti Ahad pagi ')
    expect(screen.getByText('24 / 200')).toBeInTheDocument()
    await userEvent.click(add)

    expect(await screen.findByText('Running text ditambahkan')).toBeInTheDocument()
    expect(screen.getByText('Tampil di TV dalam ±5 menit')).toBeInTheDocument()
    expect(texts().at(-1)).toBe('Kerja bakti Ahad pagi')
    expect((await rowTexts()).at(-1)).toBe('Kerja bakti Ahad pagi')
    expect(screen.getAllByText('Hasan Basri · Sedang tampil')).toHaveLength(1)
    expect(screen.getByLabelText('Teks baru')).toHaveValue('')
  })

  it('turns a text on and off, keeping the row until the filter changes', async () => {
    renderRoute(ROUTES.runningTexts, { as: takmir })
    await list()

    await userEvent.click(screen.getByRole('button', { name: /^Matikan: Kajian Tafsir/ }))
    // still listed under Aktif, now offering to put it back
    expect(screen.getByRole('button', { name: /^Tampilkan: Kajian Tafsir/ })).toBeInTheDocument()
    await waitFor(() => expect(fakeRunningText.texts[1]?.deactivated_at).not.toBeNull())
    expect(await screen.findByText(/Terakhir tampil baru saja/)).toBeInTheDocument()

    await showFilter(/^Nonaktif/)
    await userEvent.click(screen.getByRole('button', { name: /^Tampilkan: Pendaftaran qurban/ }))
    expect(screen.getByRole('button', { name: /^Matikan: Pendaftaran qurban/ })).toBeInTheDocument()
    await waitFor(() => expect(fakeRunningText.texts[3]?.active).toBe(true))
    expect(fakeRunningText.texts[3]?.deactivated_at).toBeNull()

    await showFilter(/^Aktif/)
    expect(await rowTexts()).not.toContain("Kajian Tafsir setiap Kamis ba'da Maghrib")
  })

  it('edits a text in the sheet', async () => {
    renderRoute(ROUTES.runningTexts, { as: takmir })
    await list()

    await userEvent.click(screen.getByRole('button', { name: /^Kajian Tafsir/ }))
    const sheet = screen.getByRole('dialog', { name: 'Edit running text' })
    const field = within(sheet).getByLabelText('Teks')
    await userEvent.clear(field)
    await userEvent.type(field, 'Kajian Fiqih setiap Selasa')
    await userEvent.click(within(sheet).getByRole('button', { name: 'Simpan' }))

    expect(await screen.findByText('Running text disimpan')).toBeInTheDocument()
    expect(screen.queryByRole('dialog', { name: 'Edit running text' })).not.toBeInTheDocument()
    expect(texts()[1]).toBe('Kajian Fiqih setiap Selasa')
    expect(await rowTexts()).toContain('Kajian Fiqih setiap Selasa')
  })

  it('deletes a text after confirming', async () => {
    renderRoute(ROUTES.runningTexts, { as: takmir })
    await list()

    await userEvent.click(screen.getByRole('button', { name: /^Infaq via QRIS/ }))
    await userEvent.click(
      within(screen.getByRole('dialog', { name: 'Edit running text' })).getByRole('button', {
        name: 'Hapus',
      }),
    )
    const confirm = screen.getByRole('dialog', { name: 'Hapus running text?' })
    expect(texts()).toContain('Infaq via QRIS tersedia di dekat pintu utama') // not before confirming
    await userEvent.click(within(confirm).getByRole('button', { name: 'Hapus' }))

    expect(await screen.findByText('Running text dihapus')).toBeInTheDocument()
    expect(texts()).not.toContain('Infaq via QRIS tersedia di dekat pintu utama')
    expect(screen.queryByRole('button', { name: /^Infaq via QRIS/ })).not.toBeInTheDocument()
  })

  it('reorders the filtered texts without moving the hidden ones', async () => {
    renderRoute(ROUTES.runningTexts, { as: takmir }) // opens on Aktif

    const rows = within(await list()).getAllByRole('listitem')
    rows.forEach((row, i) => {
      row.getBoundingClientRect = () => new DOMRect(0, i * 50, 300, 50)
    })
    // Infaq (3rd active) up one, past Kajian Tafsir; Pendaftaran qurban (hidden) stays last
    screen.getByRole('button', { name: /^Urutkan: Infaq/ }).focus()
    await userEvent.keyboard(' ')
    await userEvent.keyboard('{ArrowUp}')
    await userEvent.keyboard(' ')

    await waitFor(() =>
      expect(texts()).toEqual([
        'Mohon matikan nada dering HP selama di dalam masjid',
        'Infaq via QRIS tersedia di dekat pintu utama',
        "Kajian Tafsir setiap Kamis ba'da Maghrib",
        'Pendaftaran qurban dibuka sampai 20 Mei',
      ]),
    )
  })

  it('reorders with the keyboard', async () => {
    renderRoute(ROUTES.runningTexts, { as: takmir })
    await list()

    // jsdom has no layout: give each row a box so dnd-kit can find the one below
    const rows = within(await list()).getAllByRole('listitem')
    rows.forEach((row, i) => {
      row.getBoundingClientRect = () => new DOMRect(0, i * 50, 300, 50)
    })

    screen.getByRole('button', { name: /^Urutkan: Mohon matikan/ }).focus()
    await userEvent.keyboard(' ')
    await userEvent.keyboard('{ArrowDown}')
    await userEvent.keyboard(' ')

    await waitFor(() =>
      expect(texts().slice(0, 2)).toEqual([
        "Kajian Tafsir setiap Kamis ba'da Maghrib",
        'Mohon matikan nada dering HP selama di dalam masjid',
      ]),
    )
  })

  it('saves the scroll speed and previews the active texts', async () => {
    renderRoute(ROUTES.runningTexts, { as: takmir })
    await list()

    const speed = screen.getByRole('group', { name: 'Kecepatan gulir' })
    expect(within(speed).getByRole('button', { name: 'Sedang' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
    await userEvent.click(within(speed).getByRole('button', { name: 'Cepat' }))

    expect(within(speed).getByRole('button', { name: 'Cepat' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
    await waitFor(() => expect(fakeRunningText.setting.speed).toBe('fast'))
    expect(screen.getByText('INFO')).toBeInTheDocument()
  })

  it('shows an empty state with no texts', async () => {
    fakeRunningText.texts = []
    renderRoute(ROUTES.runningTexts, { as: takmir })

    expect(await screen.findByText('Belum ada running text')).toBeInTheDocument()
    expect(screen.getByText(/Tidak ada teks aktif/)).toBeInTheDocument()
  })

  it('is closed to bendahara', async () => {
    renderRoute(ROUTES.runningTexts, { as: bendahara })

    expect(await screen.findByRole('heading', { level: 1 })).not.toHaveTextContent('Running Text')
    expect(screen.queryByLabelText('Teks baru')).not.toBeInTheDocument()
  })
})
