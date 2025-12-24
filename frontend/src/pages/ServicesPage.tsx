import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { servicesAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Edit, Trash2, Save, X, Settings, Key } from 'lucide-react'

export default function ServicesPage() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [editingId, setEditingId] = useState<number | null>(null)
  const [isCreating, setIsCreating] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    contact_person: '',
    contact_phone: '',
    contact_email: '',
  })

  const { data: services, isLoading } = useQuery({
    queryKey: ['services'],
    queryFn: async () => {
      const response = await servicesAPI.list()
      return response.data || []
    },
    enabled: (user?.role === 'admin' || user?.role === 'dispatcher') && !!user,
  })

  const createMutation = useMutation({
    mutationFn: (data: any) => servicesAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['services'] })
      setIsCreating(false)
      setFormData({
        name: '',
        description: '',
        contact_person: '',
        contact_phone: '',
        contact_email: '',
      })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) => servicesAPI.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['services'] })
      setEditingId(null)
      setFormData({
        name: '',
        description: '',
        contact_person: '',
        contact_phone: '',
        contact_email: '',
      })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => servicesAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['services'] })
    },
  })

  const handleEdit = (service: any) => {
    setEditingId(service.id)
    setFormData({
      name: service.name,
      description: service.description || '',
      contact_person: service.contact_person,
      contact_phone: service.contact_phone,
      contact_email: service.contact_email,
    })
    setIsCreating(false)
  }

  const handleCreate = () => {
    setIsCreating(true)
    setEditingId(null)
    setFormData({
      name: '',
      description: '',
      contact_person: '',
      contact_phone: '',
      contact_email: '',
    })
  }

  const handleCancel = () => {
    setIsCreating(false)
    setEditingId(null)
    setFormData({
      name: '',
      description: '',
      contact_person: '',
      contact_phone: '',
      contact_email: '',
    })
  }

  const handleSave = () => {
    if (editingId) {
      updateMutation.mutate({ id: editingId, data: formData })
    } else {
      createMutation.mutate(formData)
    }
  }

  if (!user || (user.role !== 'admin' && user.role !== 'dispatcher')) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        У вас немає доступу до цієї сторінки
      </div>
    )
  }

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
        <h1 className="text-3xl font-bold text-gray-900">Управління службами</h1>
        <div className="flex items-center space-x-4">
          {user.role === 'admin' && (
            <button
              onClick={handleCreate}
              className="flex items-center space-x-2 px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90"
            >
              <Plus className="h-4 w-4" />
              <span>Додати службу</span>
            </button>
          )}
          <button
            onClick={() => navigate('/category-services')}
            className="flex items-center space-x-2 px-4 py-2 border border-primary text-primary rounded-md hover:bg-primary/5"
          >
            <Settings className="h-4 w-4" />
            <span>Призначення до категорій</span>
          </button>
          <button
            onClick={() => navigate('/service-users')}
            className="flex items-center space-x-2 px-4 py-2 border border-primary text-primary rounded-md hover:bg-primary/5"
          >
            <Settings className="h-4 w-4" />
            <span>Призначення виконавців</span>
          </button>
          <button
            onClick={() => navigate('/services/keywords')}
            className="flex items-center space-x-2 px-4 py-2 border border-primary text-primary rounded-md hover:bg-primary/5"
          >
            <Key className="h-4 w-4" />
            <span>Ключові слова</span>
          </button>
        </div>
      </div>

      {isCreating && user.role === 'admin' && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Створити нову службу</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Назва *
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="Назва служби"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Опис
              </label>
              <textarea
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                rows={3}
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Опис служби"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Контактна особа *
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                value={formData.contact_person}
                onChange={(e) => setFormData({ ...formData, contact_person: e.target.value })}
                placeholder="ПІБ контактної особи"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Телефон *
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                value={formData.contact_phone}
                onChange={(e) => setFormData({ ...formData, contact_phone: e.target.value })}
                placeholder="+380501234567"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Email *
              </label>
              <input
                type="email"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                value={formData.contact_email}
                onChange={(e) => setFormData({ ...formData, contact_email: e.target.value })}
                placeholder="email@example.com"
              />
            </div>
            <div className="flex items-center space-x-2 pt-2">
              <button
                onClick={handleSave}
                disabled={createMutation.isPending}
                className="flex items-center space-x-2 px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90 disabled:opacity-50"
              >
                <Save className="h-4 w-4" />
                <span>Зберегти</span>
              </button>
              <button
                onClick={handleCancel}
                disabled={createMutation.isPending}
                className="flex items-center space-x-2 px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 disabled:opacity-50"
              >
                <X className="h-4 w-4" />
                <span>Скасувати</span>
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Назва
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Опис
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Контактна особа
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Телефон
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Email
              </th>
              {user.role === 'admin' && (
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Дії
                </th>
              )}
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {services?.map((service: any) => (
              <tr key={service.id}>
                {editingId === service.id && user.role === 'admin' ? (
                  <>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <input
                        type="text"
                        className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                        value={formData.name}
                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      />
                    </td>
                    <td className="px-6 py-4">
                      <textarea
                        className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                        rows={2}
                        value={formData.description}
                        onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                      />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <input
                        type="text"
                        className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                        value={formData.contact_person}
                        onChange={(e) => setFormData({ ...formData, contact_person: e.target.value })}
                      />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <input
                        type="text"
                        className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                        value={formData.contact_phone}
                        onChange={(e) => setFormData({ ...formData, contact_phone: e.target.value })}
                      />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <input
                        type="email"
                        className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                        value={formData.contact_email}
                        onChange={(e) => setFormData({ ...formData, contact_email: e.target.value })}
                      />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={handleSave}
                          disabled={updateMutation.isPending}
                          className="text-green-600 hover:text-green-900 disabled:opacity-50"
                        >
                          <Save className="h-4 w-4" />
                        </button>
                        <button
                          onClick={handleCancel}
                          disabled={updateMutation.isPending}
                          className="text-red-600 hover:text-red-900 disabled:opacity-50"
                        >
                          <X className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </>
                ) : (
                  <>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">{service.name}</div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-500">{service.description || '-'}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{service.contact_person}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{service.contact_phone}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{service.contact_email}</div>
                    </td>
                    {user.role === 'admin' && (
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <div className="flex items-center justify-end space-x-2">
                          <button
                            onClick={() => handleEdit(service)}
                            className="text-blue-600 hover:text-blue-900"
                          >
                            <Edit className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => {
                              if (confirm('Ви впевнені, що хочете видалити цю службу?')) {
                                deleteMutation.mutate(service.id)
                              }
                            }}
                            className="text-red-600 hover:text-red-900"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    )}
                  </>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

