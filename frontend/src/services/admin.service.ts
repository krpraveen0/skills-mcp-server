import { apiClient } from './api'

export interface APIKey {
  id: string
  key_prefix: string
  name: string
  user_email: string
  rate_limit: number
  calls_today: number
  total_calls: number
  created_at: string
  last_used_at: string | null
  is_active: boolean
  raw_key?: string // Only present on creation
}

export interface CrawlJob {
  id: string
  started_at: string
  completed_at: string | null
  status: 'pending' | 'running' | 'completed' | 'failed'
  skills_found: number
  skills_updated: number
  skills_new: number
  github_queries: number
  error: string
}

export interface AdminStats {
  total_skills: number
  total_api_keys: number
  last_crawl_at: string | null
  last_crawl_status: string
  skills_added_today: number
  top_tags: { tag: string; count: number }[]
}

export const adminService = {
  getStats: async (): Promise<AdminStats> => {
    const { data } = await apiClient.get<AdminStats>('/api/v1/admin/stats')
    return data
  },

  listKeys: async (): Promise<{ keys: APIKey[]; count: number }> => {
    const { data } = await apiClient.get('/api/v1/admin/keys')
    return data
  },

  createKey: async (name: string, email: string, rateLimit?: number): Promise<APIKey> => {
    const { data } = await apiClient.post<APIKey>('/api/v1/admin/keys', {
      name,
      email,
      rate_limit: rateLimit,
    })
    return data
  },

  revokeKey: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/v1/admin/keys/${id}`)
  },

  listCrawlJobs: async (limit = 20): Promise<{ jobs: CrawlJob[]; count: number }> => {
    const { data } = await apiClient.get('/api/v1/admin/crawl/jobs', { params: { limit } })
    return data
  },

  triggerCrawl: async (): Promise<{ message: string }> => {
    const { data } = await apiClient.post('/api/v1/admin/crawl/trigger')
    return data
  },
}
