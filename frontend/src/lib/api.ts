import axios from 'axios'
import type {
  User,
  Appeal,
  AppealsListParams,
  AppealsListResponse,
  LoginRequest,
  RegisterRequest,
  CreateAppealRequest,
  APIResponse,
  Photo,
  Category,
  Service,
  Notification,
} from '../types'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      // Не перенаправляємо автоматично, щоб не було циклу
      // Перенаправлення буде в App.tsx через перевірку user
    }
    // Log error details for debugging
    if (error.response) {
      console.error('API Error:', {
        status: error.response.status,
        statusText: error.response.statusText,
        url: error.config?.url,
        data: error.response.data,
      })
    } else {
      console.error('Network Error:', error.message)
    }
    return Promise.reject(error)
  }
)

// Auth API
export const authAPI = {
  login: async (data: LoginRequest): Promise<APIResponse<{ token: string; user: User }>> => {
    try {
      const response = await api.post('/api/auth/login', data)
      return response.data
    } catch (error: any) {
      // Якщо це помилка від сервера з повідомленням
      if (error.response?.data?.error) {
        return {
          success: false,
          error: error.response.data.error,
        }
      }
      // Якщо це 401 - неправильні дані
      if (error.response?.status === 401) {
        return {
          success: false,
          error: 'Неправильний логін чи пароль',
        }
      }
      // Інші помилки - пробрасуємо далі
      throw error
    }
  },

  register: async (data: RegisterRequest): Promise<APIResponse<{ token: string; user: User }>> => {
    const response = await api.post('/api/auth/register', data)
    return response.data
  },

  me: async (): Promise<APIResponse<User>> => {
    const response = await api.get('/api/auth/me')
    return response.data
  },

  refreshToken: async (): Promise<APIResponse<{ token: string }>> => {
    const response = await api.post('/api/auth/refresh')
    return response.data
  },

  updateProfile: async (data: { first_name?: string; last_name?: string; phone?: string }): Promise<APIResponse<User>> => {
    const response = await api.put('/api/auth/profile', data)
    return response.data
  },

  changePassword: async (data: { current_password: string; new_password: string }): Promise<APIResponse<{ message: string }>> => {
    const response = await api.put('/api/auth/change-password', data)
    return response.data
  },
}

// Appeals API
export const appealsAPI = {
  list: async (params?: AppealsListParams): Promise<APIResponse<AppealsListResponse>> => {
    const response = await api.get('/api/appeals', { params })
    return response.data
  },

  getById: async (id: number): Promise<APIResponse<Appeal>> => {
    const response = await api.get(`/api/appeals/${id}`)
    return response.data
  },

  create: async (data: CreateAppealRequest): Promise<APIResponse<Appeal>> => {
    const response = await api.post('/api/appeals', data)
    return response.data
  },

  update: async (id: number, data: Partial<CreateAppealRequest>): Promise<APIResponse<Appeal>> => {
    const response = await api.put(`/api/appeals/${id}`, data)
    return response.data
  },

  updateStatus: async (
    id: number,
    status: string,
    comment?: string
  ): Promise<APIResponse<Appeal>> => {
    const response = await api.patch(`/api/appeals/${id}/status`, { status, comment })
    return response.data
  },

  updatePriority: async (
    id: number,
    priority: number
  ): Promise<APIResponse<Appeal>> => {
    const response = await api.patch(`/api/appeals/${id}/priority`, { priority })
    return response.data
  },

  assign: async (id: number, serviceId: number, priority?: number): Promise<APIResponse<Appeal>> => {
    const response = await api.patch(`/api/appeals/${id}/assign`, { 
      service_id: serviceId,
      priority: priority
    })
    return response.data
  },

  getStatistics: async (params?: { from_date?: string; to_date?: string }): Promise<APIResponse<any>> => {
    const response = await api.get('/api/appeals/statistics', { params })
    return response.data
  },

  classify: async (text: string): Promise<APIResponse<{ service: string; confidence: number }>> => {
    const response = await api.post('/api/appeals/classify', { text })
    return response.data
  },

  getHistory: async (id: number): Promise<APIResponse<any[]>> => {
    const response = await api.get(`/api/appeals/${id}/history`)
    return response.data
  },

  getDispatcherDashboard: async (): Promise<APIResponse<any>> => {
    const response = await api.get('/api/appeals/dashboard/dispatcher')
    return response.data
  },

  getAdminDashboard: async (): Promise<APIResponse<any>> => {
    const response = await api.get('/api/appeals/dashboard/admin')
    return response.data
  },

  getExecutorDashboard: async (): Promise<APIResponse<any>> => {
    const response = await api.get('/api/appeals/dashboard/executor')
    return response.data
  },

  getServiceStatistics: async (serviceId: number): Promise<APIResponse<any>> => {
    const response = await api.get(`/api/appeals/services/${serviceId}/statistics`)
    return response.data
  },
}

