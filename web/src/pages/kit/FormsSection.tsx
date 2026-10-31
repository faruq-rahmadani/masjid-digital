import { useState } from 'react'
import {
  Checkbox,
  Field,
  Input,
  MoneyInput,
  RadioCard,
  SearchInput,
  SegmentedControl,
  Select,
  Slider,
  Stepper,
  Textarea,
  UploadTile,
} from '@/components'
import { Group, Row, Specimen } from './Specimen'

export function FormsSection() {
  const [flow, setFlow] = useState<'in' | 'out'>('in')
  const [filter, setFilter] = useState<'all' | 'in' | 'out'>('all')
  const [hijri, setHijri] = useState('0')
  const [amount, setAmount] = useState<number | null>(750_000)
  const [correction, setCorrection] = useState(2)
  const [adzan, setAdzan] = useState(4)
  const [display, setDisplay] = useState('dim')

  return (
    <Group
      id="forms"
      description="Every control sits in a Field (label, hint, error) or carries its own label. 44px tall by default."
    >
      <Specimen
        name="Field + Input / Select / Textarea"
        usage="States: default, hint, error, disabled."
        surface="card"
      >
        <Field label="Nama masjid" hint="Tampil di header layar TV">
          <Input defaultValue="Masjid Nur Husen" />
        </Field>
        <Field label="Email atau nomor HP" error="Format email atau nomor HP tidak dikenali">
          <Input defaultValue="bendahara@" />
        </Field>
        <Field label="Kategori">
          <Select defaultValue="Infaq transfer">
            <option>Infaq transfer</option>
            <option>Kotak amal harian</option>
          </Select>
        </Field>
        <Field label="Tanggal">
          <Input type="date" defaultValue="2026-10-28" />
        </Field>
        <Field label="Zona waktu">
          <Input value="WIB (UTC+7)" disabled readOnly />
        </Field>
        <Field label="Teks baru">
          <Textarea
            defaultValue="Kerja bakti bersih masjid Ahad 2 Nov, kumpul pukul 07.00"
            maxLength={200}
            showCount
          />
        </Field>
      </Specimen>

      <Specimen name="Input · lg" usage="48px, login only." surface="card">
        <Field label="Password">
          <Input size="lg" type="password" defaultValue="rahasia123" />
        </Field>
      </Specimen>

      <Specimen
        name="MoneyInput"
        usage="Amount entry: digits only, grouped, selects all on focus. Colored by flow."
      >
        <SegmentedControl
          label="Jenis transaksi"
          variant="track"
          value={flow}
          onChange={setFlow}
          tone={(v) => (v === 'in' ? 'income' : 'expense')}
          options={[
            { value: 'in', label: 'Pemasukan' },
            { value: 'out', label: 'Pengeluaran' },
          ]}
        />
        <MoneyInput value={amount} onChange={setAmount} flow={flow} />
      </Specimen>

      <Specimen name="SearchInput" usage="Filters a list in place.">
        <SearchInput label="Cari transaksi" placeholder="Cari keterangan transaksi" />
      </Specimen>

      <Specimen
        name="SegmentedControl"
        usage="One value from 2–5 options. track for a binary type, split for a scale, pill for list filters."
      >
        <SegmentedControl
          label="Koreksi tanggal Hijriah"
          value={hijri}
          onChange={setHijri}
          options={['-2', '-1', '0', '+1', '+2'].map((v) => ({ value: v, label: v }))}
        />
        <Row>
          <SegmentedControl
            label="Filter transaksi"
            variant="pill"
            value={filter}
            onChange={setFilter}
            options={[
              { value: 'all', label: 'Semua' },
              { value: 'in', label: 'Masuk' },
              { value: 'out', label: 'Keluar' },
            ]}
          />
          <Select size="pill" aria-label="Periode" className="ml-auto">
            <option>Okt 2026</option>
          </Select>
        </Row>
      </Specimen>

      <Specimen
        name="Stepper"
        usage="Small bounded numbers. sm in tables, md standalone."
        surface="card"
      >
        <Row className="justify-between">
          <span className="font-semibold">Koreksi Subuh</span>
          <Stepper
            label="koreksi Subuh"
            min={-10}
            max={10}
            value={correction}
            onChange={setCorrection}
            format={(v) => (v > 0 ? `+${v}` : String(v))}
          />
        </Row>
        <Row className="justify-between">
          <span className="font-semibold">Durasi adzan</span>
          <Stepper
            size="md"
            label="durasi adzan"
            min={1}
            max={10}
            value={adzan}
            onChange={setAdzan}
            format={(v) => `${v} mnt`}
          />
        </Row>
      </Specimen>

      <Specimen name="Checkbox" usage="On/off settings. The mockup uses checkboxes, not switches.">
        <Row className="gap-4">
          <Checkbox label="Ingat saya" defaultChecked />
          <Checkbox label="Tandai penting" />
          <Checkbox label="Aktif" size="lg" defaultChecked />
        </Row>
      </Specimen>

      <Specimen name="RadioCard" usage="One of a few choices that need a description.">
        <fieldset className="m-0 flex min-w-0 flex-col gap-2 border-0 p-0">
          <legend className="mb-2 font-bold">Tampilan saat sholat</legend>
          <RadioCard
            name="kit-display"
            checked={display === 'dim'}
            onChange={() => setDisplay('dim')}
            title="Teks redup"
            description="“Sholat sedang berlangsung” + jam kecil"
          />
          <RadioCard
            name="kit-display"
            checked={display === 'black'}
            onChange={() => setDisplay('black')}
            title="Gelap total"
            description="Layar hitam penuh"
          />
        </fieldset>
      </Specimen>

      <Specimen name="Slider" usage="Continuous value such as volume." surface="card">
        <Field label="Volume">
          <Slider min={0} max={100} defaultValue={70} minLabel="Pelan" maxLabel="Keras" />
        </Field>
      </Specimen>

      <Specimen name="UploadTile" usage="dropzone for content uploads, tile to add one more file.">
        <UploadTile
          label="Upload gambar"
          hint="Bisa pilih banyak sekaligus · JPG/PNG/WebP, maks 5 MB · disarankan 16:9"
          accept="image/jpeg,image/png,image/webp"
          multiple
          onFiles={() => {}}
        />
        <UploadTile variant="tile" label="Tambah foto bukti" accept="image/*" onFiles={() => {}} />
      </Specimen>
    </Group>
  )
}
