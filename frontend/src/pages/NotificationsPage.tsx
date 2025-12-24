import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { notificationsAPI } from '../lib/api'
import { format } from 'date-fns'
import { uk } from 'date-fns/locale'
import { Bell, Check, CheckCheck, Trash2, ExternalLink } from 'lucide-react'
import { Link } from 'react-router-dom'
import type { Notification } from '../types'

export default function NotificationsPage() {
  const queryClient = useQueryClient()

  const { data: notifications, isLoading } = useQuery({
    queryKey: ['notifications'],
    queryFn: async () => {
      const response = await notificationsAPI.list({ limit: 100 })
      return response.data || []
    },
  })

  const markAsReadMutation = useMutation({
    mutationFn: (id: number) => notificationsAPI.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['notifications', 'unread-count'] })
    },
  })

  const markAllAsReadMutation = useMutation({
    mutationFn: () => notificationsAPI.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['notifications', 'unread-count'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => notificationsAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['notifications', 'unread-count'] })
    },
  })

  const handleMarkAsRead = (id: number) => {
    markAsReadMutation.mutate(id)
  }

  const handleMarkAllAsRead = () => {
    markAllAsReadMutation.mutate()
  }

  const handleDelete = (id: number) => {
    if (confirm('Видалити це сповіщення?')) {
      deleteMutation.mutate(id)
    }
  }

  const unreadCount = notifications?.filter((n) => !n.is_read).length || 0

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <Bell className="h-8 w-8 text-primary" />
          <h1 className="text-3xl font-bold text-gray-900">Сповіщення</h1>
          {unreadCount > 0 && (
            <span className="bg-red-500 text-white text-sm font-semibold px-3 py-1 rounded-full">
              {unreadCount} непрочитаних
            </span>
          )}
        </div>
        {unreadCount > 0 && (
          <button
            onClick={handleMarkAllAsRead}
            disabled={markAllAsReadMutation.isPending}
            className="flex items-center space-x-2 px-4 py-2 bg-primary text-white rounded-md hover:bg-primary/90 disabled:opacity-50"
          >
            <CheckCheck className="h-4 w-4" />
            <span>Позначити всі як прочитані</span>
          </button>
        )}
      </div>

      {!notifications || notifications.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
          <Bell className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-500 text-lg">Немає сповіщень</p>
        </div>
      ) : (
        <div className="space-y-3">
          {notifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
              onMarkAsRead={handleMarkAsRead}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function NotificationItem({
  notification,
  onMarkAsRead,
  onDelete,
}: {
  notification: Notification
  onMarkAsRead: (id: number) => void
  onDelete: (id: number) => void
}) {
  const getTypeIcon = () => {
    switch (notification.type) {
      case 'appeal_created':
        return '🆕'
      case 'appeal_assigned':
        return '📋'
      case 'status_changed':
        return '🔄'
      case 'comment_added':
        return '💬'
      case 'appeal_completed':
        return '✅'
      default:
        return '🔔'
    }
  }

  const getTypeColor = () => {
    switch (notification.type) {
      case 'appeal_created':
        return 'bg-blue-100 border-blue-200'
      case 'appeal_assigned':
        return 'bg-purple-100 border-purple-200'
      case 'status_changed':
        return 'bg-yellow-100 border-yellow-200'
      case 'comment_added':
        return 'bg-green-100 border-green-200'
      case 'appeal_completed':
        return 'bg-green-100 border-green-200'
      default:
        return 'bg-gray-100 border-gray-200'
    }
  }

  return (
    <div
      className={`bg-white rounded-lg shadow-sm border-2 p-4 ${
        !notification.is_read ? getTypeColor() : 'border-gray-200'
      } ${!notification.is_read ? 'font-semibold' : ''}`}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-start space-x-3">
            <span className="text-2xl">{getTypeIcon()}</span>
            <div className="flex-1">
              <div className="flex items-center space-x-2">
                <h3 className="text-lg font-semibold text-gray-900">{notification.title}</h3>
                {!notification.is_read && (
                  <span className="h-2 w-2 bg-red-500 rounded-full"></span>
                )}
              </div>
              <p className="text-gray-700 mt-1">{notification.message}</p>
              <div className="flex items-center space-x-4 mt-2 text-sm text-gray-500">
                <span>
                  {format(new Date(notification.sent_at), 'dd MMMM yyyy, HH:mm', { locale: uk })}
                </span>
                {notification.appeal_id && (
                  <Link
                    to={`/appeals/${notification.appeal_id}`}
                    className="flex items-center space-x-1 text-primary hover:underline"
                  >
                    <span>Переглянути звернення</span>
                    <ExternalLink className="h-3 w-3" />
                  </Link>
                )}
              </div>
            </div>
          </div>
        </div>
        <div className="flex items-center space-x-2 ml-4">
          {!notification.is_read && (
            <button
              onClick={() => onMarkAsRead(notification.id)}
              className="p-2 text-gray-600 hover:text-primary hover:bg-gray-100 rounded"
              title="Позначити як прочитане"
            >
              <Check className="h-5 w-5" />
            </button>
          )}
          <button
            onClick={() => onDelete(notification.id)}
            className="p-2 text-gray-600 hover:text-red-600 hover:bg-gray-100 rounded"
            title="Видалити"
          >
            <Trash2 className="h-5 w-5" />
          </button>
        </div>
      </div>
    </div>
  )
}

