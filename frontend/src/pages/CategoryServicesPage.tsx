import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { categoryServicesAPI, servicesAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Save, Check, X, ArrowLeft } from 'lucide-react'

export default function CategoryServicesPage() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [editingCategoryId, setEditingCategoryId] = useState<number | null>(null)
  const [selectedServices, setSelectedServices] = useState<{ [key: number]: number[] }>({})

  const { data: categoryServices, isLoading } = useQuery({
    queryKey: ['category-services'],
    queryFn: async () => {
      const response = await categoryServicesAPI.getAll()
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

  const assignMutation = useMutation({
    mutationFn: ({ categoryId, serviceIds }: { categoryId: number; serviceIds: number[] }) =>
      categoryServicesAPI.assignServices(categoryId, serviceIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['category-services'] })
      setEditingCategoryId(null)
      setSelectedServices({})
    },
  })

  const handleEdit = (categoryId: number, currentServiceIds: number[]) => {
    setEditingCategoryId(categoryId)
    setSelectedServices({ [categoryId]: [...currentServiceIds] })
  }

  const handleCancel = () => {
    setEditingCategoryId(null)
    setSelectedServices({})
  }

  const handleServiceToggle = (categoryId: number, serviceId: number) => {
    const current = selectedServices[categoryId] || []
    const newSelection = current.includes(serviceId)
      ? current.filter(id => id !== serviceId)
      : [...current, serviceId]
    setSelectedServices({ ...selectedServices, [categoryId]: newSelection })
  }

  const handleSave = (categoryId: number) => {
    const serviceIds = selectedServices[categoryId] || []
    assignMutation.mutate({ categoryId, serviceIds })
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
          <h1 className="text-3xl font-bold text-gray-900">Призначення служб до категорій</h1>
        </div>
      </div>
      <p className="text-gray-600 mb-6">
        Оберіть служби, які будуть відповідати за обробку звернень кожної категорії
      </p>

      <div className="space-y-6">
        {categoryServices?.map((item: any) => {
          const category = item.category
          const assignedServices = item.services || []
          const assignedServiceIds = assignedServices.map((cs: any) => cs.service_id)
          const isEditing = editingCategoryId === category.id
          const currentSelection = selectedServices[category.id] || assignedServiceIds

          return (
            <div key={category.id} className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">{category.name}</h3>
                  {category.description && (
                    <p className="text-sm text-gray-600 mt-1">{category.description}</p>
                  )}
                </div>
                {!isEditing && (
                  <button
                    onClick={() => handleEdit(category.id, assignedServiceIds)}
                    className="px-4 py-2 text-sm font-medium text-primary hover:text-primary/80 border border-primary rounded-md hover:bg-primary/5"
                  >
                    Редагувати
                  </button>
                )}
              </div>

              {isEditing ? (
                <div className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                    {allServices?.map((service: any) => (
                      <label
                        key={service.id}
                        className={`flex items-center space-x-2 p-3 border rounded-md cursor-pointer transition-colors ${
                          currentSelection.includes(service.id)
                            ? 'border-primary bg-primary/5'
                            : 'border-gray-300 hover:border-gray-400'
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={currentSelection.includes(service.id)}
                          onChange={() => handleServiceToggle(category.id, service.id)}
                          className="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary"
                        />
                        <div className="flex-1">
                          <div className="font-medium text-gray-900">{service.name}</div>
                          {service.description && (
                            <div className="text-xs text-gray-500 mt-1">{service.description}</div>
                          )}
                        </div>
                      </label>
                    ))}
                  </div>
                  <div className="flex items-center space-x-2 pt-2 border-t">
                    <button
                      onClick={() => handleSave(category.id)}
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
                  {assignedServices.length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                      {assignedServices.map((cs: any) => (
                        <span
                          key={cs.id}
                          className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-primary/10 text-primary border border-primary/20"
                        >
                          <Check className="h-3 w-3 mr-1" />
                          {cs.service?.name || 'Невідома служба'}
                        </span>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-gray-500 italic">Служби не призначені</p>
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

