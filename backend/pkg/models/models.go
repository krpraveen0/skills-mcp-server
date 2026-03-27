package models

import "time"

// Skill represents an indexed SKILL.md file from GitHub.
type Skill struct {
	ID             string         `json:"id"`
	GitHubURL      string         `json:"github_url"`
	RepoOwner      string         `json:"repo_owner"`
	RepoName       string         `json:"repo_name"`
	FilePath       string         `json:"file_path"`
	Content        string         `json:"content,omitempty"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Tags           []string       `json:"tags"`
	Stars          int            `json:"stars"`
	Forks          int            `json:"forks"`
	Watchers       int            `json:"watchers"`
	CommunityRefs  int            `json:"community_refs"`
	LastUpdatedAt  *time.Time     `json:"last_updated_at"`
	IndexedAt      time.Time      `json:"indexed_at"`
	Score          float64        `json:"score"`
	ScoreBreakdown ScoreBreakdown `json:"score_breakdown"`
	IsActive       bool           `json:"is_active"`
}

// ScoreBreakdown holds the individual scoring components.
type ScoreBreakdown struct {
	StarScore      float64 `json:"star_score"`
	AdoptionScore  float64 `json:"adoption_score"`
	RecencyScore   float64 `json:"recency_score"`
	CompositeScore float64 `json:"composite_score"`
}

// APIKey represents an authentication key.
type APIKey struct {
	ID         string     `json:"id"`
	KeyHash    string     `json:"-"`
	KeyPrefix  string     `json:"key_prefix"`
	Name       string     `json:"name"`
	UserEmail  string     `json:"user_email"`
	RateLimit  int        `json:"rate_limit"`
	CallsToday int        `json:"calls_today"`
	TotalCalls int        `json:"total_calls"`
	IsAdmin    bool       `json:"is_admin,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	IsActive   bool       `json:"is_active"`
}

// CrawlJob tracks a GitHub crawl run.
type CrawlJob struct {
	ID            string     `json:"id"`
	StartedAt     time.Time  `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	Status        string     `json:"status"`
	SkillsFound   int        `json:"skills_found"`
	SkillsUpdated int        `json:"skills_updated"`
	SkillsNew     int        `json:"skills_new"`
	GitHubQueries int        `json:"github_queries"`
	Error         string     `json:"error,omitempty"`
}

// SkillSubmission is a user-submitted URL.
type SkillSubmission struct {
	ID          string    `json:"id"`
	GitHubURL   string    `json:"github_url"`
	SubmittedBy string    `json:"submitted_by"`
	SubmittedAt time.Time `json:"submitted_at"`
	Status      string    `json:"status"`
	Notes       string    `json:"notes"`
}

// --- Request / Response types ---

// TrendingRepo aggregates skills from a single GitHub repository.
type TrendingRepo struct {
	Owner         string     `json:"owner"`
	Name          string     `json:"name"`
	GitHubURL     string     `json:"github_url"`
	Stars         int        `json:"stars"`
	Forks         int        `json:"forks"`
	Watchers      int        `json:"watchers"`
	SkillCount    int        `json:"skill_count"`
	TopScore      float64    `json:"top_score"`
	LastUpdatedAt *time.Time `json:"last_updated_at"`
	IndexedAt     time.Time  `json:"indexed_at"`
	Description   string     `json:"description"`
	Tags          []string   `json:"tags"`
}

// SearchRequest is the payload for searching skills.
type SearchRequest struct {
	Query    string   `json:"query"    form:"q"`
	Tags     []string `json:"tags"     form:"tags"`
	Limit    int      `json:"limit"    form:"limit"`
	Offset   int      `json:"offset"   form:"offset"`
	MinStars int      `json:"min_stars" form:"min_stars"`
}

// SearchResponse wraps a paginated list of skills.
type SearchResponse struct {
	Skills []Skill `json:"skills"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

// SubmitRequest is the payload for submitting a skill URL.
type SubmitRequest struct {
	GitHubURL string `json:"github_url" binding:"required"`
	Notes     string `json:"notes"`
}

// CreateAPIKeyRequest is the payload for creating a new API key.
type CreateAPIKeyRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email"`
	RateLimit int    `json:"rate_limit"`
	IsAdmin   bool   `json:"is_admin"`
}

// RegisterRequest is the self-service payload for creating a public API key.
type RegisterRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

// CreateAPIKeyResponse includes the raw key (shown once).
type CreateAPIKeyResponse struct {
	APIKey
	RawKey string `json:"raw_key"`
}

// AdminStats is the response for the admin stats endpoint.
type AdminStats struct {
	TotalSkills      int        `json:"total_skills"`
	TotalAPIKeys     int        `json:"total_api_keys"`
	LastCrawlAt      *time.Time `json:"last_crawl_at"`
	LastCrawlStatus  string     `json:"last_crawl_status"`
	SkillsAddedToday int        `json:"skills_added_today"`
	TopTags          []TagCount `json:"top_tags"`
}

// TagCount is a tag with its occurrence count.
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// ErrorResponse is the standard JSON error body.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
