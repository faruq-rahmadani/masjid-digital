import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/route'
import { fakeFridayDuties } from '@/test/fakeApi'
import { bendahara, renderRoute, takmir } from '@/test/render'

// Only Date is faked: Wednesday 2026-10-28 10:00 WIB. The fake seeds petugas for Friday 30 Okt.
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

describe('FridayDutiesPage', () => {
  it('shows the coming Friday on top and flags the empty weeks', async () => {
    at('2026-10-28T03:00:00Z')
    renderRoute(ROUTES.fridayDuties, { as: takmir })

    expect(await screen.findByText('Jumat, 30 Okt')).toBeInTheDocument()
    expect(screen.getByText('Ustadz Abdullah')).toBeInTheDocument()
    expect(screen.getByText('"Menjaga amanah"')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Ubah' })).toBeInTheDocument()
    // the 8 Fridays after it, none filled yet
    expect(screen.getAllByText('Belum diisi')).toHaveLength(8)
    expect(screen.getByText('8 Jumat belum diisi')).toBeInTheDocument()
  })

  it('fills, edits and deletes petugas, each saved right away after confirming', async () => {
    at('2026-10-28T03:00:00Z')
    renderRoute(ROUTES.fridayDuties, { as: takmir })

    await userEvent.click(await screen.findByRole('button', { name: 'Isi petugas Jumat, 6 Nov' }))
    const drawer = screen.getByRole('dialog', { name: 'Petugas Jumat, 6 Nov' })
    await userEvent.click(within(drawer).getByRole('button', { name: 'Simpan petugas' }))
    expect(await within(drawer).findByText('Isi minimal satu petugas')).toBeInTheDocument()

    await userEvent.type(within(drawer).getByLabelText('Khatib'), 'Ustadz Yusuf')
    await userEvent.click(within(drawer).getByRole('button', { name: 'Simpan petugas' }))
    expect(fakeFridayDuties.list).toHaveLength(1) // not before confirming
    await confirm('Simpan petugas Jumat?', 'Simpan')
    expect(await screen.findByText('Khatib Ustadz Yusuf')).toBeInTheDocument()
    expect(screen.getByText('7 Jumat belum diisi')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Ubah petugas Jumat, 6 Nov' }))
    const edit = screen.getByRole('dialog', { name: 'Petugas Jumat, 6 Nov' })
    expect(within(edit).getByLabelText('Khatib')).toHaveValue('Ustadz Yusuf')
    await userEvent.type(within(edit).getByLabelText('Imam'), 'Ustadz Hasan')
    await userEvent.click(within(edit).getByRole('button', { name: 'Simpan petugas' }))
    await confirm('Simpan petugas Jumat?', 'Simpan')
    expect(await screen.findByText('Khatib Ustadz Yusuf · Imam Ustadz Hasan')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Hapus petugas Jumat, 6 Nov' }))
    await confirm('Hapus petugas Jumat?', 'Hapus')
    expect(await screen.findByText('Petugas Jumat dihapus')).toBeInTheDocument()
    expect(fakeFridayDuties.list.map((d) => d.date)).toEqual(['2026-10-30'])
  })

  it('asks to fill the coming Friday when it is empty', async () => {
    fakeFridayDuties.list = []
    at('2026-10-28T03:00:00Z')
    renderRoute(ROUTES.fridayDuties, { as: takmir })

    await userEvent.click(await screen.findByRole('button', { name: 'Isi petugas' }))
    expect(screen.getByRole('dialog', { name: 'Petugas Jumat, 30 Okt' })).toBeInTheDocument()
  })

  it('is not open to the bendahara', async () => {
    renderRoute(ROUTES.fridayDuties, { as: bendahara })

    expect(await screen.findByRole('heading', { name: 'Tidak ada akses' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Petugas Jumat' })).not.toBeInTheDocument()
  })
})
