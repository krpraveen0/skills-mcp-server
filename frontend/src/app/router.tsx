import type { ReactNode } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'
import { App } from './App'
import { ExplorerPage } from '@/pages/explorer/ExplorerPage'
import { AdminDashboard } from '@/pages/admin/AdminDashboard'
import { LoginPage } from '@/pages/auth/LoginPage'
import { useAppStore } from '@/store/useAppStore'

// Auth guard — redirects to /auth if not logged in
function RequireAuth({ children }: { children: ReactNode }) {
  const isAuthenticated = useAppStore((s) => s.isAuthenticated)
  if (!isAuthenticated) return <Navigate to="/auth" replace />
  return <>{children}</>
}

export const router = createBrowserRouter([
  {
    path: '/auth',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: (
      <RequireAuth>
        <App />
      </RequireAuth>
    ),
    children: [
      { index: true, element: <ExplorerPage /> },
      { path: 'admin', element: <AdminDashboard /> },
    ],
  },
])
