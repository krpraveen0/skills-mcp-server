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

// Global error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Clear invalid key and redirect to auth
      localStorage.removeItem('skills_api_key')
      window.location.href = '/auth'
    }
    return Promise.reject(error)
  }
)
