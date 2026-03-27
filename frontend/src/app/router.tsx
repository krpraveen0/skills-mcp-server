import type { ReactNode } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'
import { App } from './App'
import { ExplorerPage } from '@/pages/explorer/ExplorerPage'
import { AdminDashboard } from '@/pages/admin/AdminDashboard'
import { LoginPage } from '@/pages/auth/LoginPage'
import { RegisterPage } from '@/pages/auth/RegisterPage'
import { SkillDetailPage } from '@/pages/skill/SkillDetailPage'
import { useAppStore } from '@/store/useAppStore'

// Admin-only guard — redirects to / if not an admin
function RequireAdmin({ children }: { children: ReactNode }) {
  const { isAuthenticated, isAdmin } = useAppStore((s) => ({
    isAuthenticated: s.isAuthenticated,
    isAdmin: s.isAdmin,
  }))
  if (!isAuthenticated) return <Navigate to="/auth" replace />
  if (!isAdmin) return <Navigate to="/" replace />
  return <>{children}</>
}

export const router = createBrowserRouter([
  {
    path: '/auth',
    element: <LoginPage />,
  },
  {
    path: '/register',
    element: <RegisterPage />,
  },
  {
    // App shell is always accessible (public browsing)
    path: '/',
    element: <App />,
    children: [
      { index: true, element: <ExplorerPage /> },
      { path: 'skills/:id', element: <SkillDetailPage /> },
      {
        path: 'admin',
        element: (
          <RequireAdmin>
            <AdminDashboard />
          </RequireAdmin>
        ),
      },
    ],
  },
])