// Photos API
export const photosAPI = {
  upload: async (
    appealId: number,
    files: File[],
    isResultPhoto = false
  ): Promise<APIResponse<Photo[]>> => {
    const formData = new FormData()
    files.forEach((file) => {
      formData.append('photos', file)
    })
    const response = await api.post(
      `/api/appeals/${appealId}/photos${isResultPhoto ? '?result=true' : ''}`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    )
    return response.data
  },

  list: async (appealId: number): Promise<APIResponse<Photo[]>> => {
    const response = await api.get(`/api/appeals/${appealId}/photos`)
    return response.data
  },

  get: async (id: number): Promise<Blob> => {
    const response = await api.get(`/api/photos/${id}`, { responseType: 'blob' })
    return response.data
  },

  delete: async (id: number): Promise<APIResponse> => {
    const response = await api.delete(`/api/photos/${id}`)
    return response.data
  },
}

// Comments API
export const commentsAPI = {
  getByAppealId: async (appealId: number): Promise<APIResponse<any[]>> => {
    const response = await api.get(`/api/appeals/${appealId}/comments`)
    return response.data
  },

  create: async (appealId: number, text: string, isInternal?: boolean): Promise<APIResponse<any>> => {
    const response = await api.post(`/api/appeals/${appealId}/comments`, {
      text,
      is_internal: isInternal || false,
    })
    return response.data
  },

  update: async (id: number, text: string, isInternal?: boolean): Promise<APIResponse<any>> => {
    const response = await api.put(`/api/comments/${id}`, {
      text,
      is_internal: isInternal || false,
    })
    return response.data
  },

  delete: async (id: number): Promise<APIResponse<{ message: string }>> => {
    const response = await api.delete(`/api/comments/${id}`)
    return response.data
  },
}

// Categories API
export const categoriesAPI = {
  list: async (): Promise<APIResponse<Category[]>> => {
    const response = await api.get('/api/categories')
    return response.data
  },

  getById: async (id: number): Promise<APIResponse<Category>> => {
    const response = await api.get(`/api/categories/${id}`)
    return response.data
  },
}

// Services API
export const servicesAPI = {
  list: async (): Promise<APIResponse<Service[]>> => {
    const response = await api.get('/api/services')
    return response.data
  },

  getById: async (id: number): Promise<APIResponse<Service>> => {
    const response = await api.get(`/api/services/${id}`)
    return response.data
  },

  create: async (data: { name: string; description?: string; contact_person: string; contact_phone: string; contact_email: string }): Promise<APIResponse<Service>> => {
    const response = await api.post('/api/services', data)
    return response.data
  },

  update: async (id: number, data: { name?: string; description?: string; keywords?: string; contact_person?: string; contact_phone?: string; contact_email?: string; is_active?: boolean }): Promise<APIResponse<Service>> => {
    const response = await api.put(`/api/services/${id}`, data)
    return response.data
  },

  delete: async (id: number): Promise<APIResponse<{ message: string }>> => {
    const response = await api.delete(`/api/services/${id}`)
    return response.data
  },
}

// Category-Services API
export const categoryServicesAPI = {
  getAll: async (): Promise<APIResponse<any[]>> => {
    const response = await api.get('/api/category-services')
    return response.data
  },

  getByCategoryId: async (categoryId: number): Promise<APIResponse<any[]>> => {
    const response = await api.get(`/api/category-services/category/${categoryId}`)
    return response.data
  },

  assignServices: async (categoryId: number, serviceIds: number[]): Promise<APIResponse<any[]>> => {
    const response = await api.post('/api/category-services/assign', {
      category_id: categoryId,
      service_ids: serviceIds,
    })
    return response.data
  },

  delete: async (categoryId: number, serviceId: number): Promise<APIResponse<{ message: string }>> => {
    const response = await api.delete(`/api/category-services/category/${categoryId}/service/${serviceId}`)
    return response.data
  },
}

