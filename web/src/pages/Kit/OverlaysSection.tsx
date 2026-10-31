import { useState } from 'react'
import {
  ActionMenu,
  Button,
  Card,
  ConfirmDialog,
  Field,
  Input,
  List,
  ListRow,
  Sheet,
  Textarea,
  useToast,
} from '@/components'
import { Group, Specimen } from '@/pages/Kit/Specimen'

export function OverlaysSection() {
  const toast = useToast()
  const [sheet, setSheet] = useState(false)
  const [confirm, setConfirm] = useState(false)
  const [reason, setReason] = useState(false)

  return (
    <Group
      id="overlays"
      description="Layers above the page. Bottom sheets keep actions in thumb reach; Escape and the backdrop close them."
    >
      <Specimen name="Sheet" usage="Short forms and details without leaving the screen.">
        <Button variant="secondary" onClick={() => setSheet(true)}>
          Edit slide
        </Button>
        <Sheet
          open={sheet}
          onClose={() => setSheet(false)}
          title="Edit: Infaq Renovasi Atap"
          footer={
            <div className="grid grid-cols-2 gap-2.5">
              <Button variant="danger">Hapus</Button>
              <Button onClick={() => setSheet(false)}>Simpan</Button>
            </div>
          }
        >
          <div className="flex flex-col gap-3">
            <Field label="Judul (tampil di caption)">
              <Input defaultValue="Infaq Renovasi Atap" />
            </Field>
            <Field label="Keterangan (opsional)">
              <Textarea />
            </Field>
          </div>
        </Sheet>
      </Specimen>

      <Specimen
        name="ActionMenu"
        usage="The ⋯ at the end of a row. Destructive actions last, in rust."
      >
        <Card padding="list">
          <List>
            <ListRow
              title="[Nama Takmir]"
              description="0812-xxxx-1001"
              trailing={
                <ActionMenu
                  label="[Nama Takmir]"
                  actions={[
                    { label: 'Ubah peran', onSelect: () => toast({ message: 'Peran diubah' }) },
                    {
                      label: 'Kirim link reset password',
                      onSelect: () => toast({ message: 'Link reset terkirim' }),
                    },
                    {
                      label: 'Nonaktifkan akun',
                      danger: true,
                      onSelect: () => toast({ tone: 'info', message: 'Akun dinonaktifkan' }),
                    },
                  ]}
                />
              }
            />
          </List>
        </Card>
      </Specimen>

      <Specimen
        name="ConfirmDialog"
        usage="Confirm before anything that locks or deletes. requireReason when the PRD needs an audit reason."
      >
        <div className="grid grid-cols-2 gap-2.5">
          <Button variant="notice" onClick={() => setConfirm(true)}>
            Tutup buku
          </Button>
          <Button variant="danger" onClick={() => setReason(true)}>
            Hapus transaksi
          </Button>
        </div>
        <ConfirmDialog
          open={confirm}
          onClose={() => setConfirm(false)}
          title="Tutup buku September?"
          description="Semua transaksi September akan terkunci dan tidak bisa diubah."
          confirmLabel="Tutup buku"
          onConfirm={() => toast({ message: 'Buku September ditutup' })}
        />
        <ConfirmDialog
          open={reason}
          onClose={() => setReason(false)}
          title="Hapus transaksi?"
          description="Token listrik Oktober · − Rp 626.667"
          confirmLabel="Hapus"
          danger
          requireReason
          onConfirm={() => toast({ message: 'Transaksi dihapus' })}
        />
      </Specimen>
    </Group>
  )
}
