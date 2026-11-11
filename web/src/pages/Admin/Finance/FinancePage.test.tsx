import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ROUTES } from '@/route'
import { FAKE_TODAY, fakeFinance } from '@/test/fakeFinance'
import { bendahara, renderRoute, takmir } from '@/test/render'
import type { Transaction } from '@/types/finance'

// The fake seeds Tunai and BSI, the four built-in kas and Renovasi Atap (terikat, target
// 3.000.000), and opens the books with Tunai 2.000.000 in Kas Umum and BSI 1.500.000 in Zakat Mal.

const sheet = (title: string) => screen.findByRole('dialog', { name: title })

const confirm = async (title: string, button: string) => {
  const dialog = await screen.findByRole('dialog', { name: title })
  await userEvent.click(within(dialog).getByRole('button', { name: button }))
}

const choose = async (scope: HTMLElement, label: string, option: string, index = 0) => {
  await userEvent.click(
    within(scope).getAllByRole('combobox', { name: label })[index] as HTMLElement,
  )
  await userEvent.click(await screen.findByRole('option', { name: option }))
}

const seedTx = (over: Partial<Transaction>) => {
  const tx: Transaction = {
    id: fakeFinance.state.nextId++,
    kind: 'in',
    amount: 0,
    date: FAKE_TODAY,
    account_id: 1,
    pocket_id: 1,
    category_id: 10,
    group_id: null,
    donor_consent: false,
    corrects_id: null,
    note: '',
    donor_note: '',
    has_photo: false,
    created_by: 3,
    updated_by: 3,
    created_at: '2026-11-11T01:00:00Z',
    updated_at: '2026-11-11T01:00:00Z',
    deleted_at: null,
    deleted_by: 0,
    ...over,
  }
  fakeFinance.state.txs.push(tx)
  return tx
}

