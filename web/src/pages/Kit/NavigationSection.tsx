import { AlignLeft, Bell, Clock, CreditCard, House, Image, Monitor, Users } from 'lucide-react'
import { useState } from 'react'
import {
  Button,
  Card,
  Input,
  List,
  ListRow,
  MenuTile,
  NavCard,
  StatusDot,
  Tabs,
  TextLink,
} from '@/components'
import { Group, Row, Specimen } from '@/pages/Kit/Specimen'

export function NavigationSection() {
  const [tab, setTab] = useState<'in' | 'out'>('in')

  return (
    <Group
      id="navigation"
      description="Home is the hub; every other screen has a back button. No tab bar. Menus show only what the role may open."
    >
      <Specimen
        name="PageHeader"
        usage="Top of every screen except Home: back or ×, title, optional meta and actions. See the top of this page."
      >
        <p className="text-ink-muted text-xs">Rendered at the top of /kit.</p>
      </Specimen>

      <Specimen name="MenuTile" usage="Home “Kelola” grid, 4 per row.">
        <nav aria-label="Kelola" className="grid grid-cols-4 gap-2">
          <MenuTile to="/kit" icon={<Image />} label="Slide" />
          <MenuTile to="/kit" icon={<AlignLeft />} label="Running text" />
          <MenuTile to="/kit" icon={<CreditCard />} label="Keuangan" />
          <MenuTile to="/kit" icon={<Clock />} label="Waktu sholat" />
          <MenuTile to="/kit" icon={<Bell />} label="Iqamah" />
          <MenuTile to="/kit" icon={<House />} label="Profil masjid" />
          <MenuTile to="/kit" icon={<Users />} label="Pengguna" />
          <MenuTile to="/kit" icon={<Monitor />} label="Layar TV" />
        </nav>
      </Specimen>

      <Specimen name="NavCard" usage="A whole card that opens a screen, with a status summary.">
        <NavCard
          to="/kit"
          title="Layar TV"
          description="1 online · 1 offline sejak 2 jam lalu"
          trailing={
            <span className="flex gap-1.5">
              <StatusDot tone="success" />
              <StatusDot tone="danger" />
            </span>
          }
        />
      </Specimen>

      <Specimen
        name="Tabs"
        usage="Switch between views of content. For a form value, use SegmentedControl."
      >
        <Tabs
          label="Jenis kategori"
          value={tab}
          onChange={setTab}
          items={[
            {
              value: 'in',
              label: 'Pemasukan',
              content: (
                <Card padding="list">
                  <List>
                    <ListRow title="Kotak amal harian" description="86 transaksi" />
                    <ListRow title="Infaq Jumat" description="24 transaksi" />
                    <ListRow title="Program Ramadhan 1448" description="Belum ada transaksi" />
                  </List>
                </Card>
              ),
            },
            {
              value: 'out',
              label: 'Pengeluaran',
              content: (
                <Card padding="list">
                  <List>
                    <ListRow title="Listrik" description="12 transaksi" />
                    <ListRow title="Kegiatan" description="31 transaksi" />
                  </List>
                </Card>
              ),
            },
          ]}
        />
        <Row>
          <Input
            placeholder="Kategori pemasukan baru"
            aria-label="Kategori baru"
            className="flex-1"
          />
          <Button size="md" className="h-11">
            Tambah
          </Button>
        </Row>
      </Specimen>

      <Specimen
        name="TextLink"
        usage="Inline links: “Lihat semua”, “Ubah”, “Kelola kategori”."
        surface="card"
      >
        <Row className="gap-4">
          <TextLink to="/kit">Lihat semua</TextLink>
          <TextLink to="/kit" size="sm">
            Ubah
          </TextLink>
          <TextLink to="/kit">Kelola kategori</TextLink>
        </Row>
      </Specimen>
    </Group>
  )
}
