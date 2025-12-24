import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { usersAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { useState } from 'react'
import { Edit, Trash2, Save, X } from 'lucide-react'

const roleLabels: Record<string, string> = {
  citizen: 'Громадянин',
  dispatcher: 'Диспетчер',
  executor: 'Виконавець',
  admin: 'Адміністратор',
}

export default function UsersPage() {
  const { user: currentUser } = useAuth()
  const [page, setPage] = useState(1)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editData, setEditData] = useState<{ role?: string; is_active?: boolean }>({})

  const queryClient = useQueryClient()

  const { data, isLoading, error } = useQuery({
    queryKey: ['users', page],
    queryFn: async () => {
      const response = await usersAPI.list({ page, limit: 20 })
      return response.data
    },
    enabled: currentUser?.role === 'admin',
    retry: false, // Не повторювати запит при помилці
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) => usersAPI.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setEditingId(null)
      setEditData({})
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => usersAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  const handleEdit = (user: any) => {
    setEditingId(user.id)
    setEditData({ role: user.role, is_active: user.is_active })
  }

  const handleSave = (id: number) => {
    updateMutation.mutate({ id, data: editData })
  }

  const handleCancel = () => {
    setEditingId(null)
    setEditData({})
  }

  if (!currentUser) {
    return (
      <div className="bg-yellow-50 border border-yellow-200 text-yellow-700 px-4 py-3 rounded">
        Завантаження...
      </div>
    )
  }

  if (currentUser.role !== 'admin') {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        <p className="font-semibold">У вас немає доступу до цієї сторінки</p>
        <p className="text-sm mt-1">Потрібна роль: <strong>admin</strong></p>
        <p className="text-sm mt-1">Ваша роль: <strong>{currentUser.role}</strong></p>
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

  if (error) {
    const errorMessage = error instanceof Error ? error.message : 'Невідома помилка'
    const statusCode = (error as any)?.response?.status
    
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        <p className="font-semibold">Помилка завантаження користувачів</p>
        <p className="text-sm mt-1">{errorMessage}</p>
        {statusCode === 403 && (
          <p className="text-xs mt-2 text-gray-600">
            Помилка 403: У вас немає доступу. Потрібна роль <strong>admin</strong>.
            Ваша роль: <strong>{currentUser?.role}</strong>
          </p>
        )}
        {statusCode === 404 && (
          <p className="text-xs mt-2 text-gray-600">
            Помилка 404: Endpoint не знайдено. Перевірте:
            <ul className="list-disc list-inside mt-1">
              <li>Чи запущений бекенд на http://localhost:8080</li>
              <li>Чи правильно налаштований VITE_API_URL</li>
            </ul>
          </p>
        )}
        {statusCode === 401 && (
          <p className="text-xs mt-2 text-gray-600">
            Помилка 401: Не авторизовано. Спробуйте вийти та зайти знову.
          </p>
        )}
        {!statusCode && (
          <p className="text-xs mt-2 text-gray-600">
            Можливо, бекенд не запущений або є проблема з мережею.
          </p>
        )}
      </div>
    )
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Управління користувачами</h1>
        <p className="text-gray-600 mt-2">Всього користувачів: {data?.total || 0}</p>
      </div>

      {data?.items && data.items.length > 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  ID
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Ім'я
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Email
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Телефон
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Роль
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Статус
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Дії
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {data.items.map((user) => (
                <tr key={user.id} className={user.id === currentUser?.id ? 'bg-blue-50' : ''}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {user.id}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {user.first_name} {user.last_name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {user.email}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {user.phone}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {editingId === user.id ? (
                      <select
                        value={editData.role || user.role}
                        onChange={(e) => setEditData({ ...editData, role: e.target.value })}
                        className="border border-gray-300 rounded-md px-2 py-1 text-sm"
                      >
                        <option value="citizen">Громадянин</option>
                        <option value="dispatcher">Диспетчер</option>
                        <option value="executor">Виконавець</option>
                        <option value="admin">Адміністратор</option>
                      </select>
                    ) : (
                      <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                        user.role === 'admin' ? 'bg-purple-100 text-purple-800' :
                        user.role === 'dispatcher' ? 'bg-blue-100 text-blue-800' :
                        user.role === 'executor' ? 'bg-green-100 text-green-800' :
                        'bg-gray-100 text-gray-800'
                      }`}>
                        {roleLabels[user.role] || user.role}
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {editingId === user.id ? (
                      <select
                        value={editData.is_active !== undefined ? (editData.is_active ? 'true' : 'false') : (user.is_active ? 'true' : 'false')}
                        onChange={(e) => setEditData({ ...editData, is_active: e.target.value === 'true' })}
                        className="border border-gray-300 rounded-md px-2 py-1 text-sm"
                      >
                        <option value="true">Активний</option>
                        <option value="false">Неактивний</option>
                      </select>
                    ) : (
                      <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                        user.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                      }`}>
                        {user.is_active ? 'Активний' : 'Неактивний'}
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    {editingId === user.id ? (
                      <div className="flex space-x-2">
                        <button
                          onClick={() => handleSave(user.id)}
                          disabled={updateMutation.isPending}
                          className="text-green-600 hover:text-green-900"
                        >
                          <Save className="h-4 w-4" />
                        </button>
                        <button
                          onClick={handleCancel}
                          className="text-gray-600 hover:text-gray-900"
                        >
                          <X className="h-4 w-4" />
                        </button>
                      </div>
                    ) : (
                      <div className="flex space-x-2">
                        <button
                          onClick={() => handleEdit(user)}
                          className="text-blue-600 hover:text-blue-900"
                        >
                          <Edit className="h-4 w-4" />
                        </button>
                        {user.id !== currentUser?.id && (
                          <button
                            onClick={() => {
                              if (confirm('Ви впевнені, що хочете деактивувати цього користувача?')) {
                                deleteMutation.mutate(user.id)
                              }
                            }}
                            disabled={deleteMutation.isPending}
                            className="text-red-600 hover:text-red-900"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        )}
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-center py-12">
          <p className="text-gray-500">Немає користувачів</p>
        </div>
      )}

      {/* Pagination */}
      {data && data.total_pages > 1 && (
        <div className="mt-6 flex justify-center space-x-2">
          <button
            onClick={() => setPage(p => Math.max(1, p - 1))}
            disabled={page === 1}
            className="px-4 py-2 border border-gray-300 rounded-md disabled:opacity-50"
          >
            Попередня
          </button>
          <span className="px-4 py-2">
            Сторінка {page} з {data.total_pages}
          </span>
          <button
            onClick={() => setPage(p => Math.min(data.total_pages, p + 1))}
            disabled={page === data.total_pages}
            className="px-4 py-2 border border-gray-300 rounded-md disabled:opacity-50"
          >
            Наступна
          </button>
        </div>
      )}
    </div>
  )
}

