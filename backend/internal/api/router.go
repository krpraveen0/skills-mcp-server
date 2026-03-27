package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/krpraveen0/skills-mcp-server/internal/auth"
	"github.com/krpraveen0/skills-mcp-server/internal/cache"
	"github.com/krpraveen0/skills-mcp-server/internal/crawler"
	"github.com/krpraveen0/skills-mcp-server/internal/db"
	"github.com/krpraveen0/skills-mcp-server/internal/mcp"
)

// RouterDeps holds all dependencies needed to build the router.
type RouterDeps struct {
	DB             *db.DB
	Cache          *cache.Redis
	Auth           *auth.Service
	Crawler        *crawler.Crawler
	MCPServer      *mcp.Server
	CacheTTLSearch int
	CacheTTLTrend  int
	CacheTTLSkill  int
	AdminAPIKey    string // bypass key for bootstrapping the first DB key
}

// NewRouter builds and returns the Gin engine with all routes configured.
func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestIDMiddleware())
	r.Use(CORSMiddleware())

	// Initialize handlers
	skillsHandler := NewSkillsHandler(deps.DB, deps.Cache, deps.CacheTTLSearch, deps.CacheTTLTrend, deps.CacheTTLSkill)
	adminHandler  := NewAdminHandler(deps.DB, deps.Auth, deps.Crawler, deps.Cache)

	// Health check (no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "skills-mcp-server"})
	})

	// MCP endpoint (auth required)
	r.POST("/mcp", AuthMiddleware(deps.Auth, deps.AdminAPIKey), deps.MCPServer.Handle)

	// API v1 — protected routes
	v1 := r.Group("/api/v1", AuthMiddleware(deps.Auth, deps.AdminAPIKey))
	{
		// Skills
		skills := v1.Group("/skills")
		{
			skills.GET("", skillsHandler.Search)
			skills.GET("/trending", skillsHandler.Trending)
			skills.GET("/:id", skillsHandler.GetByID)
			skills.POST("/submit", skillsHandler.Submit)
		}

		// Admin routes (admin key required)
		admin := v1.Group("/admin", AdminMiddleware())
		{
			admin.GET("/stats", adminHandler.Stats)

			keys := admin.Group("/keys")
			{
				keys.GET("", adminHandler.ListKeys)
				keys.POST("", adminHandler.CreateKey)
				keys.DELETE("/:id", adminHandler.RevokeKey)
			}

			crawl := admin.Group("/crawl")
			{
				crawl.GET("/jobs", adminHandler.ListCrawlJobs)
				crawl.POST("/trigger", adminHandler.TriggerCrawl)
			}

			cache := admin.Group("/cache")
			{
				cache.POST("/flush", adminHandler.FlushCache)
			}
		}
	}

	return r
}
