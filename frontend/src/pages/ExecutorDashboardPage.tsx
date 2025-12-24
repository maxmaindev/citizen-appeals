import { useQuery } from '@tanstack/react-query'
import { appealsAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { format } from 'date-fns'
import { uk } from 'date-fns/locale'
import { CheckCircle, Clock, TrendingUp, Link as LinkIcon } from 'lucide-react'
import { Link } from 'react-router-dom'

export default function ExecutorDashboardPage() {
  const { user } = useAuth()

  const { data: dashboard, isLoading, error } = useQuery({
    queryKey: ['executor-dashboard'],
    queryFn: async () => {
      const response = await appealsAPI.getExecutorDashboard()
      return response.data
    },
    enabled: user?.role === 'executor' && !!user,
  })

  if (!user || user.role !== 'executor') {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        У вас немає доступу до цього дешборду
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
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        Помилка завантаження даних
      </div>
    )
  }

  const activeAppeals = dashboard?.active_appeals || []
  const myAvgTime = dashboard?.my_avg_processing_time || 0
  const serviceAvgTime = dashboard?.service_avg_processing_time || 0

  const statusLabels: Record<string, string> = {
    new: 'Нове',
    assigned: 'Призначене',
    in_progress: 'В роботі',
    completed: 'Виконане',
    closed: 'Закрите',
    rejected: 'Відхилене',
  }

  const getPriorityLabel = (priority: number) => {
    if (priority === 3) return 'Високий'
    if (priority === 2) return 'Середній'
    return 'Низький'
  }

  const getPriorityColor = (priority: number) => {
    if (priority === 3) return 'bg-red-100 border-red-300 text-red-800'
    if (priority === 2) return 'bg-yellow-100 border-yellow-300 text-yellow-800'
    return 'bg-green-100 border-green-300 text-green-800'
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      new: 'bg-blue-100 text-blue-800',
      assigned: 'bg-yellow-100 text-yellow-800',
      in_progress: 'bg-purple-100 text-purple-800',
      completed: 'bg-green-100 text-green-800',
      closed: 'bg-gray-100 text-gray-800',
      rejected: 'bg-red-100 text-red-800',
    }
    return colors[status] || 'bg-gray-100 text-gray-800'
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Мій дешборд</h1>
      </div>

      {/* Performance Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center space-x-3 mb-4">
            <Clock className="h-8 w-8 text-primary" />
            <div>
              <h3 className="text-sm font-medium text-gray-500 mb-1">Мій середній час обробки</h3>
              <p className="text-3xl font-bold text-gray-900">
                {myAvgTime > 0 ? `${myAvgTime.toFixed(1)} дн.` : '—'}
              </p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center space-x-3 mb-4">
            <TrendingUp className="h-8 w-8 text-green-600" />
            <div>
              <h3 className="text-sm font-medium text-gray-500 mb-1">Середній час по службі</h3>
              <p className="text-3xl font-bold text-gray-900">
                {serviceAvgTime > 0 ? `${serviceAvgTime.toFixed(1)} дн.` : '—'}
              </p>
            </div>
          </div>
          {myAvgTime > 0 && serviceAvgTime > 0 && (
            <div className="mt-4">
              {myAvgTime < serviceAvgTime ? (
                <p className="text-green-600 text-sm font-semibold">
                  ✓ Ви працюєте швидше за середнє на {(serviceAvgTime - myAvgTime).toFixed(1)} дн.
                </p>
              ) : myAvgTime > serviceAvgTime ? (
                <p className="text-yellow-600 text-sm font-semibold">
                  ⚠ Ваш час на {(myAvgTime - serviceAvgTime).toFixed(1)} дн. більше за середнє
                </p>
              ) : (
                <p className="text-gray-600 text-sm">Ваш час відповідає середньому</p>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Active Appeals */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-gray-900 flex items-center space-x-2">
            <CheckCircle className="h-5 w-5 text-primary" />
            <span>Мої активні завдання</span>
          </h2>
          <span className="text-sm text-gray-500">
            Всього: {activeAppeals.length}
          </span>
        </div>
        {activeAppeals.length === 0 ? (
          <p className="text-gray-500 text-center py-8">Немає активних завдань</p>
        ) : (
          <div className="space-y-3">
            {activeAppeals.map((appeal: any) => (
              <div
                key={appeal.id}
                className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-2">
                      <Link
                        to={`/appeals/${appeal.id}`}
                        className="font-semibold text-gray-900 hover:text-primary"
                      >
                        {appeal.title}
                      </Link>
                      <span className={`px-2 py-1 rounded text-xs font-semibold ${getPriorityColor(appeal.priority)}`}>
                        Пріоритет: {getPriorityLabel(appeal.priority)}
                      </span>
                      <span className={`px-2 py-1 rounded text-xs font-semibold ${getStatusColor(appeal.status)}`}>
                        {statusLabels[appeal.status] || appeal.status}
                      </span>
                    </div>
                    <div className="flex items-center space-x-4 text-sm text-gray-600">
                      <span>Служба: {appeal.service_name || 'Не призначено'}</span>
                      <span>Днів зі створення: {appeal.days_since_creation}</span>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">
                      Створено: {format(new Date(appeal.created_at), 'dd MMMM yyyy, HH:mm', { locale: uk })}
                    </p>
                  </div>
                  <Link
                    to={`/appeals/${appeal.id}`}
                    className="ml-4 text-primary hover:text-primary/80"
                  >
                    <LinkIcon className="h-5 w-5" />
                  </Link>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

