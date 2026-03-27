package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/krpraveen0/skills-mcp-server/internal/cache"
	"github.com/krpraveen0/skills-mcp-server/internal/db"
	"github.com/krpraveen0/skills-mcp-server/pkg/models"
)

// ReposHandler handles repository-level endpoints.
type ReposHandler struct {
	db    *db.DB
	cache *cache.Redis
	ttl   int // cache TTL in seconds
}

// NewReposHandler creates a new repos handler.
func NewReposHandler(database *db.DB, redisCache *cache.Redis, ttl int) *ReposHandler {
	return &ReposHandler{db: database, cache: redisCache, ttl: ttl}
}

// TrendingRepos handles GET /api/v1/repos/trending
// Query params:
//   - period: "today" | "week" | "month" | "all" (default: "week")
//   - min_stars: integer (default: 100)
//   - limit: 1-50 (default: 10)
func (h *ReposHandler) TrendingRepos(c *gin.Context) {
	period := c.DefaultQuery("period", "week")
	minStars := parseIntParam(c.Query("min_stars"), 100, 0, 1000000)
	limit := parseIntParam(c.Query("limit"), 10, 1, 50)

	// Validate period
	validPeriods := map[string]bool{"today": true, "week": true, "month": true, "all": true}
	if !validPeriods[period] {
		period = "week"
	}

	cacheKey := "api:repos:trending:" + period + ":" + c.Query("min_stars") + ":" + c.Query("limit")
	var repos []models.TrendingRepo
	if err := h.cache.Get(c.Request.Context(), cacheKey, &repos); err == nil {
		c.JSON(http.StatusOK, gin.H{"repos": repos, "count": len(repos), "period": period})
		return
	}

	repos, err := h.db.GetTrendingRepos(c.Request.Context(), period, minStars, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: "Failed to fetch trending repos",
		})
		return
	}

	// Cache for 6 hours — refreshed after each daily crawl via cache flush
	h.cache.Set(c.Request.Context(), cacheKey, repos, time.Duration(h.ttl)*time.Second)
	c.JSON(http.StatusOK, gin.H{"repos": repos, "count": len(repos), "period": period})
}

// GetRepo handles GET /api/v1/repos/:owner/:repo
// Returns the repo metadata (from its best skill) + all skills in that repo.
func (h *ReposHandler) GetRepo(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: "owner and repo are required",
		})
		return
	}

	cacheKey := "api:repos:detail:" + owner + ":" + repo
	var payload struct {
		Owner      string         `json:"owner"`
		Name       string         `json:"name"`
		GitHubURL  string         `json:"github_url"`
		Stars      int            `json:"stars"`
		Forks      int            `json:"forks"`
		Watchers   int            `json:"watchers"`
		SkillCount int            `json:"skill_count"`
		Skills     []models.Skill `json:"skills"`
	}
	if err := h.cache.Get(c.Request.Context(), cacheKey, &payload); err == nil {
		c.JSON(http.StatusOK, payload)
		return
	}

	skills, err := h.db.GetRepoSkills(c.Request.Context(), owner, repo)
	if err != nil || len(skills) == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "not_found", Code: 404,
			Message: "Repository not found or has no indexed skills",
		})
		return
	}

	// Derive repo-level stats from the first (highest-score) skill
	top := skills[0]
	payload = struct {
		Owner      string         `json:"owner"`
		Name       string         `json:"name"`
		GitHubURL  string         `json:"github_url"`
		Stars      int            `json:"stars"`
		Forks      int            `json:"forks"`
		Watchers   int            `json:"watchers"`
		SkillCount int            `json:"skill_count"`
		Skills     []models.Skill `json:"skills"`
	}{
		Owner:      top.RepoOwner,
		Name:       top.RepoName,
		GitHubURL:  "https://github.com/" + top.RepoOwner + "/" + top.RepoName,
		Stars:      top.Stars,
		Forks:      top.Forks,
		Watchers:   top.Watchers,
		SkillCount: len(skills),
		Skills:     skills,
	}

	h.cache.Set(c.Request.Context(), cacheKey, payload, time.Duration(h.ttl)*time.Second)
	c.JSON(http.StatusOK, payload)
}
