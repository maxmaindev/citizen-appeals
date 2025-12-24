import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userServicesAPI, servicesAPI, usersAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Save, X, ArrowLeft, Plus, Search, Trash2 } from 'lucide-react'

export default function ServiceUsersPage() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [editingServiceId, setEditingServiceId] = useState<number | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedUserIds, setSelectedUserIds] = useState<{ [key: number]: number[] }>({})

  const { data: serviceUsers, isLoading } = useQuery({
    queryKey: ['user-services'],
    queryFn: async () => {
      const response = await userServicesAPI.getAll()
      return response.data || []
    },
    enabled: (user?.role === 'dispatcher' || user?.role === 'admin') && !!user,
  })

  const { data: allServices } = useQuery({
    queryKey: ['services'],
    queryFn: async () => {
      const response = await servicesAPI.list()
      return response.data || []
    },
    enabled: (user?.role === 'dispatcher' || user?.role === 'admin') && !!user,
  })

  const { data: allExecutors } = useQuery({
    queryKey: ['executors', searchQuery],
    queryFn: async () => {
      const response = await usersAPI.list({ role: 'executor', limit: 100 })
      const executors = Array.isArray(response.data) 
        ? response.data 
        : (response.data as any)?.items || []
      
      // Filter by search query
      if (searchQuery.trim()) {
        const query = searchQuery.toLowerCase()
        return executors.filter((u: any) => 
          `${u.first_name} ${u.last_name}`.toLowerCase().includes(query) ||
          u.email.toLowerCase().includes(query)
        )
      }
      return executors
    },
    enabled: (user?.role === 'dispatcher' || user?.role === 'admin') && !!user,
  })

  const assignMutation = useMutation({
    mutationFn: ({ serviceId, userIds }: { serviceId: number; userIds: number[] }) =>
      userServicesAPI.assignUsers(serviceId, userIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-services'] })
      setEditingServiceId(null)
      setSelectedUserIds({})
      setSearchQuery('')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: ({ serviceId, userId }: { serviceId: number; userId: number }) =>
      userServicesAPI.delete(serviceId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-services'] })
    },
  })

  const handleEdit = (serviceId: number, currentUserIds: number[]) => {
    setEditingServiceId(serviceId)
    setSelectedUserIds({ [serviceId]: [...currentUserIds] })
    setSearchQuery('')
  }

  const handleCancel = () => {
    setEditingServiceId(null)
    setSelectedUserIds({})
    setSearchQuery('')
  }

  const handleAddUser = (serviceId: number, userId: number) => {
    const current = selectedUserIds[serviceId] || []
    if (!current.includes(userId)) {
      setSelectedUserIds({ ...selectedUserIds, [serviceId]: [...current, userId] })
      setSearchQuery('')
    }
  }

  const handleRemoveUser = (serviceId: number, userId: number) => {
    const current = selectedUserIds[serviceId] || []
    setSelectedUserIds({ ...selectedUserIds, [serviceId]: current.filter(id => id !== userId) })
  }

  const handleSave = (serviceId: number) => {
    const userIds = selectedUserIds[serviceId] || []
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
        Призначте виконавців до служб для обробки звернень
      </p>

      <div className="space-y-6">
        {serviceUsers?.map((item: any) => {
          const service = item.service
          const assignedUsers = item.users || []
          const assignedUserIds = assignedUsers.map((u: any) => u.id)
          const isEditing = editingServiceId === service.id
          const currentSelection = selectedUserIds[service.id] || assignedUserIds

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
                  {/* Search for users */}
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                    <input
                      type="text"
                      placeholder="Пошук виконавців за ім'ям або email..."
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                    />
                  </div>

                  {/* Search results */}
                  {searchQuery.trim() && (
                    <div className="border border-gray-200 rounded-md max-h-60 overflow-y-auto">
                      {allExecutors
                        ?.filter((u: any) => !currentSelection.includes(u.id))
                        .map((executor: any) => (
                          <div
                            key={executor.id}
                            className="flex items-center justify-between p-3 border-b border-gray-100 hover:bg-gray-50"
                          >
                            <div>
                              <div className="font-medium text-gray-900">
                                {executor.first_name} {executor.last_name}
                              </div>
                              <div className="text-sm text-gray-500">{executor.email}</div>
                            </div>
                            <button
                              onClick={() => handleAddUser(service.id, executor.id)}
                              className="flex items-center space-x-1 px-3 py-1 text-sm text-primary hover:bg-primary/10 rounded-md"
                            >
                              <Plus className="h-4 w-4" />
                              <span>Додати</span>
                            </button>
                          </div>
                        ))}
                      {allExecutors?.filter((u: any) => !currentSelection.includes(u.id)).length === 0 && (
                        <div className="p-4 text-center text-gray-500 text-sm">
                          Виконавців не знайдено
                        </div>
                      )}
                    </div>
                  )}

                  {/* Selected users */}
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-2">
                      Призначені виконавці ({currentSelection.length})
                    </h4>
                    {currentSelection.length > 0 ? (
                      <div className="space-y-2">
                        {allExecutors
                          ?.filter((u: any) => currentSelection.includes(u.id))
                          .map((executor: any) => (
                            <div
                              key={executor.id}
                              className="flex items-center justify-between p-3 bg-gray-50 rounded-md border border-gray-200"
                            >
                              <div>
                                <div className="font-medium text-gray-900">
                                  {executor.first_name} {executor.last_name}
                                </div>
                                <div className="text-sm text-gray-500">{executor.email}</div>
                              </div>
                              <button
                                onClick={() => handleRemoveUser(service.id, executor.id)}
                                className="flex items-center space-x-1 px-3 py-1 text-sm text-red-600 hover:bg-red-50 rounded-md"
                              >
                                <Trash2 className="h-4 w-4" />
                                <span>Видалити</span>
                              </button>
                            </div>
                          ))}
                      </div>
                    ) : (
                      <p className="text-sm text-gray-500 italic">Виконавці не призначені</p>
                    )}
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
                    <div className="space-y-2">
                      {assignedUsers.map((executor: any) => (
                        <div
                          key={executor.id}
                          className="flex items-center justify-between p-3 bg-gray-50 rounded-md border border-gray-200"
                        >
                          <div>
                            <div className="font-medium text-gray-900">
                              {executor.first_name} {executor.last_name}
                            </div>
                            <div className="text-sm text-gray-500">{executor.email}</div>
                          </div>
                          <button
                            onClick={() => {
                              if (confirm('Видалити виконавця зі служби?')) {
                                deleteMutation.mutate({ serviceId: service.id, userId: executor.id })
                              }
                            }}
                            className="text-red-600 hover:text-red-800"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
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

