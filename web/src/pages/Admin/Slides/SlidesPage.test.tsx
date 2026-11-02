import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useEffect } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/route'
import { fakeSlide } from '@/test/fakeApi'
import { bendahara, renderRoute, takmir } from '@/test/render'

// jsdom can't lay out the cropper or draw on a canvas: the fake cropper reports a fixed area
// once mounted, and cropImage returns a JPEG named after the picked file.
const CROPPED_AREA = { x: 120, y: 40, width: 1312, height: 582 }
vi.mock('react-easy-crop', () => ({
  default: ({
    image,
    onCropComplete,
  }: {
    image: string
    onCropComplete: (area: unknown, pixels: typeof CROPPED_AREA) => void
  }) => {
    // biome-ignore lint/correctness/useExhaustiveDependencies: report once, like the real one after measuring
    useEffect(() => onCropComplete(CROPPED_AREA, CROPPED_AREA), [])
    return <img src={image} alt="" data-testid="cropper" />
  },
}))
const cropImage = vi.hoisted(() =>
  vi.fn(async (file: File, _area: unknown) => {
    const name = `${file.name.replace(/\.[^.]*$/, '')}.jpg`
    return new File([new Uint8Array(10)], name, { type: 'image/jpeg' })
  }),
)
vi.mock('./cropImage', () => ({ cropImage }))

const list = () => screen.findByRole('list')
const titles = () => fakeSlide.slides.map((s) => s.title)
const rowTitles = async () =>
  within(await list())
    .getAllByRole('listitem')
    .map(
      (row) =>
        within(row).getByRole('button', { name: /^(?!Urutkan|Matikan|Tampilkan)/ }).textContent,
    )
const showFilter = async (name: RegExp) =>
  userEvent.click(
    within(screen.getByRole('group', { name: 'Tampilkan' })).getByRole('button', { name }),
  )
const image = (name: string, type = 'image/png', size = 1024) =>
  new File([new Uint8Array(size)], name, { type })

