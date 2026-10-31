import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/route'
import { fakeCalls } from '@/test/fakeApi'
import { renderRoute, superAdmin } from '@/test/render'

const userList = () => screen.findByRole('list', { name: 'Daftar pengguna' })

async function openAddSheet() {
  await userList()
  await userEvent.click(screen.getByRole('button', { name: 'Tambah pengguna' }))
  return screen.getByRole('dialog', { name: 'Tambah pengguna' })
}

async function openActions(name: string) {
  await userEvent.click(await screen.findByRole('button', { name: `Aksi untuk ${name}` }))
}

describe('UsersPage', () => {
  it('lists you first, without a menu on your own row', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })

    const rows = within(await userList()).getAllByRole('listitem')
    expect(rows[0]).toHaveTextContent(/Ahmad Fauzi \(Anda\)\s*ID 1/)
    expect(within(rows[0]!).queryByRole('button')).not.toBeInTheDocument()
    expect(rows.at(-1)).toHaveTextContent('Nonaktif')
    expect(screen.getByText('4 akun')).toBeInTheDocument()
  })

  it('shows each user id', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })

    const row = (await within(await userList()).findByText(/Hasan Basri/)).closest('li')!
    expect(row).toHaveTextContent('ID 2')
    expect(row).toHaveTextContent('0812-3456-1001')
  })

  it('filters by name and by role', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })
    await userList()

    await userEvent.type(screen.getByRole('searchbox', { name: 'Cari pengguna' }), 'siti')
    expect(within(await userList()).getAllByRole('listitem')).toHaveLength(1)
    expect(screen.getByText('1 dari 4 pengguna')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Hapus pencarian' }))
    await userEvent.click(screen.getByRole('button', { name: 'Takmir' }))
    const rows = within(await userList()).getAllByRole('listitem')
    expect(rows).toHaveLength(2)
    expect(rows[0]).toHaveTextContent('Hasan Basri')
    expect(rows[1]).toHaveTextContent('Umar Hadi')
  })

  it('offers to clear a filter that matches nobody', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })
    await userList()

    await userEvent.type(screen.getByRole('searchbox', { name: 'Cari pengguna' }), 'tidak ada')
    expect(screen.getByText('Tidak ada pengguna yang cocok')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Hapus filter' }))
    expect(within(await userList()).getAllByRole('listitem')).toHaveLength(4)
  })

  it('validates the add form like the API', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })
    await openAddSheet()

    await userEvent.type(screen.getByLabelText('Nomor HP'), '021-555')
    await userEvent.type(screen.getByLabelText('Password awal'), 'pendek')
    await userEvent.click(screen.getByRole('button', { name: 'Simpan pengguna' }))

    expect(screen.getByText('Nama wajib diisi')).toBeInTheDocument()
    expect(screen.getByText(/Nomor HP tidak valid/)).toBeInTheDocument()
    expect(screen.getByText('Password minimal 8 karakter')).toBeInTheDocument()
  })

  it('adds a user from the sheet and shows them in the list', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })
    await openAddSheet()

    await userEvent.type(screen.getByLabelText('Nama'), 'Yusuf Ramli')
    await userEvent.type(screen.getByLabelText('Nomor HP'), '0812 9999 0001')
    await userEvent.selectOptions(screen.getByLabelText('Peran'), 'bendahara')
    await userEvent.type(screen.getByLabelText('Password awal'), 'rahasia-kuat')
    await userEvent.click(screen.getByRole('button', { name: 'Simpan pengguna' }))

    expect(await within(await userList()).findByText('Yusuf Ramli')).toBeInTheDocument()
    expect(await screen.findByText('5 akun')).toBeInTheDocument()
    expect(screen.queryByRole('dialog', { name: 'Tambah pengguna' })).not.toBeInTheDocument()
  })

  it('puts a duplicate phone error on the phone field', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })
    await openAddSheet()

    await userEvent.type(screen.getByLabelText('Nama'), 'Duplikat')
    await userEvent.type(screen.getByLabelText('Nomor HP'), '0812-3456-1001')
    await userEvent.type(screen.getByLabelText('Password awal'), 'rahasia-kuat')
    await userEvent.click(screen.getByRole('button', { name: 'Simpan pengguna' }))

    expect(await screen.findByText('Nomor HP atau email sudah terdaftar')).toBeInTheDocument()
    expect(screen.getByLabelText('Nomor HP')).toHaveAttribute('aria-invalid', 'true')
  })

  it('changes a role from the row menu', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })

    await openActions('Hasan Basri')
    await userEvent.click(screen.getByRole('button', { name: 'Ubah peran' }))
    await userEvent.click(screen.getByRole('radio', { name: /Bendahara/ }))
    await userEvent.click(screen.getByRole('button', { name: 'Simpan peran' }))

    const row = (await screen.findByText('Hasan Basri')).closest('li')!
    expect(await within(row).findByText('Bendahara')).toBeInTheDocument()
  })

  it('deactivates after confirmation and offers to reactivate', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })

    await openActions('Siti Aminah')
    await userEvent.click(screen.getByRole('button', { name: 'Nonaktifkan akun' }))
    await userEvent.click(screen.getByRole('button', { name: 'Nonaktifkan' }))

    const row = (await screen.findByText('Siti Aminah')).closest('li')!
    expect(await within(row).findByText('Nonaktif')).toBeInTheDocument()
    await openActions('Siti Aminah')
    expect(screen.getByRole('button', { name: 'Aktifkan kembali' })).toBeInTheDocument()
  })

  it('resets a password with validation', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })

    await openActions('Hasan Basri')
    await userEvent.click(screen.getByRole('button', { name: 'Atur ulang password' }))
    await userEvent.type(screen.getByLabelText('Password baru'), '123')
    await userEvent.click(screen.getByRole('button', { name: 'Simpan password' }))
    expect(screen.getByText('Password minimal 8 karakter')).toBeInTheDocument()

    await userEvent.clear(screen.getByLabelText('Password baru'))
    await userEvent.type(screen.getByLabelText('Password baru'), 'password-baru-1')
    await userEvent.click(screen.getByRole('button', { name: 'Simpan password' }))

    expect(await screen.findByText('Password Hasan Basri diatur ulang')).toBeInTheDocument()
  })

  it('replaces only the changed row after an update, without refetching the list', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })
    await userList()
    expect(fakeCalls.listUsers).toBe(1)

    await openActions('Hasan Basri')
    await userEvent.click(screen.getByRole('button', { name: 'Ubah peran' }))
    await userEvent.click(screen.getByRole('radio', { name: /Bendahara/ }))
    await userEvent.click(screen.getByRole('button', { name: 'Simpan peran' }))
    const row = (await within(await userList()).findByText(/Hasan Basri/)).closest('li')!
    await within(row).findByText('Bendahara')

    await openActions('Siti Aminah')
    await userEvent.click(screen.getByRole('button', { name: 'Nonaktifkan akun' }))
    await userEvent.click(screen.getByRole('button', { name: 'Nonaktifkan' }))
    const deactivated = (await within(await userList()).findByText(/Siti Aminah/)).closest('li')!
    await within(deactivated).findByText('Nonaktif')

    expect(fakeCalls.listUsers).toBe(1)
  })

  it('refetches the list once after adding a user (the API returns no created user)', async () => {
    renderRoute(ROUTES.users, { as: superAdmin })
    await openAddSheet()

    await userEvent.type(screen.getByLabelText('Nama'), 'Yusuf Ramli')
    await userEvent.type(screen.getByLabelText('Nomor HP'), '0812 9999 0001')
    await userEvent.type(screen.getByLabelText('Password awal'), 'rahasia-kuat')
    await userEvent.click(screen.getByRole('button', { name: 'Simpan pengguna' }))
    await within(await userList()).findByText(/Yusuf Ramli/)

    expect(fakeCalls.listUsers).toBe(2)
  })
})
