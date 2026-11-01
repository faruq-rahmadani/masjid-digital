import { Card, TextLink } from '@/components'
import { ROUTES } from '@/route'
import type { MosqueProfile } from '@/types/mosque'
import { timezoneLabel } from '@/utils/timezone'

// Where the times come from: the method is fixed (PRD), the location is the mosque profile's.
// Brand-colored like Home's header, since it is context from elsewhere, not a setting of this page.
export function ContextCard({ location }: { location: MosqueProfile | null }) {
  return (
    <Card tone="brand" className="flex flex-col gap-3">
      <div>
        <p className="text-on-brand-muted text-xs font-semibold">Metode perhitungan</p>
        <p className="text-sm font-semibold">Kemenag RI (Subuh 20°, Isya 18°)</p>
      </div>
      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="text-on-brand-muted text-xs font-semibold">Lokasi (dari profil masjid)</p>
          <p className="truncate text-sm font-semibold tabular-nums">
            {location
              ? `${location.latitude.toFixed(4)}, ${location.longitude.toFixed(4)} · ${timezoneLabel(location.timezone)}`
              : 'Belum diisi'}
          </p>
        </div>
        <TextLink
          to={ROUTES.profile}
          size="sm"
          className="text-on-brand decoration-on-brand-muted hover:decoration-on-brand shrink-0"
        >
          {location ? 'Ubah' : 'Isi lokasi'}
        </TextLink>
      </div>
    </Card>
  )
}