describe('SlidesPage', () => {
  it('opens on the active slides with their duration and image', async () => {
    renderRoute(ROUTES.slides, { as: takmir })

    expect(await rowTitles()).toEqual([
      'Kajian Rutin Ahad Pagi10 dtk · Sedang tampil',
      'Infaq Renovasi Atap15 dtk · Sedang tampil',
      'Foto Kegiatan Santri TPA10 dtk · Sedang tampil',
    ])
    expect(screen.getByText('3 aktif dari 4')).toBeInTheDocument()
    // decorative (alt=""): the title is right next to it
    expect((await list()).querySelector('img')).toHaveAttribute('src', '/fake/slides/1/image')

    await showFilter(/^Nonaktif/)
    expect(await rowTitles()).toEqual(['Poster Idul Adha 144710 dtk · Nonaktif'])
  })

  it('uploads an image only after its details are saved', async () => {
    renderRoute(ROUTES.slides, { as: takmir })
    await list()

    await userEvent.upload(
      screen.getByLabelText(/Upload gambar/),
      image('Jadwal Imam Jumat.jpg', 'image/jpeg'),
    )
    const sheet = screen.getByRole('dialog', { name: 'Slide baru' })
    expect(within(sheet).getByTestId('cropper')).toHaveAttribute('src', 'blob:preview')
    expect(within(sheet).getByLabelText('Judul')).toHaveValue('Jadwal Imam Jumat')
    expect(fakeSlide.slides).toHaveLength(4) // nothing uploaded yet

    await userEvent.type(within(sheet).getByLabelText('Keterangan (opsional)'), 'Bulan November')
    await userEvent.click(within(sheet).getByLabelText('Durasi'))
    await userEvent.click(screen.getByRole('option', { name: '20 detik' }))
    await userEvent.click(within(sheet).getByRole('button', { name: 'Simpan' }))

    expect(await screen.findByText('Slide ditambahkan')).toBeInTheDocument()
    expect(screen.getByText('Tampil di TV dalam ±5 menit')).toBeInTheDocument()
    expect(screen.queryByRole('dialog', { name: 'Slide baru' })).not.toBeInTheDocument()
    // only the framed part is uploaded
    expect(cropImage).toHaveBeenLastCalledWith(expect.any(File), CROPPED_AREA)
    expect(fakeSlide.slides.at(-1)).toMatchObject({
      title: 'Jadwal Imam Jumat',
      caption: 'Bulan November',
      duration_seconds: 20,
      active: true,
    })
    expect((await rowTitles()).at(-1)).toBe('Jadwal Imam Jumat20 dtk · Sedang tampil')
    expect(screen.getByText('4 aktif dari 5')).toBeInTheDocument()
  })

  it('previews the slide as the TV shows it, with or without its title', async () => {
    renderRoute(ROUTES.slides, { as: takmir })
    await list()

    await userEvent.upload(screen.getByLabelText(/Upload gambar/), image('Poster Qurban.png'))
    const sheet = screen.getByRole('dialog', { name: 'Slide baru' })
    await userEvent.type(
      within(sheet).getByLabelText('Keterangan (opsional)'),
      'Daftar di sekretariat',
    )
    // the preview's caption bar follows the form
    expect(within(sheet).getByText('Poster Qurban', { selector: 'div' })).toBeInTheDocument()
    expect(
      within(sheet).getByText('Daftar di sekretariat', { selector: 'div' }),
    ).toBeInTheDocument()

    await userEvent.click(within(sheet).getByLabelText('Tampilkan judul & keterangan di TV'))
    expect(within(sheet).queryByText('Poster Qurban', { selector: 'div' })).not.toBeInTheDocument()
    expect(within(sheet).getByLabelText('Keterangan (opsional)')).toBeDisabled()
    await userEvent.click(within(sheet).getByRole('button', { name: 'Simpan' }))

    expect(await screen.findByText('Slide ditambahkan')).toBeInTheDocument()
    expect(fakeSlide.slides.at(-1)).toMatchObject({ title: 'Poster Qurban', show_caption: false })
    // the list's thumbnail shows the caption bar only for slides that keep it
    const rows = within(await list()).getAllByRole('listitem')
    expect(rows[0]?.querySelector('[data-caption]')).not.toBeNull()
    expect(rows.at(-1)?.querySelector('[data-caption]')).toBeNull()
  })

  it('uploads nothing when the new slide is cancelled', async () => {
    renderRoute(ROUTES.slides, { as: takmir })
    await list()

    await userEvent.upload(screen.getByLabelText(/Upload gambar/), image('Kajian.png'))
    const sheet = screen.getByRole('dialog', { name: 'Slide baru' })
    const title = within(sheet).getByLabelText('Judul')
    await userEvent.clear(title)
    expect(within(sheet).getByRole('button', { name: 'Simpan' })).toBeDisabled()
    await userEvent.click(within(sheet).getByRole('button', { name: 'Batal' }))

    expect(screen.queryByRole('dialog', { name: 'Slide baru' })).not.toBeInTheDocument()
    expect(fakeSlide.slides).toHaveLength(4)
  })

  it('rejects a file that is too large or not an image before opening the sheet', async () => {
    const user = userEvent.setup({ applyAccept: false })
    renderRoute(ROUTES.slides, { as: takmir })
    await list()

    await user.upload(
      screen.getByLabelText(/Upload gambar/),
      image('besar.png', 'image/png', 20 * 1024 * 1024 + 1),
    )
    expect(await screen.findByText('Ukuran gambar maksimal 20 MB')).toBeInTheDocument()
    await user.upload(screen.getByLabelText(/Upload gambar/), image('animasi.gif', 'image/gif'))
    expect(await screen.findByText('Format gambar harus PNG, JPG, atau WebP')).toBeInTheDocument()

    expect(screen.queryByRole('dialog', { name: 'Slide baru' })).not.toBeInTheDocument()
    expect(fakeSlide.slides).toHaveLength(4)
  })

  it('turns a slide off, keeping the row until the filter changes', async () => {
    renderRoute(ROUTES.slides, { as: takmir })
    await list()

    await userEvent.click(screen.getByRole('button', { name: 'Matikan: Infaq Renovasi Atap' }))
    expect(
      screen.getByRole('button', { name: 'Tampilkan: Infaq Renovasi Atap' }),
    ).toBeInTheDocument()
    await waitFor(() => expect(fakeSlide.slides[1]?.active).toBe(false))
    // the rest is sent as it was
    expect(fakeSlide.slides[1]).toMatchObject({ duration_seconds: 15, caption: expect.any(String) })
    expect(screen.getByText('2 aktif dari 4')).toBeInTheDocument()

    await showFilter(/^Semua/)
    await showFilter(/^Aktif/)
    expect(await rowTitles()).not.toContain('Infaq Renovasi Atap15 dtk · Nonaktif')
  })

  it('edits the title, caption and duration in the sheet', async () => {
    renderRoute(ROUTES.slides, { as: takmir })
    await list()

    await userEvent.click(screen.getByRole('button', { name: /^Infaq Renovasi Atap/ }))
    const sheet = screen.getByRole('dialog', { name: 'Edit slide' })
    expect(within(sheet).getByRole('button', { name: 'Simpan' })).toBeDisabled()
    const title = within(sheet).getByLabelText('Judul')
    await userEvent.clear(title)
    await userEvent.type(title, 'Infaq Atap Masjid')
    await userEvent.clear(within(sheet).getByLabelText('Keterangan (opsional)'))
    await userEvent.click(within(sheet).getByLabelText('Durasi'))
    await userEvent.click(screen.getByRole('option', { name: '30 detik' }))
    await userEvent.click(within(sheet).getByRole('button', { name: 'Simpan' }))

    expect(await screen.findByText('Slide disimpan')).toBeInTheDocument()
    expect(screen.queryByRole('dialog', { name: 'Edit slide' })).not.toBeInTheDocument()
    expect(fakeSlide.slides[1]).toMatchObject({
      title: 'Infaq Atap Masjid',
      caption: '',
      duration_seconds: 30,
      show_caption: true,
      active: true,
    })
    expect(await rowTitles()).toContain('Infaq Atap Masjid30 dtk · Sedang tampil')
  })

  it('deletes a slide after confirming', async () => {
    renderRoute(ROUTES.slides, { as: takmir })
    await list()

    await userEvent.click(screen.getByRole('button', { name: /^Foto Kegiatan/ }))
    await userEvent.click(
      within(screen.getByRole('dialog', { name: 'Edit slide' })).getByRole('button', {
        name: 'Hapus',
      }),
    )
    const confirm = screen.getByRole('dialog', { name: 'Hapus slide?' })
    expect(titles()).toContain('Foto Kegiatan Santri TPA') // not before confirming
    await userEvent.click(within(confirm).getByRole('button', { name: 'Hapus' }))

    expect(await screen.findByText('Slide dihapus')).toBeInTheDocument()
    expect(titles()).not.toContain('Foto Kegiatan Santri TPA')
  })

  it('reorders the filtered slides without moving the hidden ones', async () => {
    renderRoute(ROUTES.slides, { as: takmir })

    // jsdom has no layout: give each row a box so dnd-kit can find the one above
    const rows = within(await list()).getAllByRole('listitem')
    rows.forEach((row, i) => {
      row.getBoundingClientRect = () => new DOMRect(0, i * 50, 300, 50)
    })
    screen.getByRole('button', { name: 'Urutkan: Foto Kegiatan Santri TPA' }).focus()
    await userEvent.keyboard(' ')
    await userEvent.keyboard('{ArrowUp}')
    await userEvent.keyboard(' ')

    await waitFor(() =>
      expect(titles()).toEqual([
        'Kajian Rutin Ahad Pagi',
        'Foto Kegiatan Santri TPA',
        'Infaq Renovasi Atap',
        'Poster Idul Adha 1447',
      ]),
    )
  })

  it('shows an empty state with no slides', async () => {
    fakeSlide.slides = []
    renderRoute(ROUTES.slides, { as: takmir })

    expect(await screen.findByText('Belum ada slide')).toBeInTheDocument()
    expect(screen.getByText('0 aktif dari 0')).toBeInTheDocument()
  })

  it('is closed to bendahara', async () => {
    renderRoute(ROUTES.slides, { as: bendahara })

    expect(await screen.findByRole('heading', { level: 1 })).not.toHaveTextContent('Slide')
    expect(screen.queryByLabelText(/Upload gambar/)).not.toBeInTheDocument()
  })
})
