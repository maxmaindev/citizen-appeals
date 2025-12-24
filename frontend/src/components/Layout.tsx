import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { useQuery } from '@tanstack/react-query'
import { notificationsAPI } from '../lib/api'
import { LogOut, Plus, Home, Map, Users, BarChart3, Settings, Bell, History, ChevronDown, Key, Link2 } from 'lucide-react'
import { useState, useRef, useEffect } from 'react'

// Dropdown Menu Component
function DropdownMenu({ label, icon: Icon, children, isActive }: { label: string; icon: any; children: React.ReactNode; isActive?: boolean }) {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen])

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`px-2 py-1.5 sm:px-3 sm:py-2 rounded-md text-xs sm:text-sm font-medium flex items-center space-x-1 ${
          isActive
            ? 'bg-primary text-white'
            : 'text-gray-700 hover:bg-gray-100'
        }`}
      >
        <Icon className="h-4 w-4" />
        <span>{label}</span>
        <ChevronDown className={`h-3 w-3 transition-transform ${isOpen ? 'rotate-180' : ''}`} />
      </button>
      {isOpen && (
        <div className="absolute top-full left-0 mt-1 w-48 bg-white rounded-md shadow-lg border border-gray-200 z-50 py-1">
          {children}
        </div>
      )}
    </div>
  )
}

function DropdownItem({ to, icon: Icon, label, onClick }: { to?: string; icon?: any; label: string; onClick?: () => void }) {
  const location = useLocation()
  const isActive = to ? location.pathname === to || (to !== '/' && location.pathname.startsWith(to)) : false

  const content = (
    <div className={`flex items-center space-x-2 px-4 py-2 text-sm ${
      isActive
        ? 'bg-primary/10 text-primary font-medium'
        : 'text-gray-700 hover:bg-gray-100'
    }`}>
      {Icon && <Icon className="h-4 w-4" />}
      <span>{label}</span>
    </div>
  )

  if (to) {
    return (
      <Link to={to} onClick={onClick} className="block">
        {content}
      </Link>
    )
  }

  return (
    <div onClick={onClick} className="cursor-pointer">
      {content}
    </div>
  )
}

