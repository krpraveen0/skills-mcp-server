import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AppState {
  apiKey: string | null
  isAuthenticated: boolean
  setApiKey: (key: string) => void
  clearAuth: () => void
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      apiKey: null,
      isAuthenticated: false,

      setApiKey: (key: string) => {
        localStorage.setItem('skills_api_key', key)
        set({ apiKey: key, isAuthenticated: true })
      },

      clearAuth: () => {
        localStorage.removeItem('skills_api_key')
        set({ apiKey: null, isAuthenticated: false })
      },
    }),
    {
      name: 'skills-mcp-auth',
      partialize: (state) => ({ apiKey: state.apiKey, isAuthenticated: state.isAuthenticated }),
    }
  )
)
