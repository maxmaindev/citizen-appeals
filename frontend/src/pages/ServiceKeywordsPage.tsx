import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { servicesAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, Save, X, Edit } from 'lucide-react'

export default function ServiceKeywordsPage() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [editingId, setEditingId] = useState<number | null>(null)
  const [keywordsData, setKeywordsData] = useState<{ [key: number]: string }>({})

  const { data: services, isLoading } = useQuery({
    queryKey: ['services'],
    queryFn: async () => {
      const response = await servicesAPI.list()
      return response.data || []
    },
    enabled: (user?.role === 'admin' || user?.role === 'dispatcher') && !!user,
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, keywords }: { id: number; keywords: string }) =>
      servicesAPI.update(id, { keywords }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['services'] })
      setEditingId(null)
      setKeywordsData({})
    },
  })

  const handleEdit = (service: any) => {
    setEditingId(service.id)
    setKeywordsData({
      [service.id]: service.keywords || '',
    })
  }

  const handleCancel = () => {
    setEditingId(null)
    setKeywordsData({})
  }

  const handleSave = (id: number) => {
    const keywords = keywordsData[id] || ''
    updateMutation.mutate({ id, keywords })
  }

  if (!user || (user.role !== 'admin' && user.role !== 'dispatcher')) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        У вас немає доступу до цієї сторінки
      </div>
    )
  }

  const isAdmin = user.role === 'admin'

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-4">
          <button
            onClick={() => navigate('/services')}
            className="flex items-center space-x-2 text-gray-600 hover:text-gray-900"
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <h1 className="text-3xl font-bold text-gray-900">
            {isAdmin ? 'Редагування ключових слів служб' : 'Ключові слова служб'}
          </h1>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 bg-gray-50 border-b border-gray-200">
          <p className="text-sm text-gray-600">
            Ключові слова використовуються для автоматичної класифікації звернень. 
            Введіть слова через пробіл, які найкраще описують службу.
          </p>
        </div>

        <div className="divide-y divide-gray-200">
          {services?.map((service: any) => (
            <div key={service.id} className="p-6 hover:bg-gray-50">
              {editingId === service.id ? (
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="text-lg font-semibold text-gray-900">{service.name}</h3>
                    <div className="flex items-center space-x-2">
                      <button
                        onClick={() => handleSave(service.id)}
                        disabled={updateMutation.isPending}
                        className="flex items-center space-x-2 px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90 disabled:opacity-50"
                      >
                        <Save className="h-4 w-4" />
                        <span>Зберегти</span>
                      </button>
                      <button
                        onClick={handleCancel}
                        disabled={updateMutation.isPending}
                        className="flex items-center space-x-2 px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 disabled:opacity-50"
                      >
                        <X className="h-4 w-4" />
                        <span>Скасувати</span>
                      </button>
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Ключові слова
                    </label>
                    <textarea
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                      rows={4}
                      value={keywordsData[service.id] || ''}
                      onChange={(e) =>
                        setKeywordsData({
                          ...keywordsData,
                          [service.id]: e.target.value,
                        })
                      }
                      placeholder="вода водопостачання каналізація труба прорив..."
                    />
                    <p className="mt-1 text-xs text-gray-500">
                      Введіть ключові слова через пробіл. Ці слова використовуються для автоматичної класифікації звернень.
                    </p>
                    {keywordsData[service.id] && (
                      <p className="mt-2 text-xs text-gray-600">
                        Слів: {keywordsData[service.id].trim().split(/\s+/).filter((w: string) => w.length > 0).length}
                      </p>
                    )}
                  </div>
                </div>
              ) : (
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">{service.name}</h3>
                    {service.keywords ? (
                      <div className="bg-gray-50 border border-gray-200 rounded-md p-3">
                        <p className="text-sm text-gray-700 whitespace-pre-wrap break-words">
                          {service.keywords}
                        </p>
                        <p className="mt-2 text-xs text-gray-500">
                          Слів: {service.keywords.trim().split(/\s+/).filter((w: string) => w.length > 0).length}
                        </p>
                      </div>
                    ) : (
                      <div className="bg-yellow-50 border border-yellow-200 rounded-md p-3">
                        <p className="text-sm text-yellow-700">
                          {isAdmin 
                            ? 'Ключові слова не встановлені. Натисніть "Редагувати", щоб додати.'
                            : 'Ключові слова не встановлені.'}
                        </p>
                      </div>
                    )}
                  </div>
                  {isAdmin && (
                    <div className="ml-4">
                      <button
                        onClick={() => handleEdit(service)}
                        className="flex items-center space-x-2 px-4 py-2 border border-primary text-primary rounded-md hover:bg-primary/5"
                      >
                        <Edit className="h-4 w-4" />
                        <span>Редагувати</span>
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

