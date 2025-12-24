import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { appealsAPI, photosAPI, servicesAPI, commentsAPI } from '../lib/api'
import { format, differenceInDays, addDays } from 'date-fns'
import { uk } from 'date-fns/locale'
import { useAuth } from '../contexts/AuthContext'
import MapView from '../components/MapView'
import PhotoUpload from '../components/PhotoUpload'
import { useState, useEffect } from 'react'
import { MessageSquare, Edit, Trash2, Send, X, Upload, Image, History, ChevronDown, ChevronUp } from 'lucide-react'
import type { Comment } from '../types'

export default function AppealDetailPage() {
  const { id } = useParams<{ id: string }>()
  const appealId = parseInt(id || '0')
  const { user } = useAuth()

  const { data: appeal, isLoading } = useQuery({
    queryKey: ['appeal', appealId],
    queryFn: async () => {
      const response = await appealsAPI.getById(appealId)
      return response.data
    },
    enabled: !!appealId && !!user,
  })

  const queryClient = useQueryClient()

  const statusMutation = useMutation({
    mutationFn: ({ status, comment }: { status: string; comment?: string }) =>
      appealsAPI.updateStatus(appealId, status, comment),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appeal', appealId] })
      queryClient.invalidateQueries({ queryKey: ['appeals'] })
      queryClient.invalidateQueries({ queryKey: ['appeal-history', appealId] })
      // Refetch history if panel is visible
      if (isHistoryVisible) {
        refetchHistory()
      }
      setIsEditingStatus(false)
      setStatusComment('')
    },
  })

  const priorityMutation = useMutation({
    mutationFn: (priority: number) => appealsAPI.updatePriority(appealId, priority),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appeal', appealId] })
      queryClient.invalidateQueries({ queryKey: ['appeals'] })
      setIsEditingPriority(false)
    },
  })

  const assignMutation = useMutation({
    mutationFn: ({ serviceId, priority }: { serviceId: number; priority?: number }) =>
      appealsAPI.assign(appealId, serviceId, priority),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appeal', appealId] })
      queryClient.invalidateQueries({ queryKey: ['appeals'] })
      setIsAssigning(false)
    },
  })

  const { data: services } = useQuery({
    queryKey: ['services'],
    queryFn: async () => {
      const response = await servicesAPI.list()
      return response.data || []
    },
    enabled: (user?.role === 'executor' || user?.role === 'dispatcher' || user?.role === 'admin'),
  })

  const [selectedServiceId, setSelectedServiceId] = useState<number>(0)
  const [selectedPriority, setSelectedPriority] = useState<number>(2)
  const [isEditingStatus, setIsEditingStatus] = useState<boolean>(false)
  const [isEditingPriority, setIsEditingPriority] = useState<boolean>(false)
  const [selectedStatus, setSelectedStatus] = useState<string>('')
  const [statusComment, setStatusComment] = useState<string>('')
  const [newCommentText, setNewCommentText] = useState<string>('')
  const [isInternalComment, setIsInternalComment] = useState<boolean>(false)
  const [editingCommentId, setEditingCommentId] = useState<number | null>(null)
  const [editCommentText, setEditCommentText] = useState<string>('')
  const [isAssigning, setIsAssigning] = useState<boolean>(false)
  const [isUploadingPhotos, setIsUploadingPhotos] = useState<boolean>(false)
  const [isHistoryVisible, setIsHistoryVisible] = useState<boolean>(false)

  // Update selected values when appeal loads
  useEffect(() => {
    if (appeal) {
      setSelectedServiceId(appeal.service_id || 0)
      setSelectedPriority(appeal.priority || 2)
      setSelectedStatus(appeal.status)
    }
  }, [appeal])

  const { data: photos, refetch: refetchPhotos } = useQuery({
    queryKey: ['photos', appealId],
    queryFn: async () => {
      const response = await photosAPI.list(appealId)
      return response.data || []
    },
    enabled: !!appealId && !!user,
  })

  const photoMutation = useMutation({
    mutationFn: (files: File[]) => photosAPI.upload(appealId, files),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['photos', appealId] })
      refetchPhotos()
    },
  })

  // Comments
  // History
  const { data: history, isLoading: isLoadingHistory, refetch: refetchHistory } = useQuery({
    queryKey: ['appeal-history', appealId],
    queryFn: async () => {
      const response = await appealsAPI.getHistory(appealId)
      return response.data || []
    },
    enabled: !!appealId && !!user && isHistoryVisible,
  })

  const { data: comments, refetch: refetchComments } = useQuery({
    queryKey: ['comments', appealId],
    queryFn: async () => {
      const response = await commentsAPI.getByAppealId(appealId)
      return response.data || []
    },
    enabled: !!appealId && !!user,
  })

  const createCommentMutation = useMutation({
    mutationFn: (text: string) => commentsAPI.create(appealId, text, isInternalComment),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', appealId] })
      setNewCommentText('')
      setIsInternalComment(false)
    },
  })

  const updateCommentMutation = useMutation({
    mutationFn: ({ id, text }: { id: number; text: string }) => 
      commentsAPI.update(id, text, isInternalComment),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', appealId] })
      setEditingCommentId(null)
      setEditCommentText('')
    },
  })

  const deleteCommentMutation = useMutation({
    mutationFn: (id: number) => commentsAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', appealId] })
    },
  })

  // Check if user can upload photos
  const canUploadPhotos = user && appeal && (
    (user.role === 'citizen' && appeal.user_id === user.id) ||
    (user.role === 'executor' && appeal.service_id) ||
    user.role === 'dispatcher' ||
    user.role === 'admin'
  )

  // Check if user can change status
  const canChangeStatus = user && appeal && (
    user.role === 'dispatcher' ||
    user.role === 'admin' ||
    (user.role === 'executor' && appeal.service_id)
  )

  // Users who can change priority (dispatcher/admin/executor)
  const canChangePriority = user && appeal && (
    user.role === 'dispatcher' ||
    user.role === 'admin' ||
    user.role === 'executor'
  )

  const statusLabels: Record<string, string> = {
    new: 'Нове',
    assigned: 'Призначене',
    in_progress: 'В роботі',
    completed: 'Виконане',
    closed: 'Закрите',
    rejected: 'Відхилене',
  }

  const statusColors: Record<string, string> = {
    new: 'bg-blue-100 text-blue-800',
    assigned: 'bg-yellow-100 text-yellow-800',
    in_progress: 'bg-purple-100 text-purple-800',
    completed: 'bg-green-100 text-green-800',
    closed: 'bg-gray-100 text-gray-800',
    rejected: 'bg-red-100 text-red-800',
  }

  const statusOptions = [
    { value: 'new', label: 'Нове' },
    { value: 'assigned', label: 'Призначене' },
    { value: 'in_progress', label: 'В роботі' },
    { value: 'completed', label: 'Виконане' },
    { value: 'closed', label: 'Закрите' },
    { value: 'rejected', label: 'Відхилене' },
  ]

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (!appeal) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        Звернення не знайдено
      </div>
    )
  }

  return (
    <div className="space-y-4 sm:space-y-6">
      <Link to="/" className="inline-flex items-center text-sm sm:text-base text-primary hover:text-primary/80">
        ← Назад до списку
      </Link>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-6">
        <div className="flex items-start justify-between mb-3 sm:mb-4">
          <div className="flex-1">
            <h1 className="text-xl sm:text-3xl font-bold text-gray-900 mb-2 sm:mb-0">{appeal.title}</h1>
            <div className="flex flex-wrap items-center gap-2 mt-2">
              <span className={`px-2 sm:px-3 py-1 text-xs sm:text-sm font-medium rounded-full ${
                appeal.status === 'new' ? 'bg-blue-100 text-blue-800' :
                appeal.status === 'assigned' ? 'bg-yellow-100 text-yellow-800' :
                appeal.status === 'in_progress' ? 'bg-purple-100 text-purple-800' :
                appeal.status === 'completed' ? 'bg-green-100 text-green-800' :
                appeal.status === 'closed' ? 'bg-gray-100 text-gray-800' :
                'bg-red-100 text-red-800'
              }`}>
                {statusOptions.find(s => s.value === appeal.status)?.label || appeal.status}
              </span>
              <span className={`px-2 sm:px-3 py-1 text-xs sm:text-sm font-medium rounded-full ${
                appeal.priority === 3 ? 'bg-red-100 text-red-800' :
                appeal.priority === 2 ? 'bg-yellow-100 text-yellow-800' :
                'bg-green-100 text-green-800'
              }`}>
                {appeal.priority === 3 ? 'Високий' : appeal.priority === 2 ? 'Середній' : 'Низький'}
              </span>
            </div>
            {appeal.user && (
              <div className="mt-2 text-xs sm:text-sm text-gray-600">
                <span className="font-medium">Автор:</span>{' '}
                <span>{appeal.user.first_name} {appeal.user.last_name}</span>
              </div>
            )}
          </div>
        </div>
        <p className="text-sm sm:text-base text-gray-700 mb-4 sm:mb-6">{appeal.description}</p>

        {/* Processing time indicator - 30 days limit */}
        {appeal.status !== 'closed' && appeal.status !== 'rejected' && (() => {
          const createdDate = new Date(appeal.created_at)
          const deadlineDate = addDays(createdDate, 30)
          const now = new Date()
          const daysPassed = differenceInDays(now, createdDate)
          const daysRemaining = differenceInDays(deadlineDate, now)
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
            <div className={`mb-4 sm:mb-6 p-3 sm:p-4 rounded-lg border ${getStatusBg()}`}>
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 sm:gap-0 mb-2">
                <div className="flex flex-wrap items-center gap-2">
                  <span className={`text-xs sm:text-sm font-medium ${getStatusText()}`}>
                    Термін обробки
                  </span>
                  {daysRemaining <= 0 && (
                    <span className="px-2 py-0.5 sm:py-1 text-[10px] sm:text-xs font-bold bg-red-600 text-white rounded">
                      ТЕРМІН ПРОЙШОВ!
                    </span>
                  )}
                  {daysRemaining > 0 && daysRemaining <= 5 && (
                    <span className="px-2 py-0.5 sm:py-1 text-[10px] sm:text-xs font-bold bg-red-500 text-white rounded">
                      КРИТИЧНО!
                    </span>
                  )}
                  {daysRemaining > 5 && daysRemaining <= 10 && (
                    <span className="px-2 py-0.5 sm:py-1 text-[10px] sm:text-xs font-bold bg-yellow-500 text-white rounded">
                      УВАГА!
                    </span>
                  )}
                </div>
                <div className="text-left sm:text-right">
                  <p className={`text-xs sm:text-sm font-semibold ${getStatusText()}`}>
                    {daysRemaining > 0 ? (
                      <>Залишилось: <span className="text-base sm:text-lg">{daysRemaining}</span> {daysRemaining === 1 ? 'день' : daysRemaining < 5 ? 'дні' : 'днів'}</>
                    ) : (
                      <>Прострочено на: <span className="text-base sm:text-lg">{Math.abs(daysRemaining)}</span> {Math.abs(daysRemaining) === 1 ? 'день' : Math.abs(daysRemaining) < 5 ? 'дні' : 'днів'}</>
                    )}
                  </p>
                  <p className="text-[10px] sm:text-xs text-gray-500 mt-0.5 sm:mt-1">
                    Пройшло: {daysPassed} {daysPassed === 1 ? 'день' : daysPassed < 5 ? 'дні' : 'днів'} з 30
                  </p>
                </div>
              </div>
              
              {/* Progress bar */}
              <div className="w-full bg-gray-200 rounded-full h-2 sm:h-3 overflow-hidden">
                <div
                  className={`h-full transition-all duration-300 ${getStatusColor()}`}
                  style={{ width: `${progress}%` }}
                />
              </div>
              
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 sm:gap-0 mt-2 text-[10px] sm:text-xs text-gray-600">
                <span>Створено: {format(createdDate, 'dd.MM.yyyy', { locale: uk })}</span>
                <span className="hidden sm:inline">Термін: {format(deadlineDate, 'dd MMMM yyyy', { locale: uk })}</span>
                <span className="sm:hidden">Термін: {format(deadlineDate, 'dd.MM.yyyy', { locale: uk })}</span>
              </div>
            </div>
          )
        })()}

        {/* Executors can see appeals assigned to their services */}
        {user?.role === 'executor' && !appeal.service_id && (
          <div className="mb-6 p-4 bg-yellow-50 rounded-lg border border-yellow-200">
            <p className="text-sm text-yellow-700">
              Це звернення ще не призначене до служби. Зверніться до диспечера для призначення.
            </p>
          </div>
        )}

        {/* Status change for dispatcher/executor/admin */}
        {canChangeStatus && (
          <div className="mb-4 sm:mb-6 p-3 sm:p-4 bg-gray-50 rounded-lg">
            {!isEditingStatus ? (
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 sm:gap-0">
                <div>
                  <h3 className="text-xs sm:text-sm font-medium text-gray-700 mb-1">Статус</h3>
                  <p className="text-sm sm:text-base text-gray-900 font-medium">
                    {statusOptions.find(opt => opt.value === appeal.status)?.label || appeal.status}
                  </p>
                </div>
                <button
                  onClick={() => setIsEditingStatus(true)}
                  className="px-3 sm:px-4 py-1.5 sm:py-2 bg-primary text-white rounded-md hover:bg-primary/90 text-xs sm:text-sm font-medium"
                >
                  Змінити статус
                </button>
              </div>
            ) : (
              <div className="space-y-4">
                <h3 className="text-sm font-medium text-gray-700">Змінити статус</h3>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Новий статус
                  </label>
                  <select
                    value={selectedStatus}
                    onChange={(e) => setSelectedStatus(e.target.value)}
                    disabled={statusMutation.isPending}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-primary focus:border-primary"
                  >
                    {statusOptions.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Коментар (опціонально)
                  </label>
                  <textarea
                    value={statusComment}
                    onChange={(e) => setStatusComment(e.target.value)}
                    placeholder="Додайте коментар до зміни статусу..."
                    disabled={statusMutation.isPending}
                    rows={3}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-primary focus:border-primary resize-none"
                  />
                </div>
                <div className="flex items-center space-x-2">
                  <button
                    onClick={() => {
                      if (selectedStatus !== appeal.status || statusComment.trim()) {
                        statusMutation.mutate({ 
                          status: selectedStatus, 
                          comment: statusComment.trim() || undefined 
                        })
                      } else {
                        setIsEditingStatus(false)
                      }
                    }}
                    disabled={statusMutation.isPending || selectedStatus === appeal.status}
                    className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium"
                  >
                    {statusMutation.isPending ? 'Збереження...' : 'Зберегти'}
                  </button>
                  <button
                    onClick={() => {
                      setIsEditingStatus(false)
                      setSelectedStatus(appeal.status)
                      setStatusComment('')
                    }}
                    disabled={statusMutation.isPending}
                    className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium"
                  >
                    Скасувати
                  </button>
                </div>
                {statusMutation.isError && (
                  <p className="text-sm text-red-600">
                    Помилка: {statusMutation.error instanceof Error ? statusMutation.error.message : 'Невідома помилка'}
                  </p>
                )}
              </div>
            )}
          </div>
        )}

        {/* Priority change for dispatcher/executor/admin when service is assigned */}
        {canChangePriority && (
          <div className="mb-4 sm:mb-6 p-3 sm:p-4 bg-gray-50 rounded-lg">
            {!isEditingPriority ? (
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 sm:gap-0">
                <div>
                  <h3 className="text-xs sm:text-sm font-medium text-gray-700 mb-1">Пріоритет</h3>
                  <p className="text-sm sm:text-base text-gray-900 font-medium">
                    {appeal.priority === 3
                      ? 'Високий'
                      : appeal.priority === 2
                      ? 'Середній'
                      : 'Низький'}
                  </p>
                </div>
                <button
                  onClick={() => {
                    setSelectedPriority(appeal.priority || 2)
                    setIsEditingPriority(true)
                  }}
                  className="px-3 sm:px-4 py-1.5 sm:py-2 bg-primary text-white rounded-md hover:bg-primary/90 text-xs sm:text-sm font-medium"
                >
                  Змінити пріоритет
                </button>
              </div>
            ) : (
              <div className="space-y-4">
                <h3 className="text-sm font-medium text-gray-700">Змінити пріоритет</h3>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Новий пріоритет
                  </label>
                  <select
                    value={selectedPriority}
                    onChange={(e) => setSelectedPriority(Number(e.target.value))}
                    disabled={priorityMutation.isPending}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-primary focus:border-primary"
                  >
                    <option value={1}>Низький</option>
                    <option value={2}>Середній</option>
                    <option value={3}>Високий</option>
                  </select>
                </div>
                <div className="flex items-center space-x-2">
                  <button
                    onClick={() => {
                      if (selectedPriority !== appeal.priority) {
                        priorityMutation.mutate(selectedPriority)
                      } else {
                        setIsEditingPriority(false)
                      }
                    }}
                    disabled={priorityMutation.isPending || selectedPriority === appeal.priority}
                    className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium"
                  >
                    {priorityMutation.isPending ? 'Збереження...' : 'Зберегти'}
                  </button>
                  <button
                    onClick={() => {
                      setIsEditingPriority(false)
                      setSelectedPriority(appeal.priority || 2)
                    }}
                    disabled={priorityMutation.isPending}
                    className="px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium"
                  >
                    Скасувати
                  </button>
                </div>
                {priorityMutation.isError && (
                  <p className="text-sm text-red-600">
                    Помилка: {priorityMutation.error instanceof Error ? priorityMutation.error.message : 'Невідома помилка'}
                  </p>
                )}
              </div>
            )}
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4 mb-4 sm:mb-6">
          <div>
            <span className="text-xs sm:text-sm text-gray-500">Адреса:</span>
            <p className="text-sm sm:text-base text-gray-900 break-words">{appeal.address}</p>
          </div>
          <div>
            <span className="text-xs sm:text-sm text-gray-500">Дата створення:</span>
            <p className="text-sm sm:text-base text-gray-900">
              <span className="hidden sm:inline">
                {format(new Date(appeal.created_at), 'dd MMMM yyyy, HH:mm', { locale: uk })}
              </span>
              <span className="sm:hidden">
                {format(new Date(appeal.created_at), 'dd.MM.yyyy, HH:mm', { locale: uk })}
              </span>
            </p>
          </div>
          {appeal.category && (
            <div>
              <span className="text-xs sm:text-sm text-gray-500">Категорія:</span>
              <p className="text-sm sm:text-base text-gray-900">{appeal.category.name}</p>
            </div>
          )}
          <div className={`${!appeal.service_id ? 'border-2 border-red-300 bg-red-50 rounded-lg p-2 sm:p-3' : ''}`}>
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 sm:gap-0">
              <div className="flex-1">
                <span className="text-xs sm:text-sm text-gray-500">Служба:</span>
                {appeal.service ? (
                  <Link
                    to={`/services/${appeal.service.id}`}
                    className="text-sm sm:text-base text-primary hover:underline font-medium"
                  >
                    {appeal.service.name}
                  </Link>
                ) : (
                  <p className="text-sm sm:text-base text-red-600 font-medium">Не призначено</p>
                )}
              </div>
              {(user?.role === 'dispatcher' || user?.role === 'admin') && (
                <button
                  onClick={() => setIsAssigning(!isAssigning)}
                  className={`px-3 sm:px-4 py-1.5 sm:py-2 rounded-md text-xs sm:text-sm font-medium ${
                    appeal.service_id
                      ? 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                      : 'bg-red-100 text-red-700 hover:bg-red-200'
                  }`}
                >
                  {isAssigning ? 'Скасувати' : appeal.service_id ? 'Змінити службу' : 'Призначити службу'}
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Assign appeal panel - Dispatcher/Admin */}
        {(user?.role === 'dispatcher' || user?.role === 'admin') && isAssigning && (
          <div className="mb-6 p-4 bg-green-50 rounded-lg border border-green-200">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Призначити звернення до служби</h3>
            <p className="text-sm text-gray-600 mb-4">
              Оберіть службу та встановіть пріоритет для призначення звернення
            </p>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Служба <span className="text-red-500">*</span>
                </label>
                <select
                  value={selectedServiceId}
                  onChange={(e) => setSelectedServiceId(Number(e.target.value))}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-primary focus:border-primary"
                  disabled={assignMutation.isPending}
                >
                  <option value={0}>Оберіть службу</option>
                  {services?.map((service) => (
                    <option key={service.id} value={service.id}>
                      {service.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Пріоритет
                </label>
                <select
                  value={selectedPriority}
                  onChange={(e) => setSelectedPriority(Number(e.target.value))}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-primary focus:border-primary"
                  disabled={assignMutation.isPending}
                >
                  <option value={1}>Низький</option>
                  <option value={2}>Середній</option>
                  <option value={3}>Високий</option>
                </select>
              </div>
              <div className="flex items-center space-x-2">
                <button
                  onClick={() => {
                    if (!selectedServiceId) {
                      alert('Будь ласка, оберіть службу')
                      return
                    }
                    if (confirm('Призначити звернення до служби?')) {
                      assignMutation.mutate({ 
                        serviceId: selectedServiceId, 
                        priority: selectedPriority
                      })
                    }
                  }}
                  disabled={assignMutation.isPending || !selectedServiceId || !services || services.length === 0}
                  className="px-6 py-3 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
                >
                  {assignMutation.isPending ? 'Призначення...' : 'Призначити'}
                </button>
                <button
                  onClick={() => {
                    setIsAssigning(false)
                    setSelectedServiceId(appeal.service_id || 0)
                    setSelectedPriority(appeal.priority || 2)
                  }}
                  disabled={assignMutation.isPending}
                  className="px-6 py-3 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
                >
                  Скасувати
                </button>
              </div>
              {assignMutation.isSuccess && (
                <p className="mt-3 text-sm text-green-600 font-medium">✓ Звернення успішно призначено!</p>
              )}
              {assignMutation.isError && (
                <p className="mt-3 text-sm text-red-600">
                  Помилка: {assignMutation.error instanceof Error ? assignMutation.error.message : 'Невідома помилка'}
                </p>
              )}
            </div>
          </div>
        )}

        <div className="mt-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold">Фотографії</h2>
            {photos && photos.length > 0 && (
              <span className="text-sm text-gray-500">
                {photos.length} / 5 фото
              </span>
            )}
          </div>

          {photos && photos.length > 0 ? (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
              {photos.map((photo) => {
                const photoUrl = photo.url 
                  ? (photo.url.startsWith('http') ? photo.url : `http://localhost:8080${photo.url}`)
                  : `http://localhost:8080/uploads/${photo.file_path}`
                
                return (
                  <div key={photo.id} className="relative group">
                    <img
                      src={photoUrl}
                      alt={photo.file_name}
                      className="w-full h-32 object-cover rounded-lg border border-gray-300 cursor-pointer hover:opacity-90"
                      onClick={() => window.open(photoUrl, '_blank')}
                    />
                    {photo.is_result_photo && (
                      <div className="absolute top-2 left-2 bg-green-500 text-white text-xs px-2 py-1 rounded">
                        Результат
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          ) : (
            <p className="text-gray-500 text-sm mb-4">Фотографії відсутні</p>
          )}

          {/* Photo upload for existing appeal */}
          {canUploadPhotos && (
            <div>
              {!isUploadingPhotos ? (
                <button
                  onClick={() => setIsUploadingPhotos(true)}
                  disabled={(photos?.length || 0) >= 5}
                  className="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium flex items-center space-x-2"
                >
                  <Image className="h-4 w-4" />
                  <span>Додати фото</span>
                </button>
              ) : (
                <div className="space-y-3 p-4 bg-gray-50 rounded-lg border border-gray-200">
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-medium text-gray-700">Додати фото</h3>
                    <button
                      onClick={() => setIsUploadingPhotos(false)}
                      className="text-gray-500 hover:text-gray-700"
                      title="Закрити"
                    >
                      <X className="h-4 w-4" />
                    </button>
                  </div>
                  <PhotoUpload
                    onUpload={async (files) => {
                      await photoMutation.mutateAsync(files)
                      setIsUploadingPhotos(false)
                    }}
                    maxFiles={5 - (photos?.length || 0)}
                    maxSizeMB={5}
                    disabled={photoMutation.isPending || (photos?.length || 0) >= 5}
                  />
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* History section */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mt-6">
        <button
          onClick={() => setIsHistoryVisible(!isHistoryVisible)}
          className="flex items-center justify-between w-full text-left"
        >
          <div className="flex items-center space-x-2">
            <History className="h-5 w-5 text-gray-600" />
            <h2 className="text-xl font-semibold">Історія змін статусів</h2>
            {history && history.length > 0 && (
              <span className="text-sm text-gray-500">({history.length})</span>
            )}
          </div>
          {isHistoryVisible ? (
            <ChevronUp className="h-5 w-5 text-gray-600" />
          ) : (
            <ChevronDown className="h-5 w-5 text-gray-600" />
          )}
        </button>

        {isHistoryVisible && (
          <div className="mt-4">
            {isLoadingHistory ? (
              <div className="flex items-center justify-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              </div>
            ) : history && history.length > 0 ? (
              <div className="space-y-4">
                {history.map((item: any, index: number) => (
                  <div
                    key={item.id}
                    className="relative pl-8 pb-4 border-l-2 border-gray-200 last:border-l-0 last:pb-0"
                  >
                    {index < history.length - 1 && (
                      <div className="absolute left-[-6px] top-6 w-3 h-3 bg-primary rounded-full"></div>
                    )}
                    <div className="bg-gray-50 rounded-lg p-4">
                      <div className="flex items-start justify-between mb-2">
                        <div className="flex-1">
                          <div className="flex items-center space-x-2 mb-1">
                            {item.old_status && (
                              <>
                                <span className="px-2 py-1 text-xs font-medium bg-gray-200 text-gray-700 rounded">
                                  {statusLabels[item.old_status] || item.old_status}
                                </span>
                                <span className="text-gray-400">→</span>
                              </>
                            )}
                            <span className={`px-2 py-1 text-xs font-medium rounded ${
                              statusColors[item.new_status] || 'bg-gray-200 text-gray-700'
                            }`}>
                              {statusLabels[item.new_status] || item.new_status}
                            </span>
                          </div>
                          <p className="text-sm text-gray-600 mt-1">{item.action}</p>
                          {item.comment && (
                            <div className="mt-2 p-2 bg-white rounded border border-gray-200">
                              <p className="text-sm text-gray-700">{item.comment}</p>
                            </div>
                          )}
                        </div>
                        <div className="text-right ml-4">
                          <p className="text-xs text-gray-500">
                            {format(new Date(item.created_at), 'dd MMMM yyyy, HH:mm', { locale: uk })}
                          </p>
                          {item.user && (
                            <p className="text-xs text-gray-400 mt-1">
                              {item.user.first_name} {item.user.last_name}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-gray-500 text-sm py-4">Історія змін відсутня</p>
            )}
          </div>
        )}
      </div>

      {/* Map with location */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mt-6">
        <h2 className="text-xl font-semibold mb-4">Локація</h2>
        <MapView
          latitude={appeal.latitude}
          longitude={appeal.longitude}
          height="400px"
          zoom={15}
        />
        <div className="mt-4 text-sm text-gray-600">
          <p><strong>Адреса:</strong> {appeal.address}</p>
          <p className="mt-1">
            <strong>Координати:</strong> {appeal.latitude.toFixed(6)}, {appeal.longitude.toFixed(6)}
          </p>
        </div>
      </div>

      {/* Comments section */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mt-6">
        <div className="flex items-center space-x-2 mb-4">
          <MessageSquare className="h-5 w-5 text-gray-600" />
          <h2 className="text-xl font-semibold">Коментарі</h2>
          {comments && comments.length > 0 && (
            <span className="text-sm text-gray-500">({comments.length})</span>
          )}
        </div>

        {/* Comments list */}
        {comments && comments.length > 0 ? (
          <div className="space-y-4 mb-6">
            {comments.map((comment: Comment) => (
              <div
                key={comment.id}
                className={`p-4 rounded-lg border ${
                  comment.is_internal
                    ? 'bg-blue-50 border-blue-200'
                    : 'bg-gray-50 border-gray-200'
                }`}
              >
                {editingCommentId === comment.id ? (
                  <div className="space-y-3">
                    <textarea
                      value={editCommentText}
                      onChange={(e) => setEditCommentText(e.target.value)}
                      rows={3}
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-primary focus:border-primary resize-none"
                    />
                    <div className="flex items-center space-x-2">
                      <button
                        onClick={() => {
                          if (editCommentText.trim()) {
                            updateCommentMutation.mutate({ id: comment.id, text: editCommentText })
                          }
                        }}
                        disabled={updateCommentMutation.isPending || !editCommentText.trim()}
                        className="px-3 py-1 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed text-sm flex items-center space-x-1"
                      >
                        <Send className="h-3 w-3" />
                        <span>Зберегти</span>
                      </button>
                      <button
                        onClick={() => {
                          setEditingCommentId(null)
                          setEditCommentText('')
                        }}
                        disabled={updateCommentMutation.isPending}
                        className="px-3 py-1 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400 disabled:opacity-50 disabled:cursor-not-allowed text-sm flex items-center space-x-1"
                      >
                        <X className="h-3 w-3" />
                        <span>Скасувати</span>
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-1">
                          <span className="font-medium text-gray-900">
                            {comment.user?.first_name} {comment.user?.last_name}
                          </span>
                          {comment.is_internal && (
                            <span className="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded">
                              Внутрішній
                            </span>
                          )}
                          <span className="text-xs text-gray-500">
                            {format(new Date(comment.created_at), 'dd MMMM yyyy, HH:mm', { locale: uk })}
                          </span>
                        </div>
                        <p className="text-gray-700 whitespace-pre-wrap">{comment.text}</p>
                      </div>
                      {(user?.id === comment.user_id || user?.role === 'admin' || user?.role === 'dispatcher') && (
                        <div className="flex items-center space-x-2">
                          {user?.id === comment.user_id && (
                            <button
                              onClick={() => {
                                setEditingCommentId(comment.id)
                                setEditCommentText(comment.text)
                              }}
                              className="p-1 text-gray-500 hover:text-blue-600"
                              title="Редагувати"
                            >
                              <Edit className="h-4 w-4" />
                            </button>
                          )}
                          <button
                            onClick={() => {
                              if (confirm('Ви впевнені, що хочете видалити цей коментар?')) {
                                deleteCommentMutation.mutate(comment.id)
                              }
                            }}
                            disabled={deleteCommentMutation.isPending}
                            className="p-1 text-gray-500 hover:text-red-600 disabled:opacity-50"
                            title="Видалити"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                      )}
                    </div>
                  </>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500 text-sm mb-6">Коментарі відсутні</p>
        )}

        {/* Add new comment */}
        <div className="border-t pt-4">
          <h3 className="text-sm font-medium text-gray-700 mb-2">Додати коментар</h3>
          <textarea
            value={newCommentText}
            onChange={(e) => setNewCommentText(e.target.value)}
            placeholder="Введіть ваш коментар..."
            rows={3}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-primary focus:border-primary resize-none mb-2"
          />
          {(user?.role === 'dispatcher' || user?.role === 'admin' || user?.role === 'executor') && (
            <label className="flex items-center space-x-2 mb-3">
              <input
                type="checkbox"
                checked={isInternalComment}
                onChange={(e) => setIsInternalComment(e.target.checked)}
                className="rounded border-gray-300 text-primary focus:ring-primary"
              />
              <span className="text-sm text-gray-700">Внутрішній коментар (видно тільки співробітникам)</span>
            </label>
          )}
          <button
            onClick={() => {
              if (newCommentText.trim()) {
                createCommentMutation.mutate(newCommentText)
              }
            }}
            disabled={createCommentMutation.isPending || !newCommentText.trim()}
            className="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium flex items-center space-x-2"
          >
            <Send className="h-4 w-4" />
            <span>{createCommentMutation.isPending ? 'Відправка...' : 'Відправити'}</span>
          </button>
          {createCommentMutation.isError && (
            <p className="mt-2 text-sm text-red-600">
              Помилка: {createCommentMutation.error instanceof Error ? createCommentMutation.error.message : 'Невідома помилка'}
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