describe('FinancePage', () => {
  it('asks to start pembukuan before anything is recorded', async () => {
    fakeFinance.notStarted()
    renderRoute(ROUTES.finance, { as: bendahara })

    expect(await screen.findByText('Pembukuan belum dimulai')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Mulai pembukuan/ })).toHaveAttribute(
      'href',
      ROUTES.accounts,
    )
  })

  it('records a pemasukan after confirming and updates the balance', async () => {
    renderRoute(ROUTES.finance, { as: bendahara })

    expect(await screen.findByTestId('balance')).toHaveTextContent(/3\.500\.000/)
    await userEvent.click(screen.getByRole('button', { name: 'Pemasukan' }))
    const form = await sheet('Tambah transaksi')
    await userEvent.type(within(form).getByLabelText('Nominal'), '750000')
    await choose(form, 'Kategori', 'Infaq Jumat')
    await userEvent.click(within(form).getByRole('button', { name: 'Simpan transaksi' }))
    expect(fakeFinance.state.txs).toHaveLength(0) // not before confirming
    await confirm('Simpan transaksi?', 'Simpan')

    expect(await screen.findByText('Transaksi disimpan')).toBeInTheDocument()
    expect(fakeFinance.state.txs).toMatchObject([
      { kind: 'in', amount: 750_000, pocket_id: 1, category_id: 10, account_id: 1 },
    ])
    expect(await screen.findByTestId('balance')).toHaveTextContent(/4\.250\.000/)
  })

  it('shows why an expense above the kas balance is refused', async () => {
    renderRoute(ROUTES.finance, { as: bendahara })

    await userEvent.click(await screen.findByRole('button', { name: 'Pengeluaran' }))
    const form = await sheet('Tambah transaksi')
    await userEvent.type(within(form).getByLabelText('Nominal'), '5000000')
    await choose(form, 'Kategori', 'Listrik')
    await userEvent.click(within(form).getByRole('button', { name: 'Simpan transaksi' }))
    await confirm('Simpan transaksi?', 'Simpan')

    expect(await within(form).findByText('Saldo kas Kas Umum tidak cukup')).toBeInTheDocument()
    expect(fakeFinance.state.txs).toHaveLength(0)
  })

  it('splits one receipt over two kas as one group', async () => {
    renderRoute(ROUTES.finance, { as: bendahara })

    await userEvent.click(await screen.findByRole('button', { name: 'Pemasukan' }))
    const form = await sheet('Tambah transaksi')
    await userEvent.click(within(form).getByLabelText('Bagi ke beberapa kas'))
    await choose(form, 'Kas', 'Renovasi Atap', 0)
    await choose(form, 'Kategori', 'Infaq Jumat', 0)
    await choose(form, 'Kategori', 'Zakat mal', 1)
    const [first, second] = within(form).getAllByLabelText('Nominal')
    await userEvent.type(first as HTMLElement, '400000')
    await userEvent.type(second as HTMLElement, '600000')
    expect(within(form).getByText(/Total diterima/)).toHaveTextContent(/1\.000\.000/)
    await userEvent.click(within(form).getByRole('button', { name: 'Simpan transaksi' }))
    await confirm('Simpan transaksi?', 'Simpan')

    await screen.findByText('Transaksi disimpan')
    const [a, b] = fakeFinance.state.txs
    expect([a?.pocket_id, b?.pocket_id]).toEqual([5, 2])
    expect(a?.group_id).toBeTruthy()
    expect(a?.group_id).toBe(b?.group_id)
  })

  it('moves money out of a terikat kas only with the donors consent', async () => {
    seedTx({ amount: 400_000, pocket_id: 5 })
    renderRoute(ROUTES.finance, { as: bendahara })

    await userEvent.click(await screen.findByRole('button', { name: 'Pindah dana' }))
    const form = await sheet('Pindah dana')
    await userEvent.type(within(form).getByLabelText('Nominal'), '100000')
    await choose(form, 'Dari kas', 'Renovasi Atap')
    await choose(form, 'Ke kas', 'Kas Umum')
    await userEvent.click(within(form).getByRole('button', { name: 'Pindahkan dana' }))
    expect(
      await within(form).findByText('Isi alasan dan centang persetujuan donatur.'),
    ).toBeInTheDocument()

    await userEvent.type(within(form).getByLabelText(/Alasan/), 'Sisa dana genteng, rapat 10 Nov')
    await userEvent.click(within(form).getByLabelText('Donatur sudah merelakan dana ini dialihkan'))
    await userEvent.click(within(form).getByRole('button', { name: 'Pindahkan dana' }))
    await confirm('Pindahkan dana?', 'Pindahkan')

    await screen.findByText('Dana dipindahkan')
    expect(fakeFinance.state.txs.slice(1)).toMatchObject([
      { kind: 'out', pocket_id: 5, category_id: 20, donor_consent: true },
      { kind: 'in', pocket_id: 1, category_id: 21 },
    ])
  })

  it('offers a zakat kas only setor/tarik tunai within itself', async () => {
    renderRoute(ROUTES.finance, { as: bendahara })

    await userEvent.click(await screen.findByRole('button', { name: 'Pindah dana' }))
    const form = await sheet('Pindah dana')
    await choose(form, 'Dari kas', 'Zakat Mal')
    await userEvent.click(within(form).getByRole('combobox', { name: 'Ke kas' }))

    expect(screen.getAllByRole('option').map((o) => o.textContent)).toEqual([
      'Zakat Mal (setor/tarik tunai)',
    ])
  })

  it('deletes a split as a whole, only with a reason', async () => {
    seedTx({ amount: 400_000, pocket_id: 5, group_id: 'g1', note: 'Transfer QRIS', account_id: 2 })
    seedTx({
      amount: 600_000,
      pocket_id: 2,
      category_id: 12,
      group_id: 'g1',
      note: 'Transfer QRIS',
      account_id: 2,
    })
    renderRoute(ROUTES.finance, { as: bendahara })

    await userEvent.click((await screen.findAllByText('Transfer QRIS'))[0] as HTMLElement)
    const detail = await sheet('Transfer QRIS')
    expect(within(detail).getByText('Dibagi ke beberapa kas')).toBeInTheDocument()
    await userEvent.click(within(detail).getByRole('button', { name: 'Hapus' }))
    const dialog = await screen.findByRole('dialog', { name: 'Hapus transaksi?' })
    expect(within(dialog).getByRole('button', { name: 'Hapus' })).toBeDisabled()
    await userEvent.type(within(dialog).getByRole('textbox'), 'Salah input')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Hapus' }))

    await screen.findByText('Transaksi dihapus')
    expect(fakeFinance.state.txs.every((t) => t.deleted_at !== null)).toBe(true)
    expect(fakeFinance.state.revisions.map((r) => r.reason)).toEqual(['Salah input', 'Salah input'])
  })

  it('lets takmir read Keuangan without any way to change it', async () => {
    seedTx({ amount: 250_000, note: 'Kotak amal Jumat' })
    renderRoute(ROUTES.finance, { as: takmir })

    expect(await screen.findByText('Kotak amal Jumat')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Pemasukan' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Pindah dana' })).not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /Kategori/ })).not.toBeInTheDocument()
  })
})
