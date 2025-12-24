import { useEffect } from 'react'
import { MapContainer, TileLayer, Marker, useMap, useMapEvents } from 'react-leaflet'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

// Fix for default marker icon in React
delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png',
})

interface MapPickerProps {
  latitude: number
  longitude: number
  onLocationChange: (lat: number, lng: number) => void
  height?: string
  zoom?: number
  centerOnClick?: boolean // Якщо true, карта переміщується до клікнутої точки; якщо false, тільки маркер оновлюється
}

// Component to handle map clicks
function MapClickHandler({ 
  onLocationChange, 
  centerOnClick 
}: { 
  onLocationChange: (lat: number, lng: number) => void
  centerOnClick?: boolean
}) {
  const map = useMap()
  
  useMapEvents({
    click: (e) => {
      const { lat, lng } = e.latlng
      onLocationChange(lat, lng)
      // Якщо centerOnClick === true, переміщуємо карту до клікнутої точки
      if (centerOnClick) {
        map.setView([lat, lng], map.getZoom())
      }
    },
  })
  return null
}

// Sync map view with external props (lat, lng, zoom)
// Працює тільки коли centerOnClick={true}, щоб не конфліктувати з ручним вибором точки
function MapViewSync({
  latitude,
  longitude,
  zoom,
  enabled,
}: {
  latitude: number
  longitude: number
  zoom: number
  enabled: boolean
}) {
  const map = useMap()

  useEffect(() => {
    if (enabled) {
      map.setView([latitude, longitude], zoom)
    }
  }, [latitude, longitude, zoom, map, enabled])

  return null
}

export default function MapPicker({
  latitude,
  longitude,
  onLocationChange,
  height = '400px',
  zoom = 13,
  centerOnClick = false, // За замовчуванням false - карта не переміщується при кліку
}: MapPickerProps) {
  const position: [number, number] = [latitude, longitude]

  return (
    <div className="w-full rounded-lg overflow-hidden border border-gray-300" style={{ height }}>
      <MapContainer
        center={position}
        zoom={zoom}
        style={{ height: '100%', width: '100%' }}
        scrollWheelZoom={true}
      >
        {/* MapViewSync працює тільки коли centerOnClick={true} (на сторінці налаштувань) */}
        {/* Коли centerOnClick={false} (на сторінці створення звернення), карта не синхронізується з пропсами після кліку */}
        <MapViewSync latitude={latitude} longitude={longitude} zoom={zoom} enabled={centerOnClick} />
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <Marker position={position} />
        <MapClickHandler onLocationChange={onLocationChange} centerOnClick={centerOnClick} />
      </MapContainer>
    </div>
  )
}

