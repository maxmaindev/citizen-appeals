import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { authAPI } from '../lib/api'
import type { User } from '../types'

interface AuthContextType {
  user: User | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (data: {
    email: string
    password: string
    first_name: string
    last_name: string
    phone: string
  }) => Promise<void>
  logout: () => void
  setUser: (user: User | null) => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('token')
      const savedUser = localStorage.getItem('user')

      if (token && savedUser) {
        try {
          const response = await authAPI.me()
          if (response.success && response.data) {
            setUser(response.data)
            localStorage.setItem('user', JSON.stringify(response.data))
          } else {
            localStorage.removeItem('token')
            localStorage.removeItem('user')
          }
        } catch (error) {
          console.error('Failed to fetch user:', error)
          localStorage.removeItem('token')
          localStorage.removeItem('user')
        }
      }
      setIsLoading(false)
    }

    initAuth()
  }, [])

  const login = async (email: string, password: string) => {
    try {
      const response = await authAPI.login({ email, password })
      if (response.success && response.data) {
        localStorage.setItem('token', response.data.token)
        localStorage.setItem('user', JSON.stringify(response.data.user))
        setUser(response.data.user)
      } else {
        throw new Error(response.error || 'Неправильний логін чи пароль')
      }
    } catch (error: any) {
      // Якщо це axios помилка з 401 статусом
      if (error.response?.status === 401) {
        throw new Error('Неправильний логін чи пароль')
      }
      // Якщо це помилка від API (response.error)
      if (error.response?.data?.error) {
        throw new Error(error.response.data.error)
      }
      // Інші помилки
      throw error
    }
  }

  const register = async (data: {
    email: string
    password: string
    first_name: string
    last_name: string
    phone: string
  }) => {
    const response = await authAPI.register(data)
    if (response.success && response.data) {
      localStorage.setItem('token', response.data.token)
      localStorage.setItem('user', JSON.stringify(response.data.user))
      setUser(response.data.user)
    } else {
      throw new Error(response.error || 'Registration failed')
    }
  }

  const logout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, login, register, logout, setUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

