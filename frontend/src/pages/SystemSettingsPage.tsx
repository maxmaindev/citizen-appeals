import { useEffect, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import MapPicker from '../components/MapPicker'
import { systemSettingsAPI } from '../lib/api'

export default function SystemSettingsPage() {
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['system-settings'],
    queryFn: async () => {
      const res = await systemSettingsAPI.get()
      return res.data
    },
  })

  const [cityName, setCityName] = useState('')
  const [lat, setLat] = useState(50.4501)
  const [lng, setLng] = useState(30.5234)
  const [zoom, setZoom] = useState(13)
  const [confidenceThreshold, setConfidenceThreshold] = useState(0.8)
  const [saveMessage, setSaveMessage] = useState<string | null>(null)

  useEffect(() => {
    if (data) {
      setCityName(data.city_name || '')
      setLat(data.map_center_lat ?? 50.4501)
      setLng(data.map_center_lng ?? 30.5234)
      setZoom(data.map_zoom ?? 13)
      setConfidenceThreshold(data.confidence_threshold ?? 0.8)
    }
  }, [data])

  const updateMutation = useMutation({
    mutationFn: async () =>
      systemSettingsAPI.update({
        city_name: cityName,
        map_center_lat: lat,
        map_center_lng: lng,
        map_zoom: zoom,
        confidence_threshold: confidenceThreshold,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['system-settings'] })
      setSaveMessage('Налаштування збережено')
      setTimeout(() => setSaveMessage(null), 3000)
    },
  })

  const handleLocationChange = (newLat: number, newLng: number) => {
    setLat(newLat)
    setLng(newLng)
  }

  if (isLoading) {
    return <div>Завантаження налаштувань…</div>
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Налаштування системи</h1>

      {saveMessage && (
        <div className="rounded-md bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-800">
          {saveMessage}
        </div>
      )}

      {/* Поля вводу параметрів карти */}
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Назва міста
          </label>
          <input
            type="text"
            value={cityName}
            onChange={(e) => setCityName(e.target.value)}
            className="w-full border border-gray-300 rounded-md px-3 py-2"
            placeholder="Наприклад, Львів"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Широта (lat)
            </label>
            <input
              type="number"
              step="0.0001"
              value={lat}
              onChange={(e) => setLat(parseFloat(e.target.value))}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Довгота (lng)
            </label>
            <input
              type="number"
              step="0.0001"
              value={lng}
              onChange={(e) => setLng(parseFloat(e.target.value))}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Масштаб карти (zoom)
          </label>
          <input
            type="number"
            min={3}
            max={19}
            value={zoom}
            onChange={(e) => setZoom(parseInt(e.target.value || '13', 10))}
            className="w-full border border-gray-300 rounded-md px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Поріг впевненості для класифікації
          </label>
          <input
            type="number"
            step="0.01"
            min={0}
            max={1}
            value={confidenceThreshold}
            onChange={(e) => setConfidenceThreshold(parseFloat(e.target.value) || 0.8)}
            className="w-full border border-gray-300 rounded-md px-3 py-2"
          />
          <p className="mt-1 text-xs text-gray-500">
            Мінімальна впевненість (0.0-1.0) для автоматичного призначення служби. За замовчуванням: 0.8
          </p>
        </div>

        <button
          onClick={() => updateMutation.mutate()}
          disabled={updateMutation.status === 'pending'}
          className="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90 disabled:opacity-60"
        >
          Зберегти
        </button>
      </div>

      {/* Карта для вибору центру / прев'ю для користувачів */}
      <div className="mt-6">
        <p className="mb-2 text-sm text-gray-600">
          Ця карта виглядає так само, як стартова карта у користувачів. Клікніть, щоб змінити центр.
        </p>
        <MapPicker
          latitude={lat}
          longitude={lng}
          zoom={zoom}
          onLocationChange={handleLocationChange}
          height="450px"
          centerOnClick={true}
        />
      </div>
    </div>
  )
}


