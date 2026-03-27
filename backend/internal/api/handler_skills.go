package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krpraveen0/skills-mcp-server/internal/cache"
	"github.com/krpraveen0/skills-mcp-server/internal/db"
	"github.com/krpraveen0/skills-mcp-server/pkg/models"
)

// SkillsHandler contains handlers for the public skills API.
type SkillsHandler struct {
	db             *db.DB
	cache          *cache.Redis
	cacheTTLSearch int
	cacheTTLTrend  int
	cacheTTLSkill  int
}

// NewSkillsHandler creates a new skills handler.
func NewSkillsHandler(database *db.DB, redisCache *cache.Redis,
	ttlSearch, ttlTrend, ttlSkill int) *SkillsHandler {
	return &SkillsHandler{
		db:             database,
		cache:          redisCache,
		cacheTTLSearch: ttlSearch,
		cacheTTLTrend:  ttlTrend,
		cacheTTLSkill:  ttlSkill,
	}
}

// Search handles GET /api/v1/skills?q=&tags=&limit=&offset=&min_stars=
func (h *SkillsHandler) Search(c *gin.Context) {
	query := c.Query("q")
	limit := parseIntParam(c.Query("limit"), 10, 1, 50)
	offset := parseIntParam(c.Query("offset"), 0, 0, 10000)
	minStars := parseIntParam(c.Query("min_stars"), 0, 0, 1000000)

	var tags []string
	if t := c.Query("tags"); t != "" {
		for _, tag := range strings.Split(t, ",") {
			if trimmed := strings.TrimSpace(tag); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	cacheKey := cacheKeySearch(query, tags, limit, offset, minStars)
	var resp models.SearchResponse
	if err := h.cache.Get(c.Request.Context(), cacheKey, &resp); err == nil {
		c.JSON(http.StatusOK, resp)
		return
	}

	skills, total, err := h.db.SearchSkills(c.Request.Context(), query, tags, limit, offset, minStars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: "Search failed",
		})
		return
	}

	// Strip full content from list results
	for i := range skills {
		skills[i].Content = ""
	}

	resp = models.SearchResponse{
		Skills: skills,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	h.cache.Set(c.Request.Context(), cacheKey, resp, time.Duration(h.cacheTTLSearch)*time.Second)
	c.JSON(http.StatusOK, resp)
}

// GetByID handles GET /api/v1/skills/:id
func (h *SkillsHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400, Message: "id required"})
		return
	}

	cacheKey := "api:skill:" + id
	var skill models.Skill
	if err := h.cache.Get(c.Request.Context(), cacheKey, &skill); err == nil {
		c.JSON(http.StatusOK, skill)
		return
	}

	s, err := h.db.GetSkillByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Code: 404, Message: "Skill not found"})
		return
	}

	h.cache.Set(c.Request.Context(), cacheKey, s, time.Duration(h.cacheTTLSkill)*time.Second)
	c.JSON(http.StatusOK, s)
}

// Trending handles GET /api/v1/skills/trending?limit=&category=&min_stars=
func (h *SkillsHandler) Trending(c *gin.Context) {
	limit := parseIntParam(c.Query("limit"), 20, 1, 100)
	minStars := parseIntParam(c.Query("min_stars"), 0, 0, 1000000)
	category := c.Query("category")

	cacheKey := "api:trending:" + strconv.Itoa(limit) + ":" + strconv.Itoa(minStars) + ":" + category
	var skills []models.Skill
	if err := h.cache.Get(c.Request.Context(), cacheKey, &skills); err == nil {
		c.JSON(http.StatusOK, gin.H{"skills": skills, "count": len(skills)})
		return
	}

	skills, err := h.db.ListTrendingSkills(c.Request.Context(), limit, minStars, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: "Failed to fetch trending skills",
		})
		return
	}

	// Strip content from list
	for i := range skills {
		skills[i].Content = ""
	}

	h.cache.Set(c.Request.Context(), cacheKey, skills, time.Duration(h.cacheTTLTrend)*time.Second)
	c.JSON(http.StatusOK, gin.H{"skills": skills, "count": len(skills)})
}

// Submit handles POST /api/v1/skills/submit
func (h *SkillsHandler) Submit(c *gin.Context) {
	var req models.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}

	if !strings.HasPrefix(req.GitHubURL, "https://github.com/") {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: "URL must be a valid github.com URL",
		})
		return
	}

	// Get submitter key prefix for attribution
	keyPrefix := "anonymous"
	if key, exists := c.Get(apiKeyContextKey); exists {
		if apiKey, ok := key.(*models.APIKey); ok {
			keyPrefix = apiKey.KeyPrefix
		}
	}

	sub := &models.SkillSubmission{
		ID:          uuid.New().String(),
		GitHubURL:   req.GitHubURL,
		SubmittedBy: keyPrefix,
		SubmittedAt: time.Now(),
		Status:      "pending",
		Notes:       req.Notes,
	}

	if err := h.db.CreateSubmission(c.Request.Context(), sub); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: "Failed to save submission",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"id":      sub.ID,
		"status":  "queued",
		"message": "Your skill has been queued for indexing.",
	})
}

// --- Helpers ---

func parseIntParam(s string, def, min, max int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func cacheKeySearch(query string, tags []string, limit, offset, minStars int) string {
	return "api:search:" + query + ":" + strings.Join(tags, ",") +
		":" + strconv.Itoa(limit) + ":" + strconv.Itoa(offset) +
		":" + strconv.Itoa(minStars)
}
