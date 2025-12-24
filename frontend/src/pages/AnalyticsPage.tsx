import { useQuery } from '@tanstack/react-query'
import { appealsAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { useState } from 'react'
import { format, subDays, startOfDay, endOfDay } from 'date-fns'
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

const statusLabels: Record<string, string> = {
  new: 'Нові',
  assigned: 'Призначені',
  in_progress: 'В роботі',
  completed: 'Виконані',
  closed: 'Закриті',
  rejected: 'Відхилені',
}

export default function AnalyticsPage() {
  const { user } = useAuth()
  const [dateRange, setDateRange] = useState<'7d' | '30d' | '90d' | 'all'>('30d')

  const getDateRange = () => {
    const now = new Date()
    switch (dateRange) {
      case '7d':
        return {
          from_date: startOfDay(subDays(now, 7)).toISOString(),
          to_date: endOfDay(now).toISOString(),
        }
      case '30d':
        return {
          from_date: startOfDay(subDays(now, 30)).toISOString(),
          to_date: endOfDay(now).toISOString(),
        }
      case '90d':
        return {
          from_date: startOfDay(subDays(now, 90)).toISOString(),
          to_date: endOfDay(now).toISOString(),
        }
      default:
        return {}
    }
  }

  const { data: stats, isLoading, error } = useQuery({
    queryKey: ['statistics', dateRange],
    queryFn: async () => {
      const params = getDateRange()
      const response = await appealsAPI.getStatistics(params)
      return response.data
    },
    enabled: (user?.role === 'dispatcher' || user?.role === 'admin') && !!user,
    staleTime: 60 * 1000, // Кеш на 1 хвилину
    cacheTime: 5 * 60 * 1000, // Зберігати в кеші 5 хвилин
  })

  if (!user || (user.role !== 'dispatcher' && user.role !== 'admin')) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        У вас немає доступу до аналітики
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
        Помилка завантаження статистики
      </div>
    )
  }

  // Prepare data for charts
  const statusData = stats?.by_status
    ? Object.entries(stats.by_status).map(([status, count]) => ({
        name: statusLabels[status] || status,
        value: count,
        status,
      }))
    : []

  const categoryData = stats?.by_category
    ? Object.entries(stats.by_category)
        .map(([name, count]) => ({ name, value: count }))
        .slice(0, 10)
    : []

  const serviceData = stats?.by_service
    ? Object.entries(stats.by_service)
        .map(([name, count]) => ({ name, value: count }))
        .slice(0, 10)
    : []

  const dailyTrendData = stats?.daily_trend || []

  const priorityData = stats?.by_priority
    ? Object.entries(stats.by_priority).map(([priority, count]) => ({
        name: `Пріоритет ${priority}`,
        value: count,
      }))
    : []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Аналітика та статистика</h1>
        <div className="flex space-x-2">
          <button
            onClick={() => setDateRange('7d')}
            className={`px-4 py-2 rounded-md text-sm font-medium ${
              dateRange === '7d'
                ? 'bg-primary text-white'
                : 'bg-white text-gray-700 hover:bg-gray-100'
            }`}
          >
            7 днів
          </button>
          <button
            onClick={() => setDateRange('30d')}
            className={`px-4 py-2 rounded-md text-sm font-medium ${
              dateRange === '30d'
                ? 'bg-primary text-white'
                : 'bg-white text-gray-700 hover:bg-gray-100'
            }`}
          >
            30 днів
          </button>
          <button
            onClick={() => setDateRange('90d')}
            className={`px-4 py-2 rounded-md text-sm font-medium ${
              dateRange === '90d'
                ? 'bg-primary text-white'
                : 'bg-white text-gray-700 hover:bg-gray-100'
            }`}
          >
            90 днів
          </button>
          <button
            onClick={() => setDateRange('all')}
            className={`px-4 py-2 rounded-md text-sm font-medium ${
              dateRange === 'all'
                ? 'bg-primary text-white'
                : 'bg-white text-gray-700 hover:bg-gray-100'
            }`}
          >
            Всі
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 className="text-sm font-medium text-gray-500 mb-2">Всього звернень</h3>
          <p className="text-3xl font-bold text-gray-900">{stats?.total || 0}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 className="text-sm font-medium text-gray-500 mb-2">Середній час обробки</h3>
          <p className="text-3xl font-bold text-gray-900">
            {stats?.avg_processing_days
              ? `${stats.avg_processing_days.toFixed(1)} дн.`
              : stats?.avg_processing_hours
              ? `${(stats.avg_processing_hours / 24).toFixed(1)} дн.`
              : '—'}
          </p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 className="text-sm font-medium text-gray-500 mb-2">Прострочені звернення</h3>
          <p className="text-3xl font-bold text-red-600">
            {stats?.overdue_count || 0}
          </p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 className="text-sm font-medium text-gray-500 mb-2">Виконано в термін</h3>
          <p className="text-3xl font-bold text-green-600">
            {stats?.on_time_percentage
              ? `${Math.round(stats.on_time_percentage)}%`
              : '—'}
          </p>
          {stats?.total_completed && (
            <p className="text-xs text-gray-500 mt-1">
              {stats.on_time_count} з {stats.total_completed} звернень
            </p>
          )}
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 className="text-sm font-medium text-gray-500 mb-2">Нові звернення</h3>
          <p className="text-3xl font-bold text-blue-600">
            {stats?.by_status?.new || 0}
          </p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 className="text-sm font-medium text-gray-500 mb-2">В роботі</h3>
          <p className="text-3xl font-bold text-purple-600">
            {stats?.by_status?.in_progress || 0}
          </p>
        </div>
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Status Distribution */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Розподіл по статусах</h2>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={statusData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
              >
                {statusData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>

        {/* Daily Trend */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Динаміка створення</h2>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={dailyTrendData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                dataKey="date"
                tickFormatter={(value) => format(new Date(value), 'dd.MM')}
              />
              <YAxis />
              <Tooltip
                labelFormatter={(value) => format(new Date(value), 'dd MMMM yyyy')}
              />
              <Legend />
              <Line
                type="monotone"
                dataKey="count"
                stroke="#3b82f6"
                strokeWidth={2}
                name="Кількість"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Top Categories */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Топ категорій</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={categoryData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" angle={-45} textAnchor="end" height={100} />
              <YAxis />
              <Tooltip />
              <Bar dataKey="value" fill="#3b82f6" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Top Services */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Топ служб</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={serviceData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" angle={-45} textAnchor="end" height={100} />
              <YAxis />
              <Tooltip />
              <Bar dataKey="value" fill="#10b981" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Priority Distribution */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Розподіл по пріоритетах</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={priorityData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Bar dataKey="value" fill="#ef4444" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}

