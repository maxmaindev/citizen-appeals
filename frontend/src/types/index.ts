export interface User {
  id: number
  email: string
  first_name: string
  last_name: string
  phone: string
  role: 'citizen' | 'dispatcher' | 'executor' | 'admin'
  is_active: boolean
  created_at: string
}

export interface Category {
  id: number
  name: string
  description?: string
  parent_id?: number
  is_active: boolean
}

export interface Service {
  id: number
  name: string
  description?: string
  keywords?: string
  contact_person?: string
  contact_phone?: string
  contact_email?: string
  is_active: boolean
}

export interface Appeal {
  id: number
  user_id: number
  category_id?: number
  service_id?: number
  status: 'new' | 'assigned' | 'in_progress' | 'completed' | 'closed' | 'rejected'
  title: string
  description: string
  address: string
  latitude: number
  longitude: number
  priority: number
  created_at: string
  updated_at: string
  closed_at?: string
  category?: Category
  service?: Service
  user?: User
}

export interface Photo {
  id: number
  appeal_id?: number
  comment_id?: number
  file_path: string
  file_name: string
  file_size: number
  mime_type: string
  is_result_photo: boolean
  uploaded_at: string
  url?: string
}

export interface APIResponse<T = any> {
  success: boolean
  data?: T
  error?: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  first_name: string
  last_name: string
  phone: string
}

export interface CreateAppealRequest {
  title: string
  description: string
  category_id?: number
  address: string
  latitude: number
  longitude: number
}

export interface AppealsListParams {
  page?: number
  limit?: number
  status?: string
  category_id?: number
  service_id?: number
  user_id?: number
  search?: string
  sort_by?: string
  order?: 'asc' | 'desc'
  date_from?: string
  date_to?: string
}

export interface AppealsListResponse {
  items: Appeal[]  // Бекенд повертає items, а не appeals
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface Comment {
  id: number
  appeal_id: number
  user_id: number
  text: string
  is_internal: boolean
  created_at: string
  user?: User
  photos?: Photo[]
}

export interface CreateCommentRequest {
  text: string
  is_internal?: boolean
  photo_ids?: number[]
}

export interface Notification {
  id: number
  user_id: number
  appeal_id?: number
  type: 'appeal_created' | 'appeal_assigned' | 'status_changed' | 'comment_added' | 'appeal_completed'
  title: string
  message: string
  is_read: boolean
  sent_at: string
  appeal?: Appeal
}