// Users API
export const usersAPI = {
  list: async (params?: { page?: number; limit?: number; role?: string }): Promise<APIResponse<{ items: User[]; total: number; page: number; limit: number; total_pages: number } | User[]>> => {
    const response = await api.get('/api/users', { params })
    return response.data
  },

  getById: async (id: number): Promise<APIResponse<User>> => {
    const response = await api.get(`/api/users/${id}`)
    return response.data
  },

  update: async (id: number, data: { first_name?: string; last_name?: string; phone?: string; role?: string; is_active?: boolean }): Promise<APIResponse<User>> => {
    const response = await api.put(`/api/users/${id}`, data)
    return response.data
  },

  delete: async (id: number): Promise<APIResponse<{ message: string }>> => {
    const response = await api.delete(`/api/users/${id}`)
    return response.data
  },
}

// User-Services API
export const userServicesAPI = {
  getAll: async (): Promise<APIResponse<any[]>> => {
    const response = await api.get('/api/user-services')
    return response.data
  },

  getMyServices: async (): Promise<APIResponse<Service[]>> => {
    const response = await api.get('/api/user-services/me')
    return response.data
  },

  getByServiceId: async (serviceId: number): Promise<APIResponse<User[]>> => {
    const response = await api.get(`/api/user-services/service/${serviceId}`)
    return response.data
  },

  assignUsers: async (serviceId: number, userIds: number[]): Promise<APIResponse<{ message: string }>> => {
    const response = await api.post('/api/user-services/assign', {
      service_id: serviceId,
      user_ids: userIds,
    })
    return response.data
  },

  delete: async (serviceId: number, userId: number): Promise<APIResponse<{ message: string }>> => {
    const response = await api.delete(`/api/user-services/service/${serviceId}/user/${userId}`)
    return response.data
  },
}

// System settings API
export const systemSettingsAPI = {
  get: async (): Promise<
    APIResponse<{
      city_name: string
      map_center_lat: number
      map_center_lng: number
      map_zoom: number
      confidence_threshold: number
    }>
  > => {
    const response = await api.get('/api/system-settings')
    return response.data
  },

  update: async (data: {
    city_name: string
    map_center_lat: number
    map_center_lng: number
    map_zoom: number
    confidence_threshold: number
  }): Promise<APIResponse<any>> => {
    const response = await api.put('/api/system-settings', data)
    return response.data
  },
}

// Notifications API
export const notificationsAPI = {
  list: async (params?: { page?: number; limit?: number }): Promise<APIResponse<Notification[]>> => {
    const response = await api.get('/api/notifications', { params })
    return response.data
  },

  getUnreadCount: async (): Promise<APIResponse<{ count: number }>> => {
    const response = await api.get('/api/notifications/unread-count')
    return response.data
  },

  markAsRead: async (id: number): Promise<APIResponse<{ message: string }>> => {
    const response = await api.put(`/api/notifications/${id}/read`)
    return response.data
  },

  markAllAsRead: async (): Promise<APIResponse<{ message: string }>> => {
    const response = await api.put('/api/notifications/read-all')
    return response.data
  },

  delete: async (id: number): Promise<APIResponse<{ message: string }>> => {
    const response = await api.delete(`/api/notifications/${id}`)
    return response.data
  },
}

// Classification API (ML Service)
const CLASSIFICATION_API_URL = import.meta.env.VITE_CLASSIFICATION_API_URL || 'http://localhost:8000'

const classificationApi = axios.create({
  baseURL: CLASSIFICATION_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

export interface ClassificationHistoryEntry {
  id: number
  text: string
  service: string
  confidence: number
  needs_moderation: boolean
  top_alternatives: Array<{ service: string; confidence: number }>
  timestamp: string
  processing_time_ms: number
}

export interface ClassificationHistoryResponse {
  entries: ClassificationHistoryEntry[]
  total: number
  page: number
  limit: number
}

export const classificationAPI = {
  getHistory: async (params?: {
    page?: number
    limit?: number
    service?: string
    needs_moderation?: boolean
  }): Promise<ClassificationHistoryResponse> => {
    const response = await classificationApi.get('/classifications/history', { params })
    return response.data
  },
}

export default api

