package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DatabaseURL string

	// Redis
	RedisURL      string
	RedisPassword string

	// GitHub
	GitHubToken        string
	GitHubCrawlQueries []string

	// Auth
	AdminAPIKey string
	APIKeySalt  string

	// Cache TTLs (seconds)
	CacheTTLSearch   int
	CacheTTLTrending int
	CacheTTLSkill    int

	// Crawler
	CrawlSchedule    string
	CrawlMaxResults  int
}

// Load reads configuration from environment variables.
// It optionally loads a .env file in development.
func Load() *Config {
	if env := os.Getenv("ENV"); env != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("[config] No .env file found, using environment variables")
		}
	}

	return &Config{
		Port:       getEnv("PORT", "8080"),
		Env:        getEnv("ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/skills?sslmode=disable"),

		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		GitHubToken: getEnv("GITHUB_TOKEN", ""),
		GitHubCrawlQueries: []string{
			"filename:SKILL.md",
			"filename:skills.md",
		},

		AdminAPIKey: getEnv("ADMIN_API_KEY", ""),
		APIKeySalt:  getEnv("API_KEY_SALT", "change-me-in-production"),

		CacheTTLSearch:   getEnvInt("CACHE_TTL_SEARCH", 3600),
		CacheTTLTrending: getEnvInt("CACHE_TTL_TRENDING", 86400),
		CacheTTLSkill:    getEnvInt("CACHE_TTL_SKILL", 3600),

		CrawlSchedule:   getEnv("CRAWL_SCHEDULE", "0 2 * * *"),
		CrawlMaxResults: getEnvInt("CRAWL_MAX_RESULTS", 1000),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultValue
}
