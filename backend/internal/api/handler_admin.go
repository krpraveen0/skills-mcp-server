package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/krpraveen0/skills-mcp-server/internal/auth"
	"github.com/krpraveen0/skills-mcp-server/internal/crawler"
	"github.com/krpraveen0/skills-mcp-server/internal/db"
	"github.com/krpraveen0/skills-mcp-server/pkg/models"
)

// AdminHandler contains handlers for the admin API.
type AdminHandler struct {
	db         *db.DB
	authSvc    *auth.Service
	crawlerSvc *crawler.Crawler
	cache      cacheFlushable
}

// cacheFlushable is the subset of cache.Redis needed by AdminHandler.
type cacheFlushable interface {
	DeletePattern(ctx context.Context, pattern string) error
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(database *db.DB, authSvc *auth.Service, crawlerSvc *crawler.Crawler, c cacheFlushable) *AdminHandler {
	return &AdminHandler{db: database, authSvc: authSvc, crawlerSvc: crawlerSvc, cache: c}
}

// Stats handles GET /api/v1/admin/stats
func (h *AdminHandler) Stats(c *gin.Context) {
	ctx := c.Request.Context()

	totalSkills, todaySkills, err := h.db.GetSkillStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: err.Error(),
		})
		return
	}

	jobs, _ := h.db.ListCrawlJobs(ctx, 1)
	lastCrawlStatus := "never"
	var lastCrawlAt interface{} = nil
	if len(jobs) > 0 {
		lastCrawlAt = jobs[0].CompletedAt
		lastCrawlStatus = jobs[0].Status
	}

	keys, _ := h.db.ListAPIKeys(ctx)

	c.JSON(http.StatusOK, gin.H{
		"total_skills":       totalSkills,
		"total_api_keys":     len(keys),
		"last_crawl_at":      lastCrawlAt,
		"last_crawl_status":  lastCrawlStatus,
		"skills_added_today": todaySkills,
		"top_tags":           []interface{}{},
	})
}

// ListKeys handles GET /api/v1/admin/keys
func (h *AdminHandler) ListKeys(c *gin.Context) {
	keys, err := h.db.ListAPIKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"keys": keys, "count": len(keys)})
}

// CreateKey handles POST /api/v1/admin/keys
func (h *AdminHandler) CreateKey(c *gin.Context) {
	var req models.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}

	resp, err := h.authSvc.GenerateKey(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// RevokeKey handles DELETE /api/v1/admin/keys/:id
func (h *AdminHandler) RevokeKey(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.RevokeAPIKey(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Key revoked"})
}

// ListCrawlJobs handles GET /api/v1/admin/crawl/jobs
func (h *AdminHandler) ListCrawlJobs(c *gin.Context) {
	limit := parseIntParam(c.Query("limit"), 20, 1, 100)
	jobs, err := h.db.ListCrawlJobs(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobs": jobs, "count": len(jobs)})
}

// FlushCache handles POST /api/v1/admin/cache/flush
// Evicts all api:search:* and api:trending:* keys so the Explorer shows
// fresh data without waiting for the next crawl or the TTL to expire.
func (h *AdminHandler) FlushCache(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.cache.DeletePattern(ctx, "api:search:*"); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: "flush search cache: " + err.Error(),
		})
		return
	}
	if err := h.cache.DeletePattern(ctx, "api:trending:*"); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error", Code: 500, Message: "flush trending cache: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cache flushed. Search and trending results will be refreshed from DB."})
}

// TriggerCrawl handles POST /api/v1/admin/crawl/trigger
// NOTE: must use context.Background() — c.Request.Context() is cancelled
// the moment the HTTP response is written, which would abort the crawl.
func (h *AdminHandler) TriggerCrawl(c *gin.Context) {
	// 30-minute timeout for a full crawl cycle
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	go func() {
		defer cancel()
		_, _ = h.crawlerSvc.Run(ctx)
	}()
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Crawl job triggered and running in the background.",
	})
}
