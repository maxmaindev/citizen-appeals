import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { authAPI } from '../lib/api'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { User } from '../types'
import { Save, Lock, User as UserIcon } from 'lucide-react'

export default function ProfilePage() {
  const { user, setUser } = useAuth()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<'profile' | 'password'>('profile')

  // Profile form state
  const [profileData, setProfileData] = useState({
    first_name: '',
    last_name: '',
    phone: '',
  })

  // Password form state
  const [passwordData, setPasswordData] = useState({
    current_password: '',
    new_password: '',
    confirm_password: '',
  })

  const [profileError, setProfileError] = useState('')
  const [profileSuccess, setProfileSuccess] = useState('')
  const [passwordError, setPasswordError] = useState('')
  const [passwordSuccess, setPasswordSuccess] = useState('')

  useEffect(() => {
    if (user) {
      setProfileData({
        first_name: user.first_name || '',
        last_name: user.last_name || '',
        phone: user.phone || '',
      })
    }
  }, [user])

  const updateProfileMutation = useMutation({
    mutationFn: (data: { first_name?: string; last_name?: string; phone?: string }) =>
      authAPI.updateProfile(data),
    onSuccess: (response) => {
      if (response.success && response.data) {
        setProfileSuccess('Профіль успішно оновлено')
        setProfileError('')
        // Update user in context
        setUser(response.data)
        // Update localStorage
        localStorage.setItem('user', JSON.stringify(response.data))
        // Invalidate queries
        queryClient.invalidateQueries({ queryKey: ['user'] })
      } else {
        setProfileError(response.error || 'Помилка оновлення профілю')
        setProfileSuccess('')
      }
    },
    onError: (error: any) => {
      setProfileError(error.response?.data?.error || 'Помилка оновлення профілю')
      setProfileSuccess('')
    },
  })

  const changePasswordMutation = useMutation({
    mutationFn: (data: { current_password: string; new_password: string }) =>
      authAPI.changePassword(data),
    onSuccess: (response) => {
      if (response.success) {
        setPasswordSuccess('Пароль успішно змінено')
        setPasswordError('')
        setPasswordData({
          current_password: '',
          new_password: '',
          confirm_password: '',
        })
      } else {
        setPasswordError(response.error || 'Помилка зміни пароля')
        setPasswordSuccess('')
      }
    },
    onError: (error: any) => {
      setPasswordError(error.response?.data?.error || 'Помилка зміни пароля')
      setPasswordSuccess('')
    },
  })

  const handleProfileSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setProfileError('')
    setProfileSuccess('')

    if (!profileData.first_name.trim() || !profileData.last_name.trim()) {
      setProfileError('Ім\'я та прізвище обов\'язкові')
      return
    }

    updateProfileMutation.mutate(profileData)
  }

  const handlePasswordSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setPasswordError('')
    setPasswordSuccess('')

    if (!passwordData.current_password || !passwordData.new_password || !passwordData.confirm_password) {
      setPasswordError('Всі поля обов\'язкові')
      return
    }

    if (passwordData.new_password.length < 8) {
      setPasswordError('Новий пароль повинен містити мінімум 8 символів')
      return
    }

    if (passwordData.new_password !== passwordData.confirm_password) {
      setPasswordError('Нові паролі не співпадають')
      return
    }

    changePasswordMutation.mutate({
      current_password: passwordData.current_password,
      new_password: passwordData.new_password,
    })
  }

  if (!user) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-500">Завантаження...</div>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="bg-white rounded-lg shadow-sm border p-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-6">Мій профіль</h1>

        {/* Tabs */}
        <div className="border-b border-gray-200 mb-6">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => {
                setActiveTab('profile')
                setProfileError('')
                setProfileSuccess('')
              }}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'profile'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              <div className="flex items-center space-x-2">
                <UserIcon className="h-4 w-4" />
                <span>Особиста інформація</span>
              </div>
            </button>
            <button
              onClick={() => {
                setActiveTab('password')
                setPasswordError('')
                setPasswordSuccess('')
              }}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'password'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              <div className="flex items-center space-x-2">
                <Lock className="h-4 w-4" />
                <span>Зміна пароля</span>
              </div>
            </button>
          </nav>
        </div>

        {/* Profile Tab */}
        {activeTab === 'profile' && (
          <form onSubmit={handleProfileSubmit} className="space-y-6">
            {profileError && (
              <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
                {profileError}
              </div>
            )}
            {profileSuccess && (
              <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded">
                {profileSuccess}
              </div>
            )}

            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
                  Email
                </label>
                <input
                  type="email"
                  id="email"
                  value={user.email}
                  disabled
                  className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-gray-500"
                />
                <p className="mt-1 text-xs text-gray-500">Email не можна змінити</p>
              </div>

              <div>
                <label htmlFor="role" className="block text-sm font-medium text-gray-700 mb-2">
                  Роль
                </label>
                <input
                  type="text"
                  id="role"
                  value={user.role}
                  disabled
                  className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-gray-500 capitalize"
                />
                <p className="mt-1 text-xs text-gray-500">Роль не можна змінити</p>
              </div>

              <div>
                <label htmlFor="first_name" className="block text-sm font-medium text-gray-700 mb-2">
                  Ім'я <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  id="first_name"
                  required
                  minLength={2}
                  value={profileData.first_name}
                  onChange={(e) => setProfileData({ ...profileData, first_name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                />
              </div>

              <div>
                <label htmlFor="last_name" className="block text-sm font-medium text-gray-700 mb-2">
                  Прізвище <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  id="last_name"
                  required
                  minLength={2}
                  value={profileData.last_name}
                  onChange={(e) => setProfileData({ ...profileData, last_name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                />
              </div>

              <div className="sm:col-span-2">
                <label htmlFor="phone" className="block text-sm font-medium text-gray-700 mb-2">
                  Телефон
                </label>
                <input
                  type="tel"
                  id="phone"
                  value={profileData.phone}
                  onChange={(e) => setProfileData({ ...profileData, phone: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                />
              </div>
            </div>

            <div className="flex justify-end">
              <button
                type="submit"
                disabled={updateProfileMutation.isPending}
                className="flex items-center space-x-2 px-6 py-2 bg-primary text-white rounded-md hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary disabled:opacity-50"
              >
                <Save className="h-4 w-4" />
                <span>{updateProfileMutation.isPending ? 'Збереження...' : 'Зберегти зміни'}</span>
              </button>
            </div>
          </form>
        )}

        {/* Password Tab */}
        {activeTab === 'password' && (
          <form onSubmit={handlePasswordSubmit} className="space-y-6">
            {passwordError && (
              <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
                {passwordError}
              </div>
            )}
            {passwordSuccess && (
              <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded">
                {passwordSuccess}
              </div>
            )}

            <div className="space-y-4 max-w-md">
              <div>
                <label htmlFor="current_password" className="block text-sm font-medium text-gray-700 mb-2">
                  Поточний пароль <span className="text-red-500">*</span>
                </label>
                <input
                  type="password"
                  id="current_password"
                  required
                  value={passwordData.current_password}
                  onChange={(e) =>
                    setPasswordData({ ...passwordData, current_password: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                />
              </div>

              <div>
                <label htmlFor="new_password" className="block text-sm font-medium text-gray-700 mb-2">
                  Новий пароль <span className="text-red-500">*</span>
                </label>
                <input
                  type="password"
                  id="new_password"
                  required
                  minLength={8}
                  value={passwordData.new_password}
                  onChange={(e) => setPasswordData({ ...passwordData, new_password: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                />
                <p className="mt-1 text-xs text-gray-500">Мінімум 8 символів</p>
              </div>

              <div>
                <label
                  htmlFor="confirm_password"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Підтвердження пароля <span className="text-red-500">*</span>
                </label>
                <input
                  type="password"
                  id="confirm_password"
                  required
                  minLength={8}
                  value={passwordData.confirm_password}
                  onChange={(e) =>
                    setPasswordData({ ...passwordData, confirm_password: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-primary focus:border-primary"
                />
              </div>
            </div>

            <div className="flex justify-end">
              <button
                type="submit"
                disabled={changePasswordMutation.isPending}
                className="flex items-center space-x-2 px-6 py-2 bg-primary text-white rounded-md hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary disabled:opacity-50"
              >
                <Lock className="h-4 w-4" />
                <span>
                  {changePasswordMutation.isPending ? 'Зміна пароля...' : 'Змінити пароль'}
                </span>
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}

