import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { appealsAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { format } from 'date-fns'
import { uk } from 'date-fns/locale'
import { AlertTriangle, Clock, Timer, Link as LinkIcon, ChevronDown, ChevronUp } from 'lucide-react'
import { Link } from 'react-router-dom'

export default function DispatcherDashboardPage() {
  const { user } = useAuth()
  const [expandedSections, setExpandedSections] = useState({
    overdue: true,
    approaching: true,
    stale: true,
  })

  const { data: dashboard, isLoading, error } = useQuery({
    queryKey: ['dispatcher-dashboard'],
    queryFn: async () => {
      const response = await appealsAPI.getDispatcherDashboard()
      return response.data
    },
    enabled: (user?.role === 'dispatcher' || user?.role === 'admin') && !!user,
    staleTime: 60 * 1000, // Кеш на 1 хвилину
    gcTime: 5 * 60 * 1000, // Зберігати в кеші 5 хвилин
  })

  if (!user || (user.role !== 'dispatcher' && user.role !== 'admin')) {
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

  const overdueAppeals = (dashboard as any)?.overdue_appeals || []
  const staleAppeals = (dashboard as any)?.stale_appeals || []
  const approachingAppeals = (dashboard as any)?.approaching_appeals || []

  const toggleSection = (section: 'overdue' | 'approaching' | 'stale') => {
    setExpandedSections((prev) => ({
      ...prev,
      [section]: !prev[section],
    }))
  }

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

  const AppealCard = ({ appeal, color, borderColor, bgColor, daysText }: any) => (
    <Link
      to={`/appeals/${appeal.id}`}
      className="block border-l-4 p-4 rounded hover:shadow-md transition-shadow"
      style={{
        borderLeftColor: borderColor,
        backgroundColor: bgColor,
      }}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <h3 className="font-semibold text-gray-900 hover:text-primary mb-2">
            {appeal.title}
          </h3>
          <div className="flex flex-col space-y-1 text-sm text-gray-600">
            <span>Статус: {statusLabels[appeal.status] || appeal.status}</span>
            <span>Пріоритет: {getPriorityLabel(appeal.priority)}</span>
            {appeal.service_name && (
              <span>Служба: {appeal.service_name}</span>
            )}
            <span className="font-semibold" style={{ color }}>
              {daysText}
            </span>
          </div>
          <p className="text-xs text-gray-500 mt-2">
            {appeal.created_at
              ? `Створено: ${format(new Date(appeal.created_at), 'dd MMMM yyyy', { locale: uk })}`
              : `Оновлено: ${format(new Date(appeal.updated_at), 'dd MMMM yyyy', { locale: uk })}`}
          </p>
        </div>
        <LinkIcon className="h-5 w-5 text-primary ml-4 flex-shrink-0" />
      </div>
    </Link>
  )

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Дешборд диспетчера</h1>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-red-50 border-2 border-red-200 rounded-lg shadow-sm p-6">
          <div className="flex items-center space-x-3">
            <AlertTriangle className="h-8 w-8 text-red-600" />
            <div>
              <h3 className="text-sm font-medium text-red-600 mb-1">Прострочені звернення</h3>
              <p className="text-3xl font-bold text-red-700">{overdueAppeals.length}</p>
            </div>
          </div>
        </div>
        <div className="bg-orange-50 border-2 border-orange-200 rounded-lg shadow-sm p-6">
          <div className="flex items-center space-x-3">
            <Timer className="h-8 w-8 text-orange-600" />
            <div>
              <h3 className="text-sm font-medium text-orange-600 mb-1">Наближаються до прострочення</h3>
              <p className="text-3xl font-bold text-orange-700">{approachingAppeals.length}</p>
            </div>
          </div>
        </div>
        <div className="bg-yellow-50 border-2 border-yellow-200 rounded-lg shadow-sm p-6">
          <div className="flex items-center space-x-3">
            <Clock className="h-8 w-8 text-yellow-600" />
            <div>
              <h3 className="text-sm font-medium text-yellow-600 mb-1">Без змін {'>'}7 днів</h3>
              <p className="text-3xl font-bold text-yellow-700">{staleAppeals.length}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Accordion Sections */}
      <div className="space-y-4">
        {/* Прострочені звернення */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200">
          <button
            onClick={() => toggleSection('overdue')}
            className="w-full px-6 py-4 flex items-center justify-between hover:bg-gray-50 transition-colors"
          >
            <div className="flex items-center space-x-3">
              <AlertTriangle className="h-5 w-5 text-red-600" />
              <h2 className="text-xl font-semibold text-gray-900">
                Прострочені звернення
              </h2>
              <span className="text-sm text-gray-500">({overdueAppeals.length})</span>
            </div>
            {expandedSections.overdue ? (
              <ChevronUp className="h-5 w-5 text-gray-500" />
            ) : (
              <ChevronDown className="h-5 w-5 text-gray-500" />
            )}
          </button>
          {expandedSections.overdue && (
            <div className="px-6 pb-6">
              {overdueAppeals.length === 0 ? (
                <p className="text-gray-500 text-center py-8">Немає прострочених звернень</p>
              ) : (
                <div className="space-y-3">
                  {overdueAppeals.map((appeal: any) => (
                    <AppealCard
                      key={appeal.id}
                      appeal={appeal}
                      color="#dc2626"
                      borderColor="#ef4444"
                      bgColor="#fef2f2"
                      daysText={`Прострочено на ${appeal.days_overdue} днів`}
                    />
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Наближаються до прострочення */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200">
          <button
            onClick={() => toggleSection('approaching')}
            className="w-full px-6 py-4 flex items-center justify-between hover:bg-gray-50 transition-colors"
          >
            <div className="flex items-center space-x-3">
              <Timer className="h-5 w-5 text-orange-600" />
              <h2 className="text-xl font-semibold text-gray-900">
                Наближаються до прострочення
              </h2>
              <span className="text-sm text-gray-500">({approachingAppeals.length})</span>
            </div>
            {expandedSections.approaching ? (
              <ChevronUp className="h-5 w-5 text-gray-500" />
            ) : (
              <ChevronDown className="h-5 w-5 text-gray-500" />
            )}
          </button>
          {expandedSections.approaching && (
            <div className="px-6 pb-6">
              {approachingAppeals.length === 0 ? (
                <p className="text-gray-500 text-center py-8">Немає звернень, що наближаються до прострочення</p>
              ) : (
                <div className="space-y-3">
                  {approachingAppeals.map((appeal: any) => (
                    <AppealCard
                      key={appeal.id}
                      appeal={appeal}
                      color="#ea580c"
                      borderColor="#fb923c"
                      bgColor="#fff7ed"
                      daysText={`Залишилось ${appeal.days_remaining} днів`}
                    />
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Без змін >3 днів */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200">
          <button
            onClick={() => toggleSection('stale')}
            className="w-full px-6 py-4 flex items-center justify-between hover:bg-gray-50 transition-colors"
          >
            <div className="flex items-center space-x-3">
              <Clock className="h-5 w-5 text-yellow-600" />
              <h2 className="text-xl font-semibold text-gray-900">
                Без змін {'>'}7 днів
              </h2>
              <span className="text-sm text-gray-500">({staleAppeals.length})</span>
            </div>
            {expandedSections.stale ? (
              <ChevronUp className="h-5 w-5 text-gray-500" />
            ) : (
              <ChevronDown className="h-5 w-5 text-gray-500" />
            )}
          </button>
          {expandedSections.stale && (
            <div className="px-6 pb-6">
              {staleAppeals.length === 0 ? (
                <p className="text-gray-500 text-center py-8">Немає звернень без змін</p>
              ) : (
                <div className="space-y-3">
                  {staleAppeals.map((appeal: any) => (
                    <AppealCard
                      key={appeal.id}
                      appeal={appeal}
                      color="#ca8a04"
                      borderColor="#eab308"
                      bgColor="#fefce8"
                      daysText={`Без змін ${appeal.days_stale} днів`}
                    />
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

