import axios from 'axios'

const BASE_URL = import.meta.env.VITE_API_URL || ''

export const apiClient = axios.create({
  baseURL: BASE_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Inject API key from local storage on every request
apiClient.interceptors.request.use((config) => {
  const apiKey = localStorage.getItem('skills_api_key')
  if (apiKey) {
    config.headers.Authorization = `Bearer ${apiKey}`
  }
  return config
})

// Global error handling — only redirect on 401 if the user was authenticated.
// Anonymous requests to public endpoints should NOT trigger a login redirect.
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const wasAuthenticated = Boolean(localStorage.getItem('skills_api_key'))
      if (wasAuthenticated) {
        localStorage.removeItem('skills_api_key')
        window.location.href = '/auth'
      }
    }
    return Promise.reject(error)
  }
)
