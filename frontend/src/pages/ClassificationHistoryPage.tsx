import { useQuery } from '@tanstack/react-query'
import { useAuth } from '../contexts/AuthContext'
import { classificationAPI } from '../lib/api'
import { useState, useMemo } from 'react'
import { 
  History, 
  Search, 
  ChevronLeft, 
  ChevronRight,
  TrendingUp,
  ChevronDown,
  ChevronUp
} from 'lucide-react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
} from 'recharts'

export default function ClassificationHistoryPage() {
  const { user } = useAuth()
  const [page, setPage] = useState(1)
  const [limit] = useState(50)
  const [serviceFilter, setServiceFilter] = useState<string>('')
  const [moderationFilter, setModerationFilter] = useState<boolean | null>(null)
  const [searchText, setSearchText] = useState('')
  const [chartsExpanded, setChartsExpanded] = useState(false) // По дефолту згорнуті

  const { data, isLoading, error } = useQuery({
    queryKey: ['classification-history', page, limit, serviceFilter, moderationFilter],
    queryFn: async () => {
      const params: any = {
        page,
        limit,
      }
      if (serviceFilter) {
        params.service = serviceFilter
      }
      if (moderationFilter !== null) {
        params.needs_moderation = moderationFilter
      }
      return await classificationAPI.getHistory(params)
    },
    enabled: user?.role === 'admin' && !!user,
    staleTime: 10 * 1000, // 10 секунд
  })

  // Окремий запит для графіка (всі записи для аналізу)
  const { data: chartData, isLoading: isLoadingChart } = useQuery({
    queryKey: ['classification-history-chart', serviceFilter, moderationFilter],
    queryFn: async () => {
      const params: any = {
        page: 1,
        limit: 50, // Ліміт для графіка
      }
      if (serviceFilter) {
        params.service = serviceFilter
      }
      if (moderationFilter !== null) {
        params.needs_moderation = moderationFilter
      }
      return await classificationAPI.getHistory(params)
    },
    enabled: user?.role === 'admin' && !!user,
    staleTime: 30 * 1000, // 30 секунд
  })

  // Використовуємо дані для графіка: спочатку chartData, якщо немає - то data
  const dataForChart = chartData || data

  // Фільтрація пошуку на клієнті (по тексту звернення)
  const filteredEntries = data?.entries.filter(entry => {
    if (!searchText) return true
    return entry.text.toLowerCase().includes(searchText.toLowerCase())
  }) || []

  // Підготовка даних для графіка впевненості по часу
  const confidenceChartData = useMemo(() => {
    if (!dataForChart?.entries || dataForChart.entries.length === 0) return []
    
    // Сортуємо по часу та конвертуємо впевненість у відсотки
    const sorted = [...dataForChart.entries]
      .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
      .map(entry => {
        // Беремо другий за впевненістю коефіцієнт з топ альтернатив
        const secondConfidence = entry.top_alternatives && entry.top_alternatives.length > 1
          ? parseFloat((entry.top_alternatives[1].confidence * 100).toFixed(1))
          : null
        
        // Обчислюємо gap (різниця між першою та другою впевненістю)
        const gap = secondConfidence !== null
          ? parseFloat((entry.confidence * 100 - secondConfidence).toFixed(1))
          : null
        
        return {
          time: new Date(entry.timestamp).toLocaleString('uk-UA', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
          }),
          timestamp: new Date(entry.timestamp).getTime(),
          confidence: parseFloat((entry.confidence * 100).toFixed(1)),
          secondConfidence: secondConfidence,
          gap: gap,
          service: entry.service,
        }
      })
    
    return sorted
  }, [dataForChart])

  // Підготовка даних для гістограми розподілу впевненості
  const confidenceDistribution = useMemo(() => {
    if (!dataForChart?.entries || dataForChart.entries.length === 0) return []
    
    // Розбиваємо на інтервали з кроком 2.5%: 75-77.5%, 77.5-80%, 80-82.5%, ...
    const bins = {
      '75-77.5%': 0,
      '77.5-80%': 0,
      '80-82.5%': 0,
      '82.5-85%': 0,
      '85-87.5%': 0,
      '87.5-90%': 0,
      '90-92.5%': 0,
      '92.5-95%': 0,
      '95-97.5%': 0,
      '97.5-100%': 0,
    }
    
    dataForChart.entries.forEach(entry => {
      const percent = entry.confidence * 100
      if (percent < 75) {
        // Значення нижче 75% додаємо до першого інтервалу
        bins['75-77.5%']++
      } else if (percent < 77.5) bins['75-77.5%']++
      else if (percent < 80) bins['77.5-80%']++
      else if (percent < 82.5) bins['80-82.5%']++
      else if (percent < 85) bins['82.5-85%']++
      else if (percent < 87.5) bins['85-87.5%']++
      else if (percent < 90) bins['87.5-90%']++
      else if (percent < 92.5) bins['90-92.5%']++
      else if (percent < 95) bins['92.5-95%']++
      else if (percent < 97.5) bins['95-97.5%']++
      else bins['97.5-100%']++
    })
    
    // Фільтруємо інтервали, які мають дані (count > 0)
    return Object.entries(bins)
      .filter(([_, count]) => count > 0)
      .map(([range, count]) => ({
        range,
        count,
      }))
  }, [dataForChart])

  if (!user || user.role !== 'admin') {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        <p className="font-semibold">У вас немає доступу до цієї сторінки</p>
        <p className="text-sm mt-1">Потрібна роль: <strong>admin</strong></p>
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
        <p className="font-semibold">Помилка завантаження історії класифікацій</p>
        <p className="text-sm mt-1">Перевірте, чи запущено сервіс класифікації</p>
      </div>
    )
  }

  const totalPages = data ? Math.ceil(data.total / limit) : 1

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <History className="h-8 w-8 text-primary" />
          <h1 className="text-3xl font-bold text-gray-900">Історія класифікацій</h1>
        </div>
        {data && (
          <div className="text-sm text-gray-600">
            Всього записів: <span className="font-semibold">{data.total}</span>
          </div>
        )}
      </div>

      {/* Фільтри */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Пошук по тексту */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
            <input
              type="text"
              placeholder="Пошук по тексту звернення..."
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
            />
          </div>

          {/* Фільтр за службою */}
          <div>
            <input
              type="text"
              placeholder="Фільтр за службою..."
              value={serviceFilter}
              onChange={(e) => {
                setServiceFilter(e.target.value)
                setPage(1)
              }}
              className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
            />
          </div>

          {/* Фільтр за модерацією */}
          <div>
            <select
              value={moderationFilter === null ? '' : moderationFilter.toString()}
              onChange={(e) => {
                const value = e.target.value
                setModerationFilter(value === '' ? null : value === 'true')
                setPage(1)
              }}
              className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
            >
              <option value="">Всі</option>
              <option value="true">Потребує модерації</option>
              <option value="false">Автоматично</option>
            </select>
          </div>
        </div>
      </div>

      {/* Графіки впевненості */}
      {data && data.entries.length > 0 && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200">
          {/* Заголовок з кнопкою згортання */}
          <button
            onClick={() => setChartsExpanded(!chartsExpanded)}
            className="w-full px-6 py-4 flex items-center justify-between hover:bg-gray-50 transition-colors"
          >
            <h2 className="text-lg font-semibold text-gray-900 flex items-center">
              <TrendingUp className="h-5 w-5 mr-2" />
              Графіки впевненості
            </h2>
            {chartsExpanded ? (
              <ChevronUp className="h-5 w-5 text-gray-500" />
            ) : (
              <ChevronDown className="h-5 w-5 text-gray-500" />
            )}
          </button>
          
          {/* Контент графіків */}
          {chartsExpanded && (
            <div className="px-6 pb-6 space-y-6">
              {/* Лінійний графік впевненості по часу */}
              <div>
                <h3 className="text-md font-semibold text-gray-900 mb-4">
                  Історія впевненості
                </h3>
            {isLoadingChart ? (
              <div className="flex items-center justify-center h-[300px]">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              </div>
            ) : confidenceChartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={confidenceChartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis 
                    dataKey="time" 
                    angle={-45}
                    textAnchor="end"
                    height={80}
                    interval="preserveStartEnd"
                  />
                  <YAxis 
                    domain={[75, 100]}
                    label={{ value: 'Впевненість (%)', angle: -90, position: 'insideLeft' }}
                  />
                  <Tooltip 
                    formatter={(value: any, name: string, props: any) => {
                      if (value === null || value === undefined) return ['—', name]
                      const formatted = [`${value}%`, name]
                      
                      // Додаємо gap до tooltip, якщо він є
                      if (props.payload && props.payload[0] && props.payload[0].payload.gap !== null && props.payload[0].payload.gap !== undefined) {
                        // Gap вже буде показано через окремий рядок
                        return formatted
                      }
                      
                      return formatted
                    }}
                    labelFormatter={(label) => `Час: ${label}`}
                    content={(props: any) => {
                      if (!props.active || !props.payload || props.payload.length === 0) return null
                      
                      const data = props.payload[0].payload
                      const gap = data.gap
                      
                      return (
                        <div className="bg-white border border-gray-200 rounded-lg shadow-lg p-3">
                          <p className="font-semibold text-gray-900 mb-2">{props.label}</p>
                          {props.payload.map((entry: any, index: number) => (
                            <p key={index} className="text-sm" style={{ color: entry.color }}>
                              {entry.name}: {entry.value !== null && entry.value !== undefined ? `${entry.value}%` : '—'}
                            </p>
                          ))}
                          {gap !== null && gap !== undefined && (
                            <p className="text-sm font-semibold text-gray-700 mt-2 pt-2 border-t border-gray-200">
                              Gap: {gap}%
                            </p>
                          )}
                        </div>
                      )
                    }}
                  />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="confidence"
                    stroke="#3B82F6"
                    strokeWidth={2}
                    dot={{ r: 3 }}
                    name="Впевненість (%)"
                  />
                  <Line
                    type="monotone"
                    dataKey="secondConfidence"
                    stroke="#EF4444"
                    strokeWidth={2}
                    dot={{ r: 3 }}
                    name="Друга впевненість (%)"
                  />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <p className="text-gray-500 text-center py-8">Немає даних для графіка</p>
            )}
              </div>

              {/* Гістограма розподілу впевненості */}
              <div>
                <h3 className="text-md font-semibold text-gray-900 mb-4">
                  Розподіл впевненості
                </h3>
            {isLoadingChart ? (
              <div className="flex items-center justify-center h-[300px]">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              </div>
            ) : confidenceDistribution.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={confidenceDistribution}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="range" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="count" fill="#10B981" name="Кількість звернень" />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <p className="text-gray-500 text-center py-8">Немає даних для графіка</p>
            )}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Таблиця історії */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Текст звернення
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Служба
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Впевненість
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Топ альтернативи
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredEntries.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-6 py-8 text-center text-gray-500">
                    Немає записів в історії
                  </td>
                </tr>
              ) : (
                filteredEntries.map((entry) => (
                  <tr key={entry.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 text-sm text-gray-900">
                      <div className="break-words whitespace-normal">
                        {entry.text}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
                        {entry.service}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <span className="text-gray-700 font-medium">
                        {(entry.confidence * 100).toFixed(1)}%
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {entry.top_alternatives && entry.top_alternatives.length > 0 ? (
                        <div className="flex flex-wrap gap-2">
                          {entry.top_alternatives.map((alt, idx) => (
                            <span
                              key={idx}
                              className="px-2 py-1 text-xs rounded-full bg-gray-100 text-gray-700"
                              title={`Впевненість: ${(alt.confidence * 100).toFixed(1)}%`}
                            >
                              {alt.service} ({(alt.confidence * 100).toFixed(1)}%)
                            </span>
                          ))}
                        </div>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Пагінація */}
        {data && data.total > limit && (
          <div className="bg-gray-50 px-6 py-4 border-t border-gray-200 flex items-center justify-between">
            <div className="text-sm text-gray-700">
              Показано {(page - 1) * limit + 1} - {Math.min(page * limit, data.total)} з {data.total}
            </div>
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="p-2 rounded-md border border-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100"
              >
                <ChevronLeft className="h-5 w-5" />
              </button>
              <span className="text-sm text-gray-700">
                Сторінка {page} з {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="p-2 rounded-md border border-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100"
              >
                <ChevronRight className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

