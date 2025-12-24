import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { appealsAPI, userServicesAPI, categoriesAPI, servicesAPI } from '../lib/api'
import { format, differenceInDays, addDays } from 'date-fns'
import { uk } from 'date-fns/locale'
import { useAuth } from '../contexts/AuthContext'
import { useState, useEffect } from 'react'
import { Filter, ArrowUpDown, ArrowUp, ArrowDown, X, List, Grid, Building2 } from 'lucide-react'

const statusColors: Record<string, string> = {
  new: 'bg-blue-100 text-blue-800',
  assigned: 'bg-yellow-100 text-yellow-800',
  in_progress: 'bg-purple-100 text-purple-800',
  completed: 'bg-green-100 text-green-800',
  closed: 'bg-gray-100 text-gray-800',
  rejected: 'bg-red-100 text-red-800',
}

const statusLabels: Record<string, string> = {
  new: 'Нове',
  assigned: 'Призначене',
  in_progress: 'В роботі',
  completed: 'Виконане',
  closed: 'Закрите',
  rejected: 'Відхилене',
}

export default function AppealsPage() {
  const { user } = useAuth()
  const [viewMode, setViewMode] = useState<'all' | 'my'>('all')
  const [sortBy, setSortBy] = useState<'created_at' | 'status' | 'priority'>('created_at')
  const [sortOrder, setSortOrder] = useState<'desc' | 'asc'>('desc')
  const [selectedStatus, setSelectedStatus] = useState<string | null>(null)
  const [selectedCategory, setSelectedCategory] = useState<number | null>(null)
  const [selectedService, setSelectedService] = useState<number | null>(null)
  const [showFilters, setShowFilters] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)
  
  // For executor: show all appeals or only assigned ones
  const isExecutor = user?.role === 'executor'
  const isAdminOrDispatcher = user?.role === 'admin' || user?.role === 'dispatcher'
  const isCitizen = user?.role === 'citizen' || user?.role === 'user'
  
  // Get executor's services for filtering
  const { data: executorServicesData } = useQuery({
    queryKey: ['executor-services', user?.id],
    queryFn: async () => {
      if (!user?.id) return { services: [], serviceIds: [] }
      // Get services assigned to current user
      const response = await userServicesAPI.getMyServices()
      const services = response.data || []
      // Extract service IDs and keep full service data
      const serviceIds = services.map((service: any) => service.id).filter(Boolean)
      return { services, serviceIds }
    },
    enabled: !!user && isExecutor, // Fetch when executor is logged in
  })

  const executorServices = executorServicesData?.serviceIds || []
  const executorServicesList = executorServicesData?.services || []
  
  // Отримуємо категорії для фільтрації (з кешуванням)
  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      const response = await categoriesAPI.list()
      return Array.isArray(response.data) ? response.data : response.data?.items || []
    },
    staleTime: 5 * 60 * 1000, // Кеш на 5 хвилин
    cacheTime: 10 * 60 * 1000, // Зберігати в кеші 10 хвилин
  })

  // Отримуємо служби для фільтрації (тільки для адмінів та диспетчерів, з кешуванням)
  const { data: servicesData } = useQuery({
    queryKey: ['services'],
    queryFn: async () => {
      const response = await servicesAPI.list()
      return Array.isArray(response.data) ? response.data : response.data?.items || []
    },
    enabled: isAdminOrDispatcher,
    staleTime: 5 * 60 * 1000, // Кеш на 5 хвилин
    cacheTime: 10 * 60 * 1000, // Зберігати в кеші 10 хвилин
  })

  const { data, isLoading, error } = useQuery({
    queryKey: ['appeals', viewMode, user?.id, executorServices, sortBy, sortOrder, selectedStatus, selectedCategory, selectedService, currentPage],
    queryFn: async () => {
      const params: any = {
        page: currentPage,
        limit: 20,
        sort_by: sortBy,
        sort_order: sortOrder,
      }
      
      // Фільтрація на бекенді
      if (selectedStatus) {
        params.status = selectedStatus
      }
      if (selectedCategory) {
        params.category_id = selectedCategory
      }
      if (selectedService) {
        params.service_id = selectedService
      }
      
      // If executor viewing "my" appeals, filter by service_id
      if (isExecutor && viewMode === 'my' && executorServices && executorServices.length > 0) {
        // Використовуємо перший service_id для фільтрації
        params.service_id = executorServices[0]
      }
      
      // If citizen viewing "my" appeals, filter by user_id
      if (isCitizen && viewMode === 'my' && user?.id) {
        params.user_id = user.id
      }
      
      const response = await appealsAPI.list(params)
      return response.data
    },
    enabled: !!user && (
      (!isExecutor && !isCitizen) || // Admin/Dispatcher - always enabled
      (isExecutor && (viewMode === 'all' || (viewMode === 'my' && executorServices !== undefined))) || // Executor
      (isCitizen) // Citizen - always enabled (can view all or my)
    ),
    staleTime: 30 * 1000, // Кеш на 30 секунд
  })

  // Скидаємо сторінку при зміні фільтрів
  useEffect(() => {
    setCurrentPage(1)
  }, [selectedStatus, selectedCategory, selectedService, viewMode])
  
  // Підрахунок активних фільтрів
  const activeFiltersCount = (selectedStatus ? 1 : 0) + (selectedCategory ? 1 : 0) + (selectedService ? 1 : 0)
  
  // Скидання всіх фільтрів
  const clearFilters = () => {
    setSelectedStatus(null)
    setSelectedCategory(null)
    setSelectedService(null)
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
        Помилка завантаження звернень
      </div>
    )
  }

  return (
    <div>
      <div className="mb-4 sm:mb-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-3 sm:mb-4 gap-3 sm:gap-0">
          <h1 className="text-xl sm:text-3xl font-bold text-gray-900">Звернення громадян</h1>
          <div className="flex items-center space-x-2 sm:space-x-3">
            {/* Кнопка фільтрів */}
            <button
              onClick={() => setShowFilters(!showFilters)}
              className={`relative flex items-center justify-center space-x-2 px-2 sm:px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                showFilters || activeFiltersCount > 0
                  ? 'bg-primary text-white shadow-md hover:bg-primary/90'
                  : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50 hover:border-gray-400'
              }`}
            >
              <Filter className="h-4 w-4" />
              <span className="hidden sm:inline">Фільтри</span>
              {activeFiltersCount > 0 && (
                <span className="absolute -top-1 -right-1 bg-red-500 text-white text-[10px] sm:text-xs font-bold rounded-full h-4 w-4 sm:h-5 sm:w-5 flex items-center justify-center">
                  {activeFiltersCount}
                </span>
              )}
            </button>

            {/* Сортування */}
            <div className="flex items-center space-x-1 sm:space-x-2 bg-white border border-gray-300 rounded-lg px-2 sm:px-3 py-1.5 sm:py-2 shadow-sm">
              <ArrowUpDown className="h-3.5 w-3.5 sm:h-4 sm:w-4 text-gray-500" />
              <select
                value={sortBy}
                onChange={(e) =>
                  setSortBy(e.target.value as 'created_at' | 'status' | 'priority')
                }
                className="border-0 bg-transparent text-xs sm:text-sm font-medium text-gray-700 focus:outline-none focus:ring-0 cursor-pointer"
              >
                <option value="created_at">Дата</option>
                <option value="status">Статус</option>
                <option value="priority">Пріоритет</option>
              </select>
              <button
                onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
                className="p-0.5 sm:p-1 hover:bg-gray-100 rounded transition-colors"
                title={sortOrder === 'asc' ? 'Спочатку старі' : 'Спочатку нові'}
              >
                {sortOrder === 'asc' ? (
                  <ArrowUp className="h-3.5 w-3.5 sm:h-4 sm:w-4 text-gray-600" />
                ) : (
                  <ArrowDown className="h-3.5 w-3.5 sm:h-4 sm:w-4 text-gray-600" />
                )}
              </button>
            </div>

            {/* Перемикач режиму перегляду для executor та громадян */}
            {(isExecutor || isCitizen) && (
              <div className="flex space-x-0.5 sm:space-x-1 bg-gray-100 p-0.5 sm:p-1 rounded-lg border border-gray-200">
                <button
                  onClick={() => setViewMode('all')}
                  className={`flex items-center space-x-1 px-2 sm:px-3 py-1 sm:py-1.5 rounded-md text-xs sm:text-sm font-medium transition-all duration-200 ${
                    viewMode === 'all'
                      ? 'bg-white text-gray-900 shadow-sm'
                      : 'text-gray-600 hover:text-gray-900'
                  }`}
                >
                  <List className="h-3.5 w-3.5 sm:h-4 sm:w-4" />
                  <span>Всі</span>
                </button>
                <button
                  onClick={() => setViewMode('my')}
                  className={`flex items-center space-x-1 px-2 sm:px-3 py-1 sm:py-1.5 rounded-md text-xs sm:text-sm font-medium transition-all duration-200 ${
                    viewMode === 'my'
                      ? 'bg-white text-gray-900 shadow-sm'
                      : 'text-gray-600 hover:text-gray-900'
                  }`}
                >
                  <Grid className="h-3.5 w-3.5 sm:h-4 sm:w-4" />
                  <span>Мої</span>
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Панель фільтрів */}
        {showFilters && (
          <div className="mb-4 bg-white border border-gray-200 rounded-lg shadow-sm p-4 animate-in slide-in-from-top-2 duration-200">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-semibold text-gray-900 flex items-center space-x-2">
                <Filter className="h-4 w-4" />
                <span>Фільтри</span>
              </h3>
              {activeFiltersCount > 0 && (
                <button
                  onClick={clearFilters}
                  className="flex items-center space-x-1 text-xs text-gray-600 hover:text-red-600 transition-colors"
                >
                  <X className="h-3 w-3" />
                  <span>Очистити всі</span>
                </button>
              )}
            </div>
            <div className={`grid grid-cols-1 md:grid-cols-2 ${isAdminOrDispatcher ? 'lg:grid-cols-3' : ''} gap-4`}>
              {/* Фільтр за статусом */}
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-2">
                  Статус
                </label>
                <div className="flex flex-wrap gap-2">
                  {Object.entries(statusLabels).map(([status, label]) => (
                    <button
                      key={status}
                      onClick={() => setSelectedStatus(selectedStatus === status ? null : status)}
                      className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200 ${
                        selectedStatus === status
                          ? `${statusColors[status]} shadow-md scale-105`
                          : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                      }`}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>

              {/* Фільтр за категорією */}
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-2">
                  Категорія
                </label>
                <select
                  value={selectedCategory || ''}
                  onChange={(e) => setSelectedCategory(e.target.value ? parseInt(e.target.value) : null)}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm bg-white focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                >
                  <option value="">Всі категорії</option>
                  {categoriesData?.map((category: any) => (
                    <option key={category.id} value={category.id}>
                      {category.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* Фільтр за службою (тільки для адмінів та диспетчерів) */}
              {isAdminOrDispatcher && (
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-2 flex items-center space-x-1">
                    <Building2 className="h-3 w-3" />
                    <span>Служба</span>
                  </label>
                  <select
                    value={selectedService || ''}
                    onChange={(e) => setSelectedService(e.target.value ? parseInt(e.target.value) : null)}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm bg-white focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
                  >
                    <option value="">Всі служби</option>
                    {servicesData?.filter((service: any) => service.is_active).map((service: any) => (
                      <option key={service.id} value={service.id}>
                        {service.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>
          </div>
        )}
        
        {/* Show executor's services when viewing "my" appeals */}
        {isExecutor && viewMode === 'my' && executorServicesList.length > 0 && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <p className="text-sm font-medium text-gray-700 mb-2">
              Ваші служби:
            </p>
            <div className="flex flex-wrap gap-2">
              {executorServicesList.map((service: any) => (
                <span
                  key={service.id}
                  className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm font-medium"
                >
                  {service.name}
                </span>
              ))}
            </div>
          </div>
        )}
        
        {isExecutor && viewMode === 'my' && executorServicesList.length === 0 && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <p className="text-sm text-yellow-700">
              Вам ще не призначено жодної служби. Зверніться до диспечера для призначення.
            </p>
          </div>
        )}
      </div>

      {data?.items && data.items.length > 0 ? (
        <>
        <div className="space-y-3 sm:space-y-4">
          {data.items.map((appeal) => (
            <Link
              key={appeal.id}
              to={`/appeals/${appeal.id}`}
              className="block bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-6 hover:shadow-md transition-shadow relative"
            >
              {/* Compact badge in top-right corner */}
              {appeal.status !== 'closed' && appeal.status !== 'rejected' && (() => {
                const createdDate = new Date(appeal.created_at)
                const deadlineDate = addDays(createdDate, 30)
                const now = new Date()
                const daysRemaining = differenceInDays(deadlineDate, now)
                const daysPassed = differenceInDays(now, createdDate)
                const progress = Math.min((daysPassed / 30) * 100, 100)

                const getStatusColor = () => {
                  if (daysRemaining <= 0) return 'bg-red-500'
                  if (daysRemaining <= 5) return 'bg-red-400'
                  if (daysRemaining <= 10) return 'bg-yellow-400'
                  return 'bg-green-500'
                }

                const getStatusBg = () => {
                  if (daysRemaining <= 0) return 'bg-red-50 border-red-200'
                  if (daysRemaining <= 5) return 'bg-red-50 border-red-200'
                  if (daysRemaining <= 10) return 'bg-yellow-50 border-yellow-200'
                  return 'bg-green-50 border-green-200'
                }

                const getStatusText = () => {
                  if (daysRemaining <= 0) return 'text-red-700'
                  if (daysRemaining <= 5) return 'text-red-700'
                  if (daysRemaining <= 10) return 'text-yellow-700'
                  return 'text-green-700'
                }

                return (
                  <div className={`absolute top-2 right-2 sm:top-3 sm:right-3 px-2 py-1 sm:px-3 sm:py-1.5 rounded-full border text-[10px] sm:text-xs flex items-center space-x-1 sm:space-x-2 shadow-sm ${getStatusBg()}`}>
                    <div className={`w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full ${getStatusColor()}`} />
                    <span className={`font-semibold ${getStatusText()}`}>
                      {daysRemaining > 0
                        ? `${daysRemaining}/30`
                        : `!`}
                    </span>
                    <div className="w-6 sm:w-10 bg-gray-200 rounded-full h-0.5 sm:h-1 overflow-hidden">
                      <div
                        className={`h-full ${getStatusColor()}`}
                        style={{ width: `${progress}%` }}
                      />
                    </div>
                  </div>
                )
              })()}

              <div className="flex items-start justify-between">
                <div className="flex-1 pr-12 sm:pr-8">
                  <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3 mb-2">
                    <h3 className="text-base sm:text-lg font-semibold text-gray-900 pr-8 sm:pr-0">{appeal.title}</h3>
                    <div className="flex items-center gap-2 flex-wrap">
                      <span
                        className={`px-2 py-0.5 sm:py-1 text-[10px] sm:text-xs font-medium rounded-full ${statusColors[appeal.status] || statusColors.new}`}
                      >
                        {statusLabels[appeal.status] || appeal.status}
                      </span>
                      <span className={`px-2 py-0.5 sm:py-1 text-[10px] sm:text-xs font-medium rounded-full ${
                        appeal.priority === 3 ? 'bg-red-100 text-red-800' :
                        appeal.priority === 2 ? 'bg-yellow-100 text-yellow-800' :
                        'bg-green-100 text-green-800'
                      }`}>
                        {appeal.priority === 3 ? 'Високий' : appeal.priority === 2 ? 'Середній' : 'Низький'}
                      </span>
                    </div>
                  </div>
                  <p className="text-sm sm:text-base text-gray-600 mb-2 sm:mb-3 line-clamp-2">{appeal.description}</p>
                  <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-0 sm:space-x-4 text-xs sm:text-sm text-gray-500">
                    <span className="truncate">{appeal.address}</span>
                    <span className="hidden sm:inline">
                      {format(new Date(appeal.created_at), 'dd MMMM yyyy', { locale: uk })}
                    </span>
                    <span className="sm:hidden">
                      {format(new Date(appeal.created_at), 'dd.MM.yyyy', { locale: uk })}
                    </span>
                    {appeal.category && (
                      <span className="text-primary">{appeal.category.name}</span>
                    )}
                    {appeal.user && (
                      <span className="text-gray-600 hidden sm:inline">
                        <span className="font-medium">Автор:</span>{' '}
                        {appeal.user.first_name} {appeal.user.last_name}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </Link>
          ))}
        </div>
        
        {/* Пагінація */}
        {data && data.total > 20 && (
          <div className="mt-6 flex items-center justify-between">
            <div className="text-sm text-gray-700">
              Показано {(currentPage - 1) * 20 + 1} - {Math.min(currentPage * 20, data.total)} з {data.total}
            </div>
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                disabled={currentPage === 1}
                className="px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Попередня
              </button>
              <span className="px-3 py-2 text-sm font-medium text-gray-700">
                Сторінка {currentPage} з {Math.ceil(data.total / 20)}
              </span>
              <button
                onClick={() => setCurrentPage(p => p + 1)}
                disabled={currentPage >= Math.ceil(data.total / 20)}
                className="px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Наступна
              </button>
            </div>
          </div>
        )}
        </>
      ) : (
        <div className="text-center py-12">
          <p className="text-gray-500">Немає звернень</p>
        </div>
      )}
    </div>
  )
}

