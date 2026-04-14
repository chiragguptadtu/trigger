import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import LoginPage from './pages/LoginPage'
import SignupPage from './pages/SignupPage'
import CommandsPage from './pages/CommandsPage'
import AdminUsersPage from './pages/admin/AdminUsersPage'
import AdminGroupsPage from './pages/admin/AdminGroupsPage'
import AdminConfigPage from './pages/admin/AdminConfigPage'
import AdminPermissionsPage from './pages/admin/AdminPermissionsPage'
import AppLayout from './layouts/AppLayout'
import { decodeClaims } from './utils/jwt'
import { PRIMARY, HOVER_BG } from './utils/theme'

const queryClient = new QueryClient()

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token')
  return token ? <>{children}</> : <Navigate to="/login" replace />
}

function RequireAdmin({ children }: { children: React.ReactNode }) {
  const claims = decodeClaims()
  return claims?.admin ? <>{children}</> : <Navigate to="/" replace />
}

export default function App() {
  return (
    <ConfigProvider theme={{ token: { colorPrimary: PRIMARY }, components: { Table: { rowHoverBg: HOVER_BG } } }}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/signup" element={<SignupPage />} />
            <Route
              element={
                <PrivateRoute>
                  <AppLayout />
                </PrivateRoute>
              }
            >
              <Route index element={<CommandsPage />} />
              <Route path="commands/:slug" element={<CommandsPage />} />
              <Route path="admin" element={<RequireAdmin><Navigate to="/admin/users" replace /></RequireAdmin>} />
              <Route path="admin/users" element={<RequireAdmin><AdminUsersPage /></RequireAdmin>} />
              <Route path="admin/groups" element={<RequireAdmin><AdminGroupsPage /></RequireAdmin>} />
              <Route path="admin/config" element={<RequireAdmin><AdminConfigPage /></RequireAdmin>} />
              <Route path="admin/permissions" element={<RequireAdmin><AdminPermissionsPage /></RequireAdmin>} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </QueryClientProvider>
    </ConfigProvider>
  )
}
