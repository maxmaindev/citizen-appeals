import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userServicesAPI, usersAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Save, Check, X, ArrowLeft } from 'lucide-react'

export default function UserServicesPage() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [editingServiceId, setEditingServiceId] = useState<number | null>(null)
  const [selectedUsers, setSelectedUsers] = useState<{ [key: number]: number[] }>({})

  const { data: serviceUsers, isLoading } = useQuery({
    queryKey: ['user-services'],
    queryFn: async () => {
      const response = await userServicesAPI.getAll()
      return response.data || []
    },
    enabled: (user?.role === 'dispatcher' || user?.role === 'admin') && !!user,
  })

  const { data: allExecutors } = useQuery({
    queryKey: ['executors'],
    queryFn: async () => {
      const response = await usersAPI.list({ role: 'executor', page: 1, limit: 1000 })
      // Filter executors from response
      if (response.data && 'items' in response.data) {
        return (response.data.items || []).filter((u: any) => u.role === 'executor')
      }
      return []
    },
    enabled: (user?.role === 'dispatcher' || user?.role === 'admin') && !!user,
  })

  const assignMutation = useMutation({
    mutationFn: ({ serviceId, userIds }: { serviceId: number; userIds: number[] }) =>
      userServicesAPI.assignUsers(serviceId, userIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-services'] })
      queryClient.invalidateQueries({ queryKey: ['executors'] })
      setEditingServiceId(null)
      setSelectedUsers({})
    },
  })

  const handleEdit = (serviceId: number, currentUserIds: number[]) => {
    setEditingServiceId(serviceId)
    setSelectedUsers({ [serviceId]: [...currentUserIds] })
  }

  const handleCancel = () => {
    setEditingServiceId(null)
    setSelectedUsers({})
  }

  const handleUserToggle = (serviceId: number, userId: number) => {
    const current = selectedUsers[serviceId] || []
    const newSelection = current.includes(userId)
      ? current.filter(id => id !== userId)
      : [...current, userId]
    setSelectedUsers({ ...selectedUsers, [serviceId]: newSelection })
  }

  const handleSave = (serviceId: number) => {
    const userIds = selectedUsers[serviceId] || []
    assignMutation.mutate({ serviceId, userIds })
  }

  if (!user || (user.role !== 'dispatcher' && user.role !== 'admin')) {
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
        <div>
          <button
            onClick={() => navigate('/services')}
            className="flex items-center space-x-2 text-gray-600 hover:text-gray-900 mb-2"
          >
            <ArrowLeft className="h-4 w-4" />
            <span>Назад до служб</span>
          </button>
          <h1 className="text-3xl font-bold text-gray-900">Призначення виконавців до служб</h1>
        </div>
      </div>
      <p className="text-gray-600 mb-6">
        Оберіть виконавців, які будуть працювати в кожній службі
      </p>

      <div className="space-y-6">
        {serviceUsers?.map((item: any) => {
          const service = item.service
          const assignedUsers = item.users || []
          const assignedUserIds = assignedUsers.map((u: any) => u.id)
          const isEditing = editingServiceId === service.id
          const currentSelection = selectedUsers[service.id] || assignedUserIds

          return (
            <div key={service.id} className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">{service.name}</h3>
                  {service.description && (
                    <p className="text-sm text-gray-600 mt-1">{service.description}</p>
                  )}
                </div>
                {!isEditing && (
                  <button
                    onClick={() => handleEdit(service.id, assignedUserIds)}
                    className="px-4 py-2 text-sm font-medium text-primary hover:text-primary/80 border border-primary rounded-md hover:bg-primary/5"
                  >
                    Редагувати
                  </button>
                )}
              </div>

              {isEditing ? (
                <div className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                    {allExecutors?.map((executor: any) => (
                      <label
                        key={executor.id}
                        className={`flex items-center space-x-2 p-3 border rounded-md cursor-pointer transition-colors ${
                          currentSelection.includes(executor.id)
                            ? 'border-primary bg-primary/5'
                            : 'border-gray-300 hover:border-gray-400'
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={currentSelection.includes(executor.id)}
                          onChange={() => handleUserToggle(service.id, executor.id)}
                          className="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary"
                        />
                        <div className="flex-1">
                          <div className="font-medium text-gray-900">
                            {executor.first_name} {executor.last_name}
                          </div>
                          {executor.email && (
                            <div className="text-xs text-gray-500 mt-1">{executor.email}</div>
                          )}
                        </div>
                      </label>
                    ))}
                  </div>
                  <div className="flex items-center space-x-2 pt-2 border-t">
                    <button
                      onClick={() => handleSave(service.id)}
                      disabled={assignMutation.isPending}
                      className="flex items-center space-x-2 px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90 disabled:opacity-50"
                    >
                      <Save className="h-4 w-4" />
                      <span>Зберегти</span>
                    </button>
                    <button
                      onClick={handleCancel}
                      disabled={assignMutation.isPending}
                      className="flex items-center space-x-2 px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 disabled:opacity-50"
                    >
                      <X className="h-4 w-4" />
                      <span>Скасувати</span>
                    </button>
                  </div>
                </div>
              ) : (
                <div>
                  {assignedUsers.length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                      {assignedUsers.map((u: any) => (
                        <span
                          key={u.id}
                          className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-primary/10 text-primary border border-primary/20"
                        >
                          <Check className="h-3 w-3 mr-1" />
                          {u.first_name} {u.last_name}
                        </span>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-gray-500 italic">Виконавці не призначені</p>
                  )}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

