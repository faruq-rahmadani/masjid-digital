import {
  Avatar,
  Badge,
  Card,
  Divider,
  InfoTile,
  List,
  ListRow,
  Money,
  SectionHeader,
  Stat,
  StatusDot,
  Table,
  TBody,
  Td,
  TextLink,
  THead,
  Th,
  Thumbnail,
  Tr,
} from '@/components'
import { Group, Row, Specimen } from '@/pages/Kit/Specimen'

const HEADERS = ['Tgl', 'Sbh', 'Dzh', 'Asr', 'Mgr', 'Isy']

const PREVIEW = [
  ['Kam 29', '04:41', '12:09', '15:24', '18:15', '19:25'],
  ['Jum 30', '04:41', '11:58', '15:24', '18:15', '19:25'],
  ['Sab 31', '04:41', '12:08', '15:24', '18:15', '19:25'],
] as const

export function DataDisplaySection() {
  return (
    <Group
      id="data-display"
      description="Read-only content. White cards on cream, rows split by soft lines, numbers tabular."
    >
      <Specimen
        name="Card"
        usage="default for content, brand for the one hero figure, brand-2 on green."
      >
        <Card>Kartu biasa: konten di atas latar krem.</Card>
        <Card tone="brand" padding="xl">
          <Stat label="Saldo kas saat ini" value={<Money amount={92_058_114} />} size="xl" />
          <div className="mt-3.5 grid grid-cols-2 gap-2.5">
            <Card tone="brand-2" padding="none" className="rounded-lg px-3 py-2.5">
              <Stat label="Masuk hari ini" value={<Money amount={1_250_000} />} />
            </Card>
            <Card tone="brand-2" padding="none" className="rounded-lg px-3 py-2.5">
              <Stat label="Keluar hari ini" value={<Money amount={1_126_667} />} />
            </Card>
          </div>
        </Card>
      </Specimen>

      <Specimen
        name="List + ListRow"
        usage="Records with a title, a meta line and a trailing value or action."
      >
        <Card padding="list">
          <List>
            <ListRow
              title="Hamba Allah, renovasi atap"
              description="28 Okt · Infaq transfer"
              trailing={<Money amount={750_000} flow="in" />}
            />
            <ListRow
              title="Token listrik Oktober"
              description={
                <>
                  28 Okt · Listrik · <span className="text-gold-ink font-semibold">Diedit</span>
                </>
              }
              trailing={<Money amount={626_667} flow="out" />}
            />
            <ListRow
              leading={<Thumbnail />}
              title="Kajian Rutin Ahad Pagi"
              description="10 dtk · Semua layar"
            />
            <ListRow
              leading={<Thumbnail muted />}
              title={<span className="text-ink-muted">Poster Idul Adha 1447</span>}
              description="Nonaktif"
            />
          </List>
        </Card>
      </Specimen>

      <Specimen name="List · activity" usage="Feed rows: actor in bold, action, time on the right.">
        <SectionHeader
          title="Aktivitas terakhir"
          action={<TextLink to="/kit">Lihat semua</TextLink>}
        />
        <Card padding="list">
          <List>
            {[
              ['[Bendahara]', 'menambah pemasukan Rp 750.000', '16:02'],
              ['[Takmir]', 'mengubah iqamah Maghrib jadi 7 mnt', '14:20'],
              ['[Ketua DKM]', 'menambah pengguna Bendahara', '25 Okt'],
            ].map(([actor, action, time]) => (
              <li key={action} className="flex gap-2.5 py-2.5 text-xs">
                <span className="flex-1">
                  <b>{actor}</b> {action}
                </span>
                <span className="text-ink-muted whitespace-nowrap">{time}</span>
              </li>
            ))}
          </List>
        </Card>
      </Specimen>

      <Specimen
        name="Table"
        usage="Real tabular data with column headers. Use a List instead when each row is one record with a title."
        surface="card"
      >
        <SectionHeader
          title="Pratinjau 7 hari"
          action={<TextLink to="/kit">Lihat semua</TextLink>}
        />
        <Table density="compact" caption="Jadwal sholat 3 hari ke depan">
          <THead>
            <Tr>
              {HEADERS.map((h) => (
                <Th key={h}>{h}</Th>
              ))}
            </Tr>
          </THead>
          <TBody>
            {PREVIEW.map(([day, ...times]) => (
              <Tr key={day}>
                <Th scope="row">{day}</Th>
                {times.map((t, i) => (
                  <Td key={HEADERS[i + 1]}>{t}</Td>
                ))}
              </Tr>
            ))}
          </TBody>
        </Table>
      </Specimen>

      <Specimen name="Badge" usage="Short status or role. Color follows meaning, never decoration.">
        <Row>
          <Badge tone="success" size="md">
            Online
          </Badge>
          <Badge tone="danger" size="md">
            Offline
          </Badge>
          <Badge tone="success">Takmir</Badge>
          <Badge tone="gold">Bendahara</Badge>
          <Badge>Super admin</Badge>
          <Badge tone="gold" caps>
            Baru
          </Badge>
          <Badge tone="gold" caps>
            Penting
          </Badge>
          <Badge tone="notice">Langsung tersimpan</Badge>
        </Row>
      </Specimen>

      <Specimen name="Avatar" usage="Initials; tone follows the role.">
        <Row>
          <Avatar name="[Ketua DKM]" />
          <Avatar name="[Nama Takmir]" tone="success" />
          <Avatar name="[Nama Bendahara]" tone="gold" />
          <Avatar name="[Takmir lama]" tone="neutral" />
          <Avatar name="[Ketua DKM]" size="lg" />
        </Row>
      </Specimen>

      <Specimen name="Stat" usage="A labelled figure, mostly on brand surfaces." surface="brand">
        <Row className="justify-between">
          <Stat label="Sholat berikutnya" value="Maghrib 18:15" size="lg" />
          <Stat label="dalam" value="00:47" size="lg" align="end" valueClassName="text-gold" />
        </Row>
      </Specimen>

      <Specimen
        name="Money"
        usage="Integer Rupiah. Signed and colored when it is a cash flow."
        surface="card"
      >
        <Row className="gap-4">
          <Money amount={92_058_114} className="font-bold" />
          <Money amount={750_000} flow="in" />
          <Money amount={626_667} flow="out" />
          <Money amount={1_250_000} flow="in" withSymbol />
        </Row>
      </Specimen>

      <Specimen name="InfoTile" usage="Read-only label/value pairs, two per row." surface="card">
        <div className="grid grid-cols-2 gap-2.5">
          <InfoTile label="Koordinat" value="-6.2088, 106.8456" />
          <InfoTile label="Zona waktu" value="WIB (UTC+7)" />
        </div>
      </Specimen>

      <Specimen
        name="StatusDot · Thumbnail · Divider"
        usage="Small supporting pieces."
        surface="card"
      >
        <Row className="gap-4">
          <StatusDot tone="success" label="Online" />
          <StatusDot tone="danger" label="Offline" />
          <StatusDot tone="warning" label="Perlu perhatian" />
          <StatusDot tone="neutral" label="Tidak diketahui" />
          <Thumbnail />
          <Thumbnail shape="square" />
        </Row>
        <Divider label="atau" />
        <Divider />
      </Specimen>
    </Group>
  )
}
