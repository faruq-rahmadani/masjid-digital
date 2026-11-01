import 'leaflet/dist/leaflet.css'
import { divIcon, type LatLngTuple } from 'leaflet'
import { useEffect } from 'react'
import { MapContainer, Marker, TileLayer, useMap, useMapEvents } from 'react-leaflet'
import { type Coordinates, roundCoordinates } from '../utils'

// Loaded lazily by LocationCard: Leaflet is the biggest dependency and only this screen needs it.

const INDONESIA: { center: LatLngTuple; zoom: number } = { center: [-2.5, 118], zoom: 4 }
const PIN_ZOOM = 16

// An SVG pin in the brand color instead of Leaflet's PNG marker, whose image paths break under Vite.
const pin = divIcon({
  className: '',
  iconSize: [36, 36],
  iconAnchor: [18, 34],
  html: `<svg width="36" height="36" viewBox="0 0 24 24" class="fill-brand stroke-card" stroke-width="1.2" aria-hidden="true"><path d="M12 2a7 7 0 0 0-7 7c0 5 7 13 7 13s7-8 7-13a7 7 0 0 0-7-7zm0 9.5A2.5 2.5 0 1 1 12 6.5a2.5 2.5 0 0 1 0 5z"/></svg>`,
})

type LocationMapProps = {
  value: Coordinates | null
  onChange: (value: Coordinates) => void
}

export function LocationMap({ value, onChange }: LocationMapProps) {
  return (
    // `isolate` keeps Leaflet's z-indexed panes under the sticky ActionBar.
    <div className="isolate h-40 overflow-hidden rounded-lg">
      <MapContainer
        center={value ? [value.latitude, value.longitude] : INDONESIA.center}
        zoom={value ? PIN_ZOOM : INDONESIA.zoom}
        className="bg-media size-full"
      >
        <TileLayer
          url="https://tile.openstreetmap.org/{z}/{x}/{y}.png"
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          maxZoom={19}
        />
        <PlaceOnTap onChange={onChange} />
        <FollowPin value={value} />
        {value && (
          <Marker
            position={[value.latitude, value.longitude]}
            icon={pin}
            draggable
            eventHandlers={{
              dragend: (event) => {
                const { lat, lng } = event.target.getLatLng()
                onChange(roundCoordinates({ latitude: lat, longitude: lng }))
              },
            }}
          />
        )}
      </MapContainer>
    </div>
  )
}

function PlaceOnTap({ onChange }: Pick<LocationMapProps, 'onChange'>) {
  useMapEvents({
    click: (event) =>
      onChange(roundCoordinates({ latitude: event.latlng.lat, longitude: event.latlng.lng })),
  })
  return null
}

// Pan to the pin when it moves from outside the map (search result, GPS).
function FollowPin({ value }: Pick<LocationMapProps, 'value'>) {
  const map = useMap()
  const latitude = value?.latitude
  const longitude = value?.longitude
  useEffect(() => {
    if (latitude === undefined || longitude === undefined) return
    map.setView([latitude, longitude], Math.max(map.getZoom(), PIN_ZOOM))
  }, [map, latitude, longitude])
  return null
}
