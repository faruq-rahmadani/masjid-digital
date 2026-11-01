// Place search for the Profil masjid map, via OpenStreetMap's Nominatim (no key, max 1 request/s:
// https://operations.osmfoundation.org/policies/nominatim/). Called on submit only, never per keystroke.
// It is a third-party host, so it uses fetch instead of the API client (no session token).

export type Place = {
  /** Full name from Nominatim, e.g. "Masjid Istiqlal, Jalan Taman Wijaya Kusuma, ...". */
  name: string
  latitude: number
  longitude: number
}

type NominatimResult = { display_name: string; lat: string; lon: string }

const NOMINATIM_URL = 'https://nominatim.openstreetmap.org/search'

export async function searchPlaces(query: string): Promise<Place[]> {
  const params = new URLSearchParams({
    q: query,
    format: 'jsonv2',
    countrycodes: 'id',
    'accept-language': 'id',
    limit: '5',
  })
  let resp: Response
  try {
    resp = await fetch(`${NOMINATIM_URL}?${params}`)
  } catch {
    throw new Error('Pencarian lokasi gagal, periksa koneksi internet')
  }
  if (!resp.ok) throw new Error('Pencarian lokasi sedang tidak tersedia, coba lagi sebentar')
  const results = (await resp.json()) as NominatimResult[]
  return results.map((r) => ({
    name: r.display_name,
    latitude: Number(r.lat),
    longitude: Number(r.lon),
  }))
}
