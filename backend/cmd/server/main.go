package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krpraveen0/skills-mcp-server/internal/api"
	"github.com/krpraveen0/skills-mcp-server/internal/auth"
	"github.com/krpraveen0/skills-mcp-server/internal/cache"
	"github.com/krpraveen0/skills-mcp-server/internal/config"
	"github.com/krpraveen0/skills-mcp-server/internal/crawler"
	"github.com/krpraveen0/skills-mcp-server/internal/db"
	"github.com/krpraveen0/skills-mcp-server/internal/mcp"
)

func main() {
	cfg := config.Load()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("[main] Starting skills-mcp-server (env=%s)", cfg.Env)

	// --- Database ---
	database, err := db.New(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("[main] Database init failed: %v", err)
	}
	defer database.Close()

	// --- Redis ---
	redisCache, err := cache.New(cfg.RedisURL, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("[main] Redis init failed: %v", err)
	}
	defer redisCache.Close()

	// --- Services ---
	authSvc := auth.NewService(database, redisCache)
	crawlerSvc := crawler.New(
		database,
		redisCache,
		cfg.GitHubToken,
		cfg.GitHubCrawlQueries,
		cfg.CrawlMaxResults,
	)
	mcpServer := mcp.NewServer(
		database, redisCache,
		cfg.CacheTTLSearch, cfg.CacheTTLTrending, cfg.CacheTTLSkill,
	)

	// --- Router ---
	router := api.NewRouter(api.RouterDeps{
		DB:             database,
		Cache:          redisCache,
		Auth:           authSvc,
		Crawler:        crawlerSvc,
		MCPServer:      mcpServer,
		CacheTTLSearch: cfg.CacheTTLSearch,
		CacheTTLTrend:  cfg.CacheTTLTrending,
		CacheTTLSkill:  cfg.CacheTTLSkill,
		AdminAPIKey:    cfg.AdminAPIKey,
	})

	// --- HTTP Server ---
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// --- Graceful shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[main] HTTP server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] Server error: %v", err)
		}
	}()

	<-quit
	log.Println("[main] Shutdown signal received, draining connections...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[main] Forced shutdown: %v", err)
	}
	log.Println("[main] Server stopped cleanly")
}
