import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './contexts/AuthContext'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import AppealsPage from './pages/AppealsPage'
import AppealsMapPage from './pages/AppealsMapPage'
import AppealDetailPage from './pages/AppealDetailPage'
import CreateAppealPage from './pages/CreateAppealPage'
import UsersPage from './pages/UsersPage'
import AnalyticsPage from './pages/AnalyticsPage'
import ProfilePage from './pages/ProfilePage'
import ServicesPage from './pages/ServicesPage'
import ServiceKeywordsPage from './pages/ServiceKeywordsPage'
import CategoryServicesPage from './pages/CategoryServicesPage'
import ServiceUsersPage from './pages/ServiceUsersPage'
import SystemSettingsPage from './pages/SystemSettingsPage'
import NotificationsPage from './pages/NotificationsPage'
import DispatcherDashboardPage from './pages/DispatcherDashboardPage'
import AdminDashboardPage from './pages/AdminDashboardPage'
import ServiceDetailPage from './pages/ServiceDetailPage'
import ExecutorDashboardPage from './pages/ExecutorDashboardPage'
import ClassificationHistoryPage from './pages/ClassificationHistoryPage'
import Layout from './components/Layout'

function App() {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <Routes>
      <Route path="/login" element={!user ? <LoginPage /> : <Navigate to="/" />} />
      <Route path="/register" element={!user ? <RegisterPage /> : <Navigate to="/" />} />
      
      <Route
        path="/"
        element={user ? <Layout /> : <Navigate to="/login" />}
      >
        <Route index element={<AppealsPage />} />
        <Route path="map" element={<AppealsMapPage />} />
        <Route path="analytics" element={<AnalyticsPage />} />
        <Route path="appeals/:id" element={<AppealDetailPage />} />
        <Route path="appeals/create" element={<CreateAppealPage />} />
        <Route path="users" element={<UsersPage />} />
        <Route path="services" element={<ServicesPage />} />
        <Route path="services/keywords" element={<ServiceKeywordsPage />} />
        <Route path="services/assignments" element={<CategoryServicesPage />} />
        <Route path="category-services" element={<CategoryServicesPage />} />
        <Route path="service-users" element={<ServiceUsersPage />} />
        <Route path="system-settings" element={<SystemSettingsPage />} />
        <Route path="profile" element={<ProfilePage />} />
        <Route path="notifications" element={<NotificationsPage />} />
        <Route path="dashboard/dispatcher" element={<DispatcherDashboardPage />} />
        <Route path="dashboard/admin" element={<AdminDashboardPage />} />
        <Route path="dashboard/admin/services/:serviceId" element={<ServiceDetailPage />} />
        <Route path="services/:serviceId" element={<ServiceDetailPage />} />
        <Route path="dashboard/executor" element={<ExecutorDashboardPage />} />
        <Route path="classifications/history" element={<ClassificationHistoryPage />} />
        <Route path="*" element={<Navigate to="/" />} />
      </Route>
    </Routes>
  )
}

export default App

