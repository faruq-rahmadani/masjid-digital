import {
  Banner,
  Button,
  Callout,
  Card,
  EmptyState,
  ProgressBar,
  Skeleton,
  Spinner,
  useToast,
} from '@/components'
import { Group, Row, Specimen } from './Specimen'

export function FeedbackSection() {
  const toast = useToast()

  return (
    <Group
      id="feedback"
      description="Tell people what happened and what to do next. Every list has empty, loading and error states."
    >
      <Specimen
        name="Callout"
        usage="notice for consequences before acting, danger for a problem that needs attention."
      >
        <Callout>
          Saldo setelah disimpan: <b>Rp 92.808.114</b>. Mengedit atau menghapus transaksi nanti
          wajib mengisi alasan.
        </Callout>
        <Callout
          title="Tutup buku September"
          action={
            <Button size="md" variant="notice">
              Tutup buku
            </Button>
          }
        >
          Transaksi September akan terkunci
        </Callout>
        <Callout tone="danger">
          Layar tetap menampilkan jadwal dari cache, tapi konten baru belum masuk. Cek internet atau
          nyalakan ulang TV.
        </Callout>
      </Specimen>

      <Specimen
        name="Toast · tone"
        usage="Passing feedback after an action, above the action bar. success and info 4s, warning 6s, danger 8s; hover or focus pauses. Max 3 at once."
      >
        <div className="grid grid-cols-2 gap-2.5">
          <Button
            variant="secondary"
            onClick={() => toast({ message: 'Tersimpan dan dikirim ke layar' })}
          >
            Success
          </Button>
          <Button
            variant="secondary"
            onClick={() =>
              toast({
                tone: 'info',
                message: 'Layar TV memuat ulang',
                description: 'TV Ruang Utama akan tampil lagi dalam beberapa detik.',
              })
            }
          >
            Info
          </Button>
          <Button
            variant="secondary"
            onClick={() =>
              toast({
                tone: 'warning',
                message: 'Tersimpan, 1 layar belum menerima',
                description: 'TV Teras offline. Konten masuk saat terhubung lagi.',
              })
            }
          >
            Warning
          </Button>
          <Button
            variant="secondary"
            onClick={() =>
              toast({
                tone: 'danger',
                message: 'Gagal menyimpan',
                description: 'Periksa koneksi internet lalu coba lagi.',
              })
            }
          >
            Danger
          </Button>
        </div>
      </Specimen>

      <Specimen
        name="Toast · action"
        usage="One short action, usually Batalkan (undo) after a reversible change. Pass duration: null to keep it until closed."
      >
        <div className="grid grid-cols-2 gap-2.5">
          <Button
            variant="secondary"
            onClick={() =>
              toast({
                message: 'Slide dinonaktifkan',
                description: 'Poster Idul Adha 1447',
                action: {
                  label: 'Batalkan',
                  onClick: () => toast({ message: 'Slide diaktifkan lagi' }),
                },
              })
            }
          >
            Dengan undo
          </Button>
          <Button
            variant="secondary"
            onClick={() =>
              toast({
                tone: 'info',
                message: 'Versi baru tersedia',
                action: { label: 'Muat ulang', onClick: () => {} },
                duration: null,
              })
            }
          >
            Tetap tampil
          </Button>
        </div>
      </Specimen>

      <Specimen
        name="Banner"
        usage="Screen-wide status in a full-width strip under the PageHeader. Stays while the condition lasts; dismissible only when it is just information."
      >
        <div className="-mx-5 flex flex-col gap-2">
          <Banner tone="info" title="Mode lihat saja">
            Peran Takmir hanya bisa melihat keuangan.
          </Banner>
          <Banner
            tone="warning"
            title="Anda sedang offline"
            action={<button type="button">Coba lagi</button>}
          >
            Perubahan belum bisa disimpan sampai internet tersambung.
          </Banner>
          <Banner tone="warning">
            Oktober 2026 sudah tutup buku. Transaksi di bulan ini tidak bisa diubah.
          </Banner>
          <Banner tone="success" onDismiss={() => {}}>
            Semua layar menerima jadwal terbaru.
          </Banner>
          <Banner tone="danger" title="2 layar offline lebih dari 1 jam">
            Jadwal tetap tampil dari cache, tapi konten baru belum masuk.
          </Banner>
        </div>
      </Specimen>

      <Specimen
        name="Which one?"
        usage="Pick by scope and lifetime."
        surface="card"
        className="gap-0 py-1"
      >
        {[
          ['Toast', 'One action just finished', 'Disappears by itself'],
          ['Banner', 'The whole screen or app', 'While the condition lasts'],
          ['Callout', 'One card or form', 'Part of the content'],
        ].map(([name, scope, lifetime]) => (
          <div
            key={name}
            className="grid grid-cols-[4.5rem_1fr] gap-x-3 border-b border-line-soft py-2.5 text-sm last:border-0"
          >
            <span className="font-mono font-semibold">{name}</span>
            <span>
              {scope}
              <span className="block text-xs text-ink-muted">{lifetime}</span>
            </span>
          </div>
        ))}
      </Specimen>

      <Specimen
        name="EmptyState"
        usage="Say what will appear here and offer the action that fills it."
      >
        <EmptyState
          title="Belum ada transaksi"
          description="Transaksi bulan ini akan muncul di sini."
          action={<Button size="md">+ Pemasukan</Button>}
        />
      </Specimen>

      <Specimen
        name="Skeleton · Spinner"
        usage="Skeleton when the layout is known, Spinner for short inline waits."
      >
        <Card className="flex flex-col gap-2.5">
          <Skeleton className="h-4 w-2/3" />
          <Skeleton className="h-3 w-1/3" />
          <Skeleton className="h-4 w-1/2" />
        </Card>
        <Row>
          <Spinner /> <span className="text-sm text-ink-muted">Memuat transaksi…</span>
        </Row>
      </Specimen>

      <Specimen
        name="ProgressBar"
        usage="Known progress: iqamah countdown, program targets."
        surface="card"
      >
        <ProgressBar label="Sisa waktu iqamah" value={89} />
        <ProgressBar label="Target renovasi atap" value={42} tone="income" />
      </Specimen>
    </Group>
  )
}
