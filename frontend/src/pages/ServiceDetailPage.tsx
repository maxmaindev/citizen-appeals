import { useQuery } from '@tanstack/react-query'
import { useParams, Link, useLocation } from 'react-router-dom'
import { appealsAPI } from '../lib/api'
import {
  ArrowLeft,
  Building2,
  FileText,
  Clock,
  CheckCircle,
  AlertTriangle,
  TrendingUp,
  BarChart3,
  Calendar,
  List,
} from 'lucide-react'
import { format } from 'date-fns'
import { uk } from 'date-fns/locale'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  LineChart,
  Line,
} from 'recharts'

const STATUS_COLORS: Record<string, string> = {
  new: '#3B82F6',
  assigned: '#F59E0B',
  in_progress: '#A855F7',
  completed: '#10B981',
  closed: '#10B981',
  rejected: '#EF4444',
}

const STATUS_LABELS: Record<string, string> = {
  new: 'Нові',
  assigned: 'Призначені',
  in_progress: 'В роботі',
  completed: 'Виконано',
  closed: 'Закриті',
  rejected: 'Відхилені',
}

export default function ServiceDetailPage() {
  const { serviceId } = useParams<{ serviceId: string }>()
  const serviceIdNum = serviceId ? parseInt(serviceId, 10) : 0
  const location = useLocation()

  const { data, isLoading, error } = useQuery({
    queryKey: ['service-statistics', serviceIdNum],
    queryFn: async () => {
      const response = await appealsAPI.getServiceStatistics(serviceIdNum)
      if (!response.success) {
        throw new Error(response.error || 'Failed to load service statistics')
      }
      return response.data
    },
    enabled: !!serviceIdNum,
  })

  // Визначаємо, звідки прийшли (з адмін дешборду чи з іншого місця)
  const isFromAdminDashboard = location.pathname.includes('/dashboard/admin/services/')
  const backUrl = isFromAdminDashboard ? '/dashboard/admin' : '/'

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-800">
            {error instanceof Error ? error.message : 'Помилка завантаження статистики'}
          </p>
          <Link
            to={backUrl}
            className="mt-4 inline-flex items-center text-primary hover:underline"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Повернутися назад
          </Link>
        </div>
      </div>
    )
  }

  const service = data.service
  const overall = data.overall
  const monthlyTrend = data.monthly_trend || []
  const statusDistribution = data.status_distribution || []
  const recentAppeals = data.recent_appeals || []

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-6">
        <Link
          to={backUrl}
          className="inline-flex items-center text-gray-600 hover:text-gray-900 mb-4"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Повернутися назад
        </Link>
        <div className="flex items-center space-x-3">
          <Building2 className="h-8 w-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{service.name}</h1>
            {service.description && (
              <p className="text-gray-600 mt-1">{service.description}</p>
            )}
          </div>
        </div>
        {service.contact_person && (
          <div className="mt-4 text-sm text-gray-600">
            <p>
              <strong>Контактна особа:</strong> {service.contact_person}
            </p>
            {service.contact_phone && (
              <p>
                <strong>Телефон:</strong> {service.contact_phone}
              </p>
            )}
            {service.contact_email && (
              <p>
                <strong>Email:</strong> {service.contact_email}
              </p>
            )}
          </div>
        )}
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 mb-1">Всього звернень</p>
              <p className="text-3xl font-bold text-gray-900">{overall.total_appeals}</p>
            </div>
            <FileText className="h-8 w-8 text-blue-500" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 mb-1">Прострочених</p>
              <p className="text-3xl font-bold text-red-600">{overall.overdue_count}</p>
              <p className="text-xs text-gray-400 mt-1">незакриті {'>'}30 дн.</p>
            </div>
            <AlertTriangle className="h-8 w-8 text-red-500" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 mb-1">Середній час</p>
              <p className="text-3xl font-bold text-gray-900">
                {overall.avg_days > 0 ? `${overall.avg_days.toFixed(1)}` : '—'}
              </p>
              <p className="text-xs text-gray-400 mt-1">днів</p>
            </div>
            <Clock className="h-8 w-8 text-yellow-500" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 mb-1">В термін</p>
              <p className="text-3xl font-bold text-green-600">
                {overall.on_time_percentage > 0
                  ? `${overall.on_time_percentage.toFixed(1)}%`
                  : '—'}
              </p>
              <p className="text-xs text-gray-400 mt-1">% закритих ≤30 дн.</p>
            </div>
            <CheckCircle className="h-8 w-8 text-green-500" />
          </div>
        </div>
      </div>

      {/* Status Breakdown */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Status Distribution Chart */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
            <BarChart3 className="h-5 w-5 mr-2" />
            Розподіл по статусах
          </h2>
          {statusDistribution.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={statusDistribution}
                  dataKey="count"
                  nameKey="status"
                  cx="50%"
                  cy="50%"
                  outerRadius={100}
                  label={({ status, count }) => `${STATUS_LABELS[status] || status}: ${count}`}
                >
                  {statusDistribution.map((entry: any) => (
                    <Cell
                      key={entry.status}
                      fill={STATUS_COLORS[entry.status] || '#94A3B8'}
                    />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-gray-500 text-center py-8">Немає даних</p>
          )}
        </div>

        {/* Monthly Trend */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
            <TrendingUp className="h-5 w-5 mr-2" />
            Динаміка за місяцями (останні 6 міс.)
          </h2>
          {monthlyTrend.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={monthlyTrend}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="month" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="count"
                  stroke="#3B82F6"
                  strokeWidth={2}
                  name="Кількість звернень"
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-gray-500 text-center py-8">Немає даних</p>
          )}
        </div>
      </div>

      {/* Detailed Statistics Table */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
          <List className="h-5 w-5 mr-2" />
          Детальна статистика
        </h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <p className="text-sm text-gray-500">Нові</p>
            <p className="text-2xl font-bold text-blue-600">{overall.new_count}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Призначені</p>
            <p className="text-2xl font-bold text-yellow-600">{overall.assigned_count}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">В роботі</p>
            <p className="text-2xl font-bold text-purple-600">{overall.in_progress_count}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Виконано</p>
            <p className="text-2xl font-bold text-green-600">{overall.completed_count}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Відхилено</p>
            <p className="text-2xl font-bold text-red-600">{overall.rejected_count}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">В термін</p>
            <p className="text-2xl font-bold text-green-600">{overall.on_time_count}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Всього закрито</p>
            <p className="text-2xl font-bold text-gray-900">{overall.total_completed}</p>
          </div>
        </div>
      </div>

      {/* Recent Appeals */}
      {recentAppeals.length > 0 && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
            <Calendar className="h-5 w-5 mr-2" />
            Останні звернення
          </h2>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    ID
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Заголовок
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Статус
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Створено
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Закрито
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {recentAppeals.map((appeal: any) => (
                  <tr key={appeal.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 whitespace-nowrap">
                      <Link
                        to={`/appeals/${appeal.id}`}
                        className="text-sm font-medium text-primary hover:underline"
                      >
                        #{appeal.id}
                      </Link>
                    </td>
                    <td className="px-4 py-3">
                      <div className="text-sm text-gray-900">{appeal.title}</div>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      <span
                        className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                          appeal.status === 'new'
                            ? 'bg-blue-100 text-blue-800'
                            : appeal.status === 'assigned'
                            ? 'bg-yellow-100 text-yellow-800'
                            : appeal.status === 'in_progress'
                            ? 'bg-purple-100 text-purple-800'
                            : appeal.status === 'completed' || appeal.status === 'closed'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        }`}
                      >
                        {STATUS_LABELS[appeal.status] || appeal.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">
                      {format(new Date(appeal.created_at), 'dd.MM.yyyy HH:mm', { locale: uk })}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">
                      {appeal.closed_at
                        ? format(new Date(appeal.closed_at), 'dd.MM.yyyy HH:mm', { locale: uk })
                        : '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

