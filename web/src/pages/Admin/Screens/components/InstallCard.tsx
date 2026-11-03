import { Card, CopyField } from '@/components'
import { QrCode } from './QrCode'

// "Pasang di TV": the address to open on a TV, as a QR code and a copyable link.
export function InstallCard({ url }: { url: string }) {
  return (
    <Card tone="brand" padding="xl" className="flex flex-col gap-3.5">
      <h2 className="text-md font-bold">Pasang di TV</h2>
      <div className="flex items-center gap-3.5">
        <div className="bg-canvas text-brand flex size-24 shrink-0 items-center justify-center rounded-lg p-2">
          <QrCode value={url} label="QR code alamat layar TV" className="size-20" />
        </div>
        <p className="text-on-brand-muted text-sm leading-relaxed">
          Buka alamat ini di browser TV, atau scan QR dari HP lalu kirim ke TV. Layar langsung
          muncul di daftar bawah.
        </p>
      </div>
      <CopyField value={url} tone="brand" />
      <p className="text-on-brand-muted text-xs">
        Tip: tambahkan <b className="text-on-brand">?layar=teras</b> di akhir alamat supaya layar
        langsung bernama.
      </p>
    </Card>
  )
}
