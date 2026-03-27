package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/krpraveen0/skills-mcp-server/internal/cache"
	"github.com/krpraveen0/skills-mcp-server/internal/config"
	"github.com/krpraveen0/skills-mcp-server/internal/crawler"
	"github.com/krpraveen0/skills-mcp-server/internal/db"
)

func main() {
	cfg := config.Load()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("[worker] Starting crawl worker (env=%s)", cfg.Env)

	// --- Database ---
	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[worker] Database init failed: %v", err)
	}
	defer database.Close()

	// --- Redis ---
	redisCache, err := cache.New(cfg.RedisURL, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("[worker] Redis init failed: %v", err)
	}
	defer redisCache.Close()

	// --- Crawler ---
	crawlerSvc := crawler.New(
		database,
		redisCache,
		cfg.GitHubToken,
		cfg.GitHubCrawlQueries,
		cfg.CrawlMaxResults,
	)

	// Run once immediately on startup
	go func() {
		log.Printf("[worker] Running initial crawl on startup...")
		if _, err := crawlerSvc.Run(context.Background()); err != nil {
			log.Printf("[worker] Initial crawl failed: %v", err)
		}
	}()

	// Schedule recurring crawls
	c := cron.New(cron.WithLocation(nil))
	c.AddFunc(cfg.CrawlSchedule, func() {
		log.Printf("[worker] Scheduled crawl triggered")
		if _, err := crawlerSvc.Run(context.Background()); err != nil {
			log.Printf("[worker] Scheduled crawl failed: %v", err)
		}

		// Reset daily API key counters at midnight
		database.ResetDailyCallCounts(context.Background())
	})

	c.Start()
	log.Printf("[worker] Cron scheduler started with schedule: %s", cfg.CrawlSchedule)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[worker] Shutting down...")
	ctx := c.Stop()
	<-ctx.Done()
	log.Println("[worker] Worker stopped cleanly")
}
