import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { appealsAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { TrendingUp, Clock, Calendar, Building2, AlertTriangle, CheckCircle } from 'lucide-react'
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'

export default function AdminDashboardPage() {
  const { user } = useAuth()
  const navigate = useNavigate()

  const { data: dashboard, isLoading, error } = useQuery({
    queryKey: ['admin-dashboard'],
    queryFn: async () => {
      const response = await appealsAPI.getAdminDashboard()
      return response.data
    },
    enabled: user?.role === 'admin' && !!user,
    staleTime: 60 * 1000, // Кеш на 1 хвилину
    cacheTime: 5 * 60 * 1000, // Зберігати в кеші 5 хвилин
  })

  if (!user || user.role !== 'admin') {
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

  const topServices = dashboard?.top_services_by_speed || []
  const monthlyTrend = dashboard?.monthly_trend || []
  const dayOfWeekStats = dashboard?.day_of_week_stats || []
  const allServicesStats = dashboard?.all_services_stats || []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Дешборд адміністратора</h1>
      </div>

      {/* Summary Cards */}
      {allServicesStats.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Всього служб</h3>
            <p className="text-3xl font-bold text-gray-900">{allServicesStats.length}</p>
            <p className="text-xs text-gray-400 mt-1">Активних служб в системі</p>
          </div>
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Всього звернень</h3>
            <p className="text-3xl font-bold text-gray-900">
              {allServicesStats.reduce((sum: number, s: any) => sum + s.total_appeals, 0)}
            </p>
            <p className="text-xs text-gray-400 mt-1">Загальна кількість звернень</p>
          </div>
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2 flex items-center space-x-1">
              <AlertTriangle className="h-4 w-4 text-red-600" />
              <span>Прострочених</span>
            </h3>
            <p className="text-3xl font-bold text-red-600">
              {allServicesStats.reduce((sum: number, s: any) => sum + s.overdue_count, 0)}
            </p>
            <p className="text-xs text-gray-400 mt-1">Незакриті звернення {'>'}30 днів</p>
          </div>
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2 flex items-center space-x-1">
              <CheckCircle className="h-4 w-4 text-green-600" />
              <span>Середній % в термін</span>
            </h3>
            <p className="text-3xl font-bold text-green-600">
              {(() => {
                const servicesWithData = allServicesStats.filter((s: any) => s.total_completed > 0)
                if (servicesWithData.length === 0) return '—'
                const avg = servicesWithData.reduce((sum: number, s: any) => sum + s.on_time_percentage, 0) / servicesWithData.length
                return `${Math.round(avg)}%`
              })()}
            </p>
            <p className="text-xs text-gray-400 mt-1">% закритих в термін (≤30 днів)</p>
          </div>
        </div>
      )}

      {/* All Services Statistics */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center space-x-2">
          <Building2 className="h-5 w-5 text-primary" />
          <span>Статистика по всім службам</span>
        </h2>
        {allServicesStats.length === 0 ? (
          <p className="text-gray-500 text-center py-8">Немає даних</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Служба
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Всього
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Нові
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Призначені
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                    В роботі
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Виконано
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                    <div className="flex flex-col items-center">
                      <span>Прострочено</span>
                      <span className="text-xs font-normal normal-case text-gray-400 mt-0.5">незакриті {'>'}30 дн.</span>
                    </div>
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                    <div className="flex flex-col items-center">
                      <span>Серед. час</span>
                      <span className="text-xs font-normal normal-case text-gray-400 mt-0.5">днів</span>
                    </div>
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                    <div className="flex flex-col items-center">
                      <span>В термін</span>
                      <span className="text-xs font-normal normal-case text-gray-400 mt-0.5">% закритих ≤30 дн.</span>
                    </div>
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {allServicesStats.map((service: any) => (
                  <tr
                    key={service.id}
                    className="hover:bg-gray-50 cursor-pointer transition-colors"
                    onClick={() => navigate(`/dashboard/admin/services/${service.id}`)}
                  >
                    <td className="px-4 py-3 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">{service.name}</div>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-center">
                      <span className="text-sm text-gray-900 font-semibold">{service.total_appeals}</span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-center">
                      <span className="text-sm text-blue-600 font-medium">{service.new_count}</span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-center">
                      <span className="text-sm text-yellow-600 font-medium">{service.assigned_count}</span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-center">
                      <span className="text-sm text-purple-600 font-medium">{service.in_progress_count}</span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-center">
                      <span className="text-sm text-green-600 font-medium">{service.completed_count}</span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-center">
                      {service.overdue_count > 0 ? (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                          <AlertTriangle className="h-3 w-3 mr-1" />
                          {service.overdue_count}
                        </span>
                      ) : (
                        <span className="text-sm text-gray-500">0</span>
                      )}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-center">
                      {service.avg_days > 0 ? (
                        <span className={`text-sm font-medium ${
                          service.avg_days <= 7 ? 'text-green-600' : 
                          service.avg_days <= 15 ? 'text-yellow-600' : 'text-red-600'
                        }`}>
                          {service.avg_days.toFixed(1)} дн.
                        </span>
                      ) : (
                        <span className="text-sm text-gray-400">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-center">
                      {service.total_completed > 0 ? (
                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                          service.on_time_percentage >= 80 ? 'bg-green-100 text-green-800' :
                          service.on_time_percentage >= 60 ? 'bg-yellow-100 text-yellow-800' :
                          'bg-red-100 text-red-800'
                        }`}>
                          {service.on_time_percentage >= 80 && <CheckCircle className="h-3 w-3 mr-1" />}
                          {service.on_time_percentage.toFixed(0)}%
                        </span>
                      ) : (
                        <span className="text-sm text-gray-400">—</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Service Comparison Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Average Processing Time by Service */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center space-x-2">
            <Clock className="h-5 w-5 text-primary" />
            <span>Середній час обробки по службах</span>
          </h2>
          {allServicesStats.filter((s: any) => s.avg_days > 0).length === 0 ? (
            <p className="text-gray-500 text-center py-8">Немає даних</p>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart 
                data={allServicesStats
                  .filter((s: any) => s.avg_days > 0)
                  .sort((a: any, b: any) => a.avg_days - b.avg_days)
                  .slice(0, 10)
                }
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="name"
                  angle={-45}
                  textAnchor="end"
                  height={120}
                  interval={0}
                />
                <YAxis />
                <Tooltip formatter={(value: number) => `${value.toFixed(1)} днів`} />
                <Bar dataKey="avg_days" fill="#3b82f6" name="Середній час (дні)" />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* On-Time Percentage by Service */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center space-x-2">
            <CheckCircle className="h-5 w-5 text-green-600" />
            <span>% виконаних в термін по службах</span>
          </h2>
          {allServicesStats.filter((s: any) => s.total_completed > 0).length === 0 ? (
            <p className="text-gray-500 text-center py-8">Немає даних</p>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart 
                data={allServicesStats
                  .filter((s: any) => s.total_completed > 0)
                  .sort((a: any, b: any) => b.on_time_percentage - a.on_time_percentage)
                  .slice(0, 10)
                }
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="name"
                  angle={-45}
                  textAnchor="end"
                  height={120}
                  interval={0}
                />
                <YAxis domain={[0, 100]} />
                <Tooltip formatter={(value: number) => `${value.toFixed(0)}%`} />
                <Bar dataKey="on_time_percentage" fill="#10b981" name="В термін (%)" />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>

      {/* Top Services by Speed */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center space-x-2">
          <Clock className="h-5 w-5 text-primary" />
          <span>Топ служб по швидкості обробки (за останні 90 днів)</span>
        </h2>
        {topServices.length === 0 ? (
          <p className="text-gray-500 text-center py-8">Немає даних</p>
        ) : (
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
              {topServices.slice(0, 6).map((service: any, index: number) => (
                <div
                  key={index}
                  className="bg-gradient-to-br from-blue-50 to-blue-100 border border-blue-200 rounded-lg p-4"
                >
                  <h3 className="font-semibold text-gray-900 mb-2">{service.name}</h3>
                  <div className="space-y-1 text-sm">
                    <p className="text-gray-600">
                      Середній час: <span className="font-semibold">{service.avg_days.toFixed(1)} днів</span>
                    </p>
                    <p className="text-gray-600">
                      Звернень: <span className="font-semibold">{service.total_appeals}</span>
                    </p>
                  </div>
                </div>
              ))}
            </div>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={topServices}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="name"
                  angle={-45}
                  textAnchor="end"
                  height={120}
                  interval={0}
                />
                <YAxis />
                <Tooltip formatter={(value: number) => `${value.toFixed(1)} днів`} />
                <Bar dataKey="avg_days" fill="#3b82f6" name="Середній час (дні)" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Monthly Trend */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center space-x-2">
            <TrendingUp className="h-5 w-5 text-primary" />
            <span>Динаміка звернень по місяцях</span>
          </h2>
          {monthlyTrend.length === 0 ? (
            <p className="text-gray-500 text-center py-8">Немає даних</p>
          ) : (
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
                  stroke="#3b82f6"
                  strokeWidth={2}
                  name="Кількість звернень"
                />
              </LineChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Day of Week Heat Map */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center space-x-2">
            <Calendar className="h-5 w-5 text-primary" />
            <span>Активність по днях тижня</span>
          </h2>
          {dayOfWeekStats.length === 0 ? (
            <p className="text-gray-500 text-center py-8">Немає даних</p>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={dayOfWeekStats}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="count" fill="#10b981" name="Кількість звернень" />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>
    </div>
  )
}

