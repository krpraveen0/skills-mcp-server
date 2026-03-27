import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AppState {
  apiKey: string | null
  isAuthenticated: boolean
  isAdmin: boolean
  setApiKey: (key: string, isAdmin?: boolean) => void
  clearAuth: () => void
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      apiKey: null,
      isAuthenticated: false,
      isAdmin: false,

      setApiKey: (key: string, isAdmin = false) => {
        localStorage.setItem('skills_api_key', key)
        set({ apiKey: key, isAuthenticated: true, isAdmin })
      },

      clearAuth: () => {
        localStorage.removeItem('skills_api_key')
        set({ apiKey: null, isAuthenticated: false, isAdmin: false })
      },
    }),
    {
      name: 'skills-mcp-auth',
      partialize: (state) => ({
        apiKey: state.apiKey,
        isAuthenticated: state.isAuthenticated,
        isAdmin: state.isAdmin,
      }),
    }
  )
)