export default function Layout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  // Poll for unread notifications count every 30 seconds
  const { data: unreadCountData } = useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: async () => {
      const response = await notificationsAPI.getUnreadCount()
      return response.data?.count || 0
    },
    enabled: !!user,
    refetchInterval: 30000, // Poll every 30 seconds
  })

  const unreadCount = unreadCountData || 0

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  // Визначаємо, чи активне адмін меню
  const isAdminMenuActive = location.pathname === '/users' || 
                            location.pathname === '/system-settings' || 
                            location.pathname === '/classifications/history'

  // Визначаємо, чи активне меню служб
  const isServicesMenuActive = location.pathname.startsWith('/services') || 
                               location.pathname.startsWith('/category-services') ||
                               location.pathname.startsWith('/service-users')

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm border-b">
        <div className="w-full px-3 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-14 sm:h-16">
            <Link to="/" className="flex items-center space-x-2">
              <Home className="h-6 w-6 text-primary" />
              <span className="hidden custom:block text-xl font-bold text-gray-900">Звернення громадян</span>
            </Link>
            <nav className="flex items-center space-x-2 sm:space-x-4">
                <Link
                  to="/"
                  className={`px-2 py-1.5 sm:px-3 sm:py-2 rounded-md text-xs sm:text-sm font-medium ${
                    location.pathname === '/'
                      ? 'bg-primary text-white'
                      : 'text-gray-700 hover:bg-gray-100'
                  }`}
                >
                  Список
                </Link>
                <Link
                  to="/map"
                  className={`px-2 py-1.5 sm:px-3 sm:py-2 rounded-md text-xs sm:text-sm font-medium flex items-center space-x-1 ${
                    location.pathname === '/map'
                      ? 'bg-primary text-white'
                      : 'text-gray-700 hover:bg-gray-100'
                  }`}
                >
                  <Map className="h-4 w-4" />
                  <span>Карта</span>
                </Link>
                {user?.role === 'admin' && (
                  <Link
                    to="/analytics"
                    className={`px-2 py-1.5 sm:px-3 sm:py-2 rounded-md text-xs sm:text-sm font-medium flex items-center space-x-1 ${
                      location.pathname === '/analytics'
                        ? 'bg-primary text-white'
                        : 'text-gray-700 hover:bg-gray-100'
                    }`}
                  >
                    <BarChart3 className="h-4 w-4" />
                    <span>Аналітика</span>
                  </Link>
                )}
                {(user?.role === 'admin' || user?.role === 'dispatcher') && (
                  <Link
                    to={user?.role === 'admin' ? '/dashboard/admin' : '/dashboard/dispatcher'}
                    className={`px-2 py-1.5 sm:px-3 sm:py-2 rounded-md text-xs sm:text-sm font-medium flex items-center space-x-1 ${
                      location.pathname.startsWith('/dashboard')
                        ? 'bg-primary text-white'
                        : 'text-gray-700 hover:bg-gray-100'
                    }`}
                  >
                    <BarChart3 className="h-4 w-4" />
                    <span>Дешборд</span>
                  </Link>
                )}
                {(user?.role === 'admin' || user?.role === 'dispatcher') && (
                  <DropdownMenu 
                    label="Служби" 
                    icon={Settings}
                    isActive={isServicesMenuActive}
                  >
                    <DropdownItem to="/services" label="Список служб" icon={Settings} />
                    {user?.role === 'admin' && (
                      <DropdownItem to="/services/keywords" label="Ключові слова" icon={Key} />
                    )}
                    <DropdownItem to="/category-services" label="Призначення до категорій" icon={Link2} />
                    <DropdownItem to="/service-users" label="Призначення виконавців" icon={Users} />
                  </DropdownMenu>
                )}
                {user?.role === 'executor' && (
                  <Link
                    to="/dashboard/executor"
                    className={`px-2 py-1.5 sm:px-3 sm:py-2 rounded-md text-xs sm:text-sm font-medium flex items-center space-x-1 ${
                      location.pathname === '/dashboard/executor'
                        ? 'bg-primary text-white'
                        : 'text-gray-700 hover:bg-gray-100'
                    }`}
                  >
                    <BarChart3 className="h-4 w-4" />
                    <span>Мій дешборд</span>
                  </Link>
                )}
                {user?.role === 'admin' && (
                  <DropdownMenu 
                    label="Адміністрування" 
                    icon={Settings}
                    isActive={isAdminMenuActive}
                  >
                    <DropdownItem to="/users" label="Користувачі" icon={Users} />
                    <DropdownItem to="/system-settings" label="Налаштування системи" icon={Settings} />
                    <DropdownItem to="/classifications/history" label="Історія класифікацій" icon={History} />
                  </DropdownMenu>
                )}
            </nav>

            <div className="flex items-center space-x-2 sm:space-x-4">
              {user && (
                <>
                  <Link
                    to="/notifications"
                    className="relative flex items-center justify-center p-1.5 sm:p-2 text-gray-700 hover:text-primary hover:bg-gray-100 rounded-md"
                    title="Сповіщення"
                  >
                    <Bell className="h-5 w-5" />
                    {unreadCount > 0 && (
                      <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs font-bold rounded-full h-5 w-5 flex items-center justify-center">
                        {unreadCount > 9 ? '9+' : unreadCount}
                      </span>
                    )}
                  </Link>
                  <Link
                    to="/profile"
                    className="hidden sm:inline text-sm text-gray-700 hover:text-primary cursor-pointer"
                  >
                    {user.first_name} 
                  </Link>
                  <span className="hidden sm:inline-flex text-xs text-gray-500 bg-gray-100 px-2 py-1 rounded">
                    {user.role}
                  </span>
                  <Link
                    to="/appeals/create"
                    className="flex items-center space-x-1 px-3 py-1.5 sm:px-4 sm:py-2 bg-primary text-white rounded-md hover:bg-primary/90 text-xs sm:text-sm"
                  >
                    <Plus className="h-4 w-4" />
                    <span>Створити</span>
                  </Link>
                  <button
                    onClick={handleLogout}
                    className="hidden sm:flex items-center space-x-1 px-3 py-1.5 sm:px-4 sm:py-2 text-xs sm:text-sm text-gray-700 hover:text-gray-900"
                  >
                    <LogOut className="h-4 w-4" />
                    <span>Вийти</span>
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      </nav>

      <main className={location.pathname === '/map' ? '' : 'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'}>
        <Outlet />
      </main>
    </div>
  )
}

