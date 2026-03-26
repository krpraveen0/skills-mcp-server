import { apiClient } from './api'

export interface Skill {
  id: string
  github_url: string
  repo_owner: string
  repo_name: string
  file_path: string
  content?: string
  title: string
  description: string
  tags: string[]
  stars: number
  forks: number
  watchers: number
  community_refs: number
  last_updated_at: string | null
  indexed_at: string
  score: number
  score_breakdown: {
    star_score: number
    adoption_score: number
    recency_score: number
    composite_score: number
  }
}

export interface SearchResponse {
  skills: Skill[]
  total: number
  limit: number
  offset: number
}

export interface TrendingResponse {
  skills: Skill[]
  count: number
}

export const skillsService = {
  search: async (query: string, tags?: string[], limit = 10, offset = 0): Promise<SearchResponse> => {
    const params: Record<string, string | number> = { limit, offset }
    if (query) params.q = query
    if (tags?.length) params.tags = tags.join(',')
    const { data } = await apiClient.get<SearchResponse>('/api/v1/skills', { params })
    return data
  },

  getById: async (id: string): Promise<Skill> => {
    const { data } = await apiClient.get<Skill>(`/api/v1/skills/${id}`)
    return data
  },

  getTrending: async (limit = 20, category?: string): Promise<TrendingResponse> => {
    const params: Record<string, string | number> = { limit }
    if (category) params.category = category
    const { data } = await apiClient.get<TrendingResponse>('/api/v1/skills/trending', { params })
    return data
  },

  submit: async (githubUrl: string, notes?: string): Promise<{ id: string; status: string; message: string }> => {
    const { data } = await apiClient.post('/api/v1/skills/submit', { github_url: githubUrl, notes })
    return data
  },
}
