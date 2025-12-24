import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { appealsAPI, categoriesAPI, photosAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import MapPicker from '../components/MapPicker'
import PhotoSelector from '../components/PhotoSelector'

// Reverse geocoding функція для отримання адреси з координат
async function reverseGeocode(lat: number, lng: number): Promise<string> {
  try {
    const response = await fetch(
      `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}&zoom=18&addressdetails=1&accept-language=uk`,
      {
        headers: {
          'User-Agent': 'CitizenAppeals/1.0', // Nominatim вимагає User-Agent
        },
      }
    )
    
    if (!response.ok) {
      throw new Error('Failed to fetch address')
    }
    
    const data = await response.json()
    
    if (data.address) {
      // Формуємо адресу: тільки вулиця, номер будинку та місто
      const parts: string[] = []
      
      // Вулиця
      if (data.address.road) parts.push(data.address.road)
      
      // Номер будинку
      if (data.address.house_number) parts.push(data.address.house_number)
      
      // Місто (city, town або village)
      const city = data.address.city || data.address.town || data.address.village
      if (city) parts.push(city)
      
      return parts.length > 0 ? parts.join(', ') : data.display_name || ''
    }
    
    return data.display_name || ''
  } catch (error) {
    console.error('Reverse geocoding error:', error)
    return ''
  }
}

export default function CreateAppealPage() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    category_id: '',
    address: '',
    latitude: 50.4501,
    longitude: 30.5234,
    priority: 2, // Середній за замовчуванням
  })
  const [errors, setErrors] = useState<{ [key: string]: string }>({})
  const [selectedPhotos, setSelectedPhotos] = useState<File[]>([])
  const [isLoadingAddress, setIsLoadingAddress] = useState(false)
  const [addressManuallyEdited, setAddressManuallyEdited] = useState(false)

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      const response = await categoriesAPI.list()
      return response.data || []
    },
    enabled: !!user,
  })

  const [createdAppealId, setCreatedAppealId] = useState<number | null>(null)

  const mutation = useMutation({
    mutationFn: appealsAPI.create,
    onSuccess: async (response) => {
      if (response.success && response.data) {
        setCreatedAppealId(response.data.id)
        // Інвалідуємо кеш списку звернень
        queryClient.invalidateQueries({ queryKey: ['appeals'] })
        
        // Автоматично завантажуємо фото, якщо вони були вибрані
        if (selectedPhotos.length > 0) {
          try {
            await photosAPI.upload(response.data.id, selectedPhotos)
            queryClient.invalidateQueries({ queryKey: ['photos', response.data.id] })
            setSelectedPhotos([])
          } catch (err) {
            console.error('Failed to upload photos:', err)
            // Не блокуємо успішне створення звернення, навіть якщо фото не завантажилися
          }
        }
      }
    },
  })

  const validateForm = () => {
    const newErrors: { [key: string]: string } = {}

    if (formData.title.trim().length < 5) {
      newErrors.title = 'Заголовок має містити мінімум 5 символів'
    }

    if (formData.description.trim().length < 10) {
      newErrors.description = 'Опис має містити мінімум 10 символів'
    }

    if (!formData.address.trim()) {
      newErrors.address = 'Адреса обов\'язкова'
    }

    if (!formData.category_id) {
      newErrors.category_id = 'Категорія обов\'язкова'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!validateForm()) {
      return
    }

    mutation.mutate({
      title: formData.title.trim(),
      description: formData.description.trim(),
      address: formData.address.trim(),
      latitude: formData.latitude,
      longitude: formData.longitude,
      priority: formData.priority,
      category_id: parseInt(formData.category_id),
    })
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Створити звернення</h1>

      <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 space-y-6">
        <div>
          <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-2">
            Заголовок * <span className="text-gray-500 text-xs">(мінімум 5 символів)</span>
          </label>
          <input
            id="title"
            type="text"
            required
            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-primary focus:border-primary ${
              errors.title ? 'border-red-300' : 'border-gray-300'
            }`}
            value={formData.title}
            onChange={(e) => {
              setFormData({ ...formData, title: e.target.value })
              if (errors.title) setErrors({ ...errors, title: '' })
            }}
          />
          {errors.title && (
            <p className="mt-1 text-sm text-red-600">{errors.title}</p>
          )}
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
            Опис * <span className="text-gray-500 text-xs">(мінімум 10 символів)</span>
          </label>
          <textarea
            id="description"
            required
            rows={5}
            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-primary focus:border-primary ${
              errors.description ? 'border-red-300' : 'border-gray-300'
            }`}
            value={formData.description}
            onChange={(e) => {
              setFormData({ ...formData, description: e.target.value })
              if (errors.description) setErrors({ ...errors, description: '' })
            }}
          />
          {errors.description && (
            <p className="mt-1 text-sm text-red-600">{errors.description}</p>
          )}
          <p className="mt-1 text-xs text-gray-500">
            Введено: {formData.description.length} символів
          </p>
        </div>

        <div>
          <label htmlFor="category_id" className="block text-sm font-medium text-gray-700 mb-2">
            Категорія * <span className="text-gray-500 text-xs">(обов'язково)</span>
          </label>
          <select
            id="category_id"
            required
            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-primary focus:border-primary ${
              errors.category_id ? 'border-red-300' : 'border-gray-300'
            }`}
            value={formData.category_id}
            onChange={(e) => {
              setFormData({ ...formData, category_id: e.target.value })
              if (errors.category_id) setErrors({ ...errors, category_id: '' })
            }}
          >
            <option value="">Оберіть категорію</option>
            {categories?.map((cat) => (
              <option key={cat.id} value={cat.id}>
                {cat.name}
              </option>
            ))}
          </select>
          {errors.category_id && (
            <p className="mt-1 text-sm text-red-600">{errors.category_id}</p>
          )}
        </div>

        <div>
          <label htmlFor="address" className="block text-sm font-medium text-gray-700 mb-2">
            Адреса *
          </label>
          <div className="relative">
            <input
              id="address"
              type="text"
              required
              className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-primary focus:border-primary ${
                errors.address ? 'border-red-300' : 'border-gray-300'
              } ${isLoadingAddress ? 'bg-gray-50' : ''}`}
              value={formData.address}
              onChange={(e) => {
                setFormData({ ...formData, address: e.target.value })
                setAddressManuallyEdited(true) // Відмічаємо, що користувач редагував адресу вручну
                if (errors.address) setErrors({ ...errors, address: '' })
              }}
              placeholder={isLoadingAddress ? 'Отримання адреси...' : 'Введіть адресу або виберіть на карті'}
              disabled={isLoadingAddress}
            />
            {isLoadingAddress && (
              <div className="absolute right-3 top-2.5">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
              </div>
            )}
          </div>
          {errors.address && (
            <p className="mt-1 text-sm text-red-600">{errors.address}</p>
          )}
          <p className="mt-1 text-xs text-gray-500">
            Адреса автоматично заповниться при виборі точки на карті
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Оберіть локацію на карті
          </label>
          <MapPicker
            latitude={formData.latitude}
            longitude={formData.longitude}
            onLocationChange={async (lat, lng) => {
              setFormData({ ...formData, latitude: lat, longitude: lng })
              
              // Автоматично отримуємо адресу тільки якщо користувач не редагував її вручну
              if (!addressManuallyEdited) {
                setIsLoadingAddress(true)
                try {
                  const address = await reverseGeocode(lat, lng)
                  if (address) {
                    setFormData((prev) => ({ ...prev, address, latitude: lat, longitude: lng }))
                  }
                } catch (error) {
                  console.error('Failed to get address:', error)
                } finally {
                  setIsLoadingAddress(false)
                }
              }
            }}
          />
          <div className="mt-2 text-sm text-gray-500">
            Клікніть на карту, щоб вибрати координати
          </div>
          <div className="grid grid-cols-2 gap-4 mt-4">
            <div>
              <label htmlFor="latitude" className="block text-sm font-medium text-gray-700 mb-2">
                Широта
              </label>
              <input
                id="latitude"
                type="number"
                step="any"
                required
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                value={formData.latitude.toFixed(6)}
                onChange={(e) => setFormData({ ...formData, latitude: parseFloat(e.target.value) || 0 })}
              />
            </div>
            <div>
              <label htmlFor="longitude" className="block text-sm font-medium text-gray-700 mb-2">
                Довгота
              </label>
              <input
                id="longitude"
                type="number"
                step="any"
                required
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                value={formData.longitude.toFixed(6)}
                onChange={(e) => setFormData({ ...formData, longitude: parseFloat(e.target.value) || 0 })}
              />
            </div>
          </div>
        </div>

        {/* Priority selector - only for dispatcher/admin */}
        {(user?.role === 'dispatcher' || user?.role === 'admin') && (
          <div>
            <label htmlFor="priority" className="block text-sm font-medium text-gray-700 mb-2">
              Пріоритет
            </label>
            <select
              id="priority"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
              value={formData.priority}
              onChange={(e) => setFormData({ ...formData, priority: Number(e.target.value) })}
            >
              <option value={1}>Низький</option>
              <option value={2}>Середній</option>
              <option value={3}>Високий</option>
            </select>
          </div>
        )}

        {/* Photo selector - під картою */}
        <div>
          <PhotoSelector
            onFilesSelected={(files) => {
              setSelectedPhotos(files)
            }}
            maxFiles={5}
            maxSizeMB={5}
            disabled={mutation.isPending || mutation.isSuccess}
          />
          {selectedPhotos.length > 0 && !mutation.isSuccess && (
            <p className="mt-2 text-sm text-gray-600">
              Вибрано {selectedPhotos.length} фото. Вони будуть завантажені після створення звернення.
            </p>
          )}
        </div>

        {mutation.error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
            <p className="font-medium">Помилка створення звернення:</p>
            <p className="text-sm mt-1">
              {mutation.error instanceof Error 
                ? mutation.error.message 
                : 'Перевірте правильність заповнення форми'}
            </p>
            <p className="text-xs mt-2 text-red-600">
              Заголовок: мінімум 5 символів | Опис: мінімум 10 символів
            </p>
          </div>
        )}
        
        {/* Success message and navigation */}
        {mutation.isSuccess && createdAppealId && (
          <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded">
            <p className="font-medium">Звернення успішно створено!</p>
            {selectedPhotos.length > 0 && (
              <p className="text-sm mt-1">Фотографії завантажуються...</p>
            )}
          </div>
        )}

        <div className="flex space-x-4">
          {mutation.isSuccess ? (
            <>
              <button
                onClick={() => navigate('/')}
                className="flex-1 bg-gray-100 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-200"
              >
                Перейти до списку
              </button>
              <button
                onClick={() => navigate(`/appeals/${createdAppealId}`)}
                className="flex-1 bg-primary text-white px-4 py-2 rounded-md hover:bg-primary/90"
              >
                Переглянути звернення
              </button>
            </>
          ) : (
            <>
              <button
                type="submit"
                disabled={mutation.isPending}
                className="flex-1 bg-primary text-white px-4 py-2 rounded-md hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
              >
                {mutation.isPending ? 'Створення...' : 'Створити звернення'}
              </button>
              <button
                type="button"
                onClick={() => navigate('/')}
                className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
              >
                Скасувати
              </button>
            </>
          )}
        </div>
      </form>
    </div>
  )
}

