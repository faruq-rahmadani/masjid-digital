import { EllipsisVertical, Settings, Trash2 } from 'lucide-react'
import { Button, CopyField, IconButton } from '@/components'
import { Group, Row, Specimen } from '@/pages/Kit/Specimen'

export function ActionsSection() {
  return (
    <Group id="actions" description="Things people press. One primary action per screen.">
      <Specimen
        name="Button · variant"
        usage="primary for the main action, secondary for the rest, danger for destructive, income/expense only on the Keuangan entry points."
      >
        <Row>
          <Button size="md">Primary</Button>
          <Button size="md" variant="secondary">
            Secondary
          </Button>
          <Button size="md" variant="outline">
            Outline
          </Button>
          <Button size="md" variant="ghost">
            Ghost
          </Button>
        </Row>
        <Row>
          <Button size="md" variant="danger">
            Hapus
          </Button>
          <Button size="md" variant="gold">
            Beri nama
          </Button>
          <Button size="md" variant="notice">
            Tutup buku
          </Button>
        </Row>
        <div className="grid grid-cols-2 gap-2.5">
          <Button variant="income">+ Pemasukan</Button>
          <Button variant="expense">− Pengeluaran</Button>
        </div>
      </Specimen>

      <Specimen
        name="Button · size"
        usage="xl (52) for the screen's primary action, lg (48) for section actions, md (40) in cards, sm (36) inline in rows."
      >
        <Button size="xl" block>
          Simpan dan kirim ke layar
        </Button>
        <Button size="lg" block variant="secondary">
          Tambah pengguna
        </Button>
        <Row>
          <Button size="md" variant="secondary">
            Reload
          </Button>
          <Button size="sm" variant="secondary">
            Ganti logo
          </Button>
        </Row>
      </Specimen>

      <Specimen name="Button · state" usage="loading keeps the label and blocks double submits.">
        <Row>
          <Button size="md" loading>
            Menyimpan
          </Button>
          <Button size="md" disabled>
            Nonaktif
          </Button>
          <Button size="md" variant="danger">
            <Trash2 aria-hidden /> Dengan ikon
          </Button>
        </Row>
      </Specimen>

      <Specimen
        name="IconButton"
        usage="Icon-only, always with a label. card in headers, brand on the hero, ghost for the ⋯ row menu."
      >
        <div className="grid grid-cols-3 gap-2 text-center text-xs">
          <div className="border-line-soft bg-canvas flex flex-col items-center gap-1.5 rounded-xl border p-3">
            <IconButton label="Pengaturan" icon={<Settings />} />
            card
          </div>
          <div className="bg-brand text-on-brand flex flex-col items-center gap-1.5 rounded-xl p-3">
            <IconButton variant="brand" label="Akun saya" icon={<Settings />} />
            brand
          </div>
          <div className="bg-card flex flex-col items-center gap-1.5 rounded-xl p-3">
            <IconButton variant="ghost" label="Aksi" icon={<EllipsisVertical />} />
            ghost
          </div>
        </div>
      </Specimen>

      <Specimen name="CopyField" usage="Read-only value with Salin (the TV URL)." surface="brand">
        <p className="text-md font-bold">Pasang di TV</p>
        <CopyField tone="brand" value="masjid.app/tv/nur-husen" />
      </Specimen>
    </Group>
  )
}
