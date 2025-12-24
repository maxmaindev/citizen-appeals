import { useQuery } from '@tanstack/react-query'
import { MapContainer, TileLayer, useMap } from 'react-leaflet'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
// @ts-ignore
import 'leaflet.markercluster'
// @ts-ignore
import 'leaflet.markercluster/dist/MarkerCluster.css'
// @ts-ignore
import 'leaflet.markercluster/dist/MarkerCluster.Default.css'
import { appealsAPI, systemSettingsAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { format } from 'date-fns'
import { useEffect, useRef } from 'react'

// Fix for default marker icon
delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png',
})

// Custom marker icons for different statuses
const createStatusIcon = (status: string) => {
  const colors: Record<string, string> = {
    new: '#3B82F6', // blue
    assigned: '#FBBF24', // yellow
    in_progress: '#A855F7', // purple
    completed: '#10B981', // green
    closed: '#6B7280', // gray
    rejected: '#EF4444', // red
  }

  const color = colors[status] || colors.new

  return L.divIcon({
    className: 'custom-marker',
    html: `<div style="
      background-color: ${color};
      width: 20px;
      height: 20px;
      border-radius: 50%;
      border: 3px solid white;
      box-shadow: 0 2px 4px rgba(0,0,0,0.3);
    "></div>`,
    iconSize: [20, 20],
    iconAnchor: [10, 10],
  })
}

const statusLabels: Record<string, string> = {
  new: 'Нове',
  assigned: 'Призначене',
  in_progress: 'В роботі',
  completed: 'Виконане',
  closed: 'Закрите',
  rejected: 'Відхилене',
}

// Component for marker clustering
function MarkerClusterGroup({ appeals, createStatusIcon, statusLabels }: { 
  appeals: any[]
  createStatusIcon: (status: string) => L.DivIcon
  statusLabels: Record<string, string>
}) {
  const map = useMap()
  const clusterGroupRef = useRef<any>(null)

  useEffect(() => {
    if (!map) return

    // Create cluster group
    // @ts-ignore
    const clusterGroup = L.markerClusterGroup({
      chunkedLoading: true,
      maxClusterRadius: 50, // Радіус кластера в пікселях
      iconCreateFunction: function (cluster: any) {
        const count = cluster.getChildCount()
        let size = 'small'
        if (count > 100) size = 'large'
        else if (count > 10) size = 'medium'

        return L.divIcon({
          html: `<div class="marker-cluster marker-cluster-${size}">
            <span>${count}</span>
          </div>`,
          className: 'marker-cluster-container',
          iconSize: L.point(40, 40),
        })
      },
    })

    // Add markers to cluster group
    appeals.forEach((appeal) => {
      const lat = Number(appeal.latitude)
      const lng = Number(appeal.longitude)
      const marker = L.marker([lat, lng], { icon: createStatusIcon(appeal.status) })

      const popupContent = `
        <div style="min-width: 200px;">
          <a href="/appeals/${appeal.id}" style="font-weight: 600; color: #3b82f6; text-decoration: none; display: block; margin-bottom: 8px;">
            ${appeal.title}
          </a>
          <p style="font-size: 14px; color: #4b5563; margin-bottom: 8px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;">
            ${appeal.description}
          </p>
          <div style="font-size: 12px; color: #6b7280; line-height: 1.5;">
            <p><strong>Статус:</strong> ${statusLabels[appeal.status] || appeal.status}</p>
            <p><strong>Адреса:</strong> ${appeal.address}</p>
            <p><strong>Дата:</strong> ${format(new Date(appeal.created_at), 'dd MMM yyyy')}</p>
            ${appeal.category ? `<p><strong>Категорія:</strong> ${appeal.category.name}</p>` : ''}
          </div>
          <a href="/appeals/${appeal.id}" style="margin-top: 8px; display: inline-block; font-size: 12px; color: #3b82f6; text-decoration: none;">
            Детальніше →
          </a>
        </div>
      `
      marker.bindPopup(popupContent)
      clusterGroup.addLayer(marker)
    })

    clusterGroupRef.current = clusterGroup
    map.addLayer(clusterGroup)

    // Cleanup
    return () => {
      if (clusterGroupRef.current) {
        map.removeLayer(clusterGroupRef.current)
        clusterGroupRef.current = null
      }
    }
  }, [map, appeals, createStatusIcon, statusLabels])

  return null
}

