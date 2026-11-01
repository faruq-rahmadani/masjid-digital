import { useMutation } from '@tanstack/react-query'
import { LocateFixed, MapPin } from 'lucide-react'
import { lazy, Suspense, useState } from 'react'
import {
  Button,
  Card,
  InfoTile,
  List,
  SearchInput,
  SegmentedControl,
  Skeleton,
  useToast,
} from '@/components'
import { searchPlaces } from '@/services/geocodingService'
import type { MosqueTimezone } from '@/types/mosque'
import { TIMEZONES } from '@/utils/timezone'
import { type Coordinates, formatCoordinates, roundCoordinates } from '../utils'

const LocationMap = lazy(async () => ({ default: (await import('./LocationMap')).LocationMap }))

type LocationCardProps = {
  location: Coordinates | null
  /** A new point from search, GPS, a tap or a dragged pin. */
  onLocationChange: (value: Coordinates) => void
  locationError?: string
  timezone: MosqueTimezone
  onTimezoneChange: (value: MosqueTimezone) => void
}

const TIMEZONE_OPTIONS = TIMEZONES.map((tz) => ({ value: tz.value, label: tz.label }))

export function LocationCard({
  location,
  onLocationChange,
  locationError,
  timezone,
  onTimezoneChange,
}: LocationCardProps) {
  const toast = useToast()
  const [query, setQuery] = useState('')
  const search = useMutation({ mutationFn: searchPlaces })
  const [locating, setLocating] = useState(false)
  const offset = TIMEZONES.find((tz) => tz.value === timezone)?.offset

  const runSearch = () => {
    if (query.trim() !== '') search.mutate(query.trim())
  }

  const pick = (value: Coordinates) => {
    search.reset()
    onLocationChange(roundCoordinates(value))
  }

  const locate = () => {
    setLocating(true)
    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        setLocating(false)
        pick({ latitude: coords.latitude, longitude: coords.longitude })
      },
      (error) => {
        setLocating(false)
        toast({
          tone: 'danger',
          message:
            error.code === error.PERMISSION_DENIED
              ? 'Izin lokasi ditolak'
              : 'Lokasi Anda tidak bisa ditentukan',
          description:
            error.code === error.PERMISSION_DENIED
              ? 'Aktifkan izin lokasi di pengaturan browser, atau cari alamat masjid.'
              : 'Coba lagi di luar ruangan, atau cari alamat masjid.',
        })
      },
      { enableHighAccuracy: true, timeout: 15_000 },
    )
  }

  return (
    <Card className="flex flex-col gap-3">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-base font-bold">Lokasi</h2>
        {'geolocation' in navigator && (
          <Button size="sm" variant="ghost" loading={locating} onClick={locate}>
            {!locating && <LocateFixed aria-hidden className="size-4" />} Pakai lokasi saya
          </Button>
        )}
      </div>

      <SearchInput
        label="Cari lokasi"
        placeholder="Cari alamat atau nama tempat"
        enterKeyHint="search"
        value={query}
        onChange={(event) => setQuery(event.currentTarget.value)}
        onClear={() => {
          setQuery('')
          search.reset()
        }}
        onKeyDown={(event) => {
          // Enter searches; without this it would submit the profile form around it.
          if (event.key !== 'Enter') return
          event.preventDefault()
          runSearch()
        }}
      />

      {search.isPending && <p className="text-ink-muted text-xs">Mencari lokasi…</p>}
      {search.isError && (
        <p className="text-expense text-xs font-semibold">{search.error.message}</p>
      )}
      {search.data?.length === 0 && (
        <p className="text-ink-muted text-xs">Tidak ditemukan. Coba nama jalan atau kelurahan.</p>
      )}
      {search.data && search.data.length > 0 && (
        <List aria-label="Hasil pencarian lokasi" className="border-line rounded-lg border">
          {search.data.map((place) => (
            <li key={`${place.latitude},${place.longitude}`}>
              <button
                type="button"
                onClick={() => pick(place)}
                className="min-h-touch flex w-full items-start gap-2 px-3 py-2.5 text-left text-sm"
              >
                <MapPin aria-hidden className="text-ink-muted mt-0.5 size-4 shrink-0" />
                <span className="line-clamp-2">{place.name}</span>
              </button>
            </li>
          ))}
        </List>
      )}

      <Suspense fallback={<Skeleton className="h-40 rounded-lg" />}>
        <LocationMap value={location} onChange={onLocationChange} />
      </Suspense>
      <p
        className={locationError ? 'text-expense text-xs font-semibold' : 'text-ink-muted text-xs'}
      >
        {locationError ??
          (location
            ? 'Geser pin atau ketuk peta untuk menyesuaikan titik masjid.'
            : 'Cari alamat, pakai lokasi Anda, atau ketuk peta untuk menaruh pin.')}
      </p>

      <InfoTile label="Koordinat" value={location ? formatCoordinates(location) : 'Belum diatur'} />

      <div className="flex flex-col gap-1.5">
        <p className="text-ink-soft text-xs font-semibold">Zona waktu</p>
        <SegmentedControl
          label="Zona waktu"
          options={TIMEZONE_OPTIONS}
          value={timezone}
          onChange={onTimezoneChange}
        />
        <p className="text-ink-muted text-xs">
          {offset}. Diisi otomatis dari titik lokasi, ubah jika keliru.
        </p>
      </div>
    </Card>
  )
}