export default function AppealsMapPage() {
  const { user } = useAuth()

  const { data, isLoading, error } = useQuery({
    queryKey: ['appeals', 'map'],
    queryFn: async () => {
      const response = await appealsAPI.list({ page: 1, limit: 1000 }) // Більше звернень для карти
      return response.data
    },
    enabled: !!user,
  })

  const { data: settingsData } = useQuery({
    queryKey: ['system-settings'],
    queryFn: async () => {
      const res = await systemSettingsAPI.get()
      return res.data
    },
    // Тепер GET /api/system-settings доступний для всіх автентифікованих користувачів,
    // тому карта може брати центр/зум із цих налаштувань.
    enabled: !!user,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          Помилка завантаження звернень: {error instanceof Error ? error.message : 'Невідома помилка'}
        </div>
      </div>
    )
  }

  // Фільтруємо тільки звернення з валідними координатами
  const allAppeals = data?.items || []
  const appealsWithCoords = allAppeals.filter(
    (appeal) =>
      appeal.latitude != null &&
      appeal.longitude != null &&
      !isNaN(Number(appeal.latitude)) &&
      !isNaN(Number(appeal.longitude)) &&
      Number(appeal.latitude) >= -90 &&
      Number(appeal.latitude) <= 90 &&
      Number(appeal.longitude) >= -180 &&
      Number(appeal.longitude) <= 180
  )

  // Debug info
  console.log('All appeals:', allAppeals.length)
  console.log('Appeals with coords:', appealsWithCoords.length)
  if (appealsWithCoords.length > 0) {
    console.log('First appeal coords:', {
      lat: appealsWithCoords[0].latitude,
      lng: appealsWithCoords[0].longitude,
    })
  }

  const center: [number, number] = [
    settingsData?.map_center_lat ?? 50.4501,
    settingsData?.map_center_lng ?? 30.5234,
  ]

  const zoom = settingsData?.map_zoom ?? (appealsWithCoords.length > 0 ? 12 : 13)

  if (appealsWithCoords.length === 0 && allAppeals.length === 0) {
    return (
      <div className="h-[calc(100vh-8rem)] flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-500 text-lg">Немає звернень для відображення на карті</p>
          <p className="text-gray-400 text-sm mt-2">Створіть звернення, щоб побачити його на карті</p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-[calc(100vh-4rem)] relative" style={{ minHeight: '600px' }}>
      <div className="absolute left-4 bottom-4 z-[1000] bg-white rounded-lg shadow-lg p-4 max-w-xs">
        <h2 className="text-lg font-semibold mb-3">Звернення на карті</h2>
        <div className="space-y-2 text-sm">
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 rounded-full bg-blue-500 border-2 border-white"></div>
            <span>Нові</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 rounded-full bg-yellow-500 border-2 border-white"></div>
            <span>Призначені</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 rounded-full bg-purple-500 border-2 border-white"></div>
            <span>В роботі</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 rounded-full bg-green-500 border-2 border-white"></div>
            <span>Виконані</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 rounded-full bg-gray-500 border-2 border-white"></div>
            <span>Закриті</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 rounded-full bg-red-500 border-2 border-white"></div>
            <span>Відхилені</span>
          </div>
        </div>
        <p className="mt-3 text-xs text-gray-500">
          Звернень на карті: {appealsWithCoords.length}
        </p>
        {allAppeals.length > appealsWithCoords.length && (
          <p className="mt-1 text-xs text-orange-600">
            {allAppeals.length - appealsWithCoords.length} звернень без координат
          </p>
        )}
      </div>

      <MapContainer
        center={center}
        zoom={zoom}
        style={{ height: '100%', width: '100%', zIndex: 1 }}
        scrollWheelZoom={true}
        key={appealsWithCoords.length} // Force re-render when appeals change
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        {appealsWithCoords.length > 0 && (
          <MarkerClusterGroup
            appeals={appealsWithCoords}
            createStatusIcon={createStatusIcon}
            statusLabels={statusLabels}
          />
        )}
      </MapContainer>
    </div>
  )
}

