package crawler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krpraveen0/skills-mcp-server/internal/cache"
	"github.com/krpraveen0/skills-mcp-server/internal/db"
	"github.com/krpraveen0/skills-mcp-server/internal/scorer"
	"github.com/krpraveen0/skills-mcp-server/pkg/models"
)

// Crawler orchestrates GitHub crawling and skill indexing.
type Crawler struct {
	db       *db.DB
	cache    *cache.Redis
	github   *GitHubClient
	scorer   *scorer.Engine
	queries  []string
	maxItems int
}

// New creates a new Crawler instance.
func New(database *db.DB, redisCache *cache.Redis, githubToken string,
	queries []string, maxItems int) *Crawler {
	return &Crawler{
		db:       database,
		cache:    redisCache,
		github:   NewGitHubClient(githubToken),
		scorer:   scorer.New(database),
		queries:  queries,
		maxItems: maxItems,
	}
}

// Run executes a full crawl cycle and returns the job record.
func (c *Crawler) Run(ctx context.Context) (*models.CrawlJob, error) {
	job := &models.CrawlJob{
		ID:        uuid.New().String(),
		StartedAt: time.Now(),
		Status:    "running",
	}

	if err := c.db.CreateCrawlJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create crawl job: %w", err)
	}

	log.Printf("[crawler] Starting crawl job %s", job.ID)

	// Run the crawl
	if err := c.runCrawl(ctx, job); err != nil {
		job.Status = "failed"
		job.Error = err.Error()
		log.Printf("[crawler] Crawl job %s failed: %v", job.ID, err)
	} else {
		job.Status = "completed"
		log.Printf("[crawler] Crawl job %s completed: found=%d new=%d updated=%d",
			job.ID, job.SkillsFound, job.SkillsNew, job.SkillsUpdated)
	}

	now := time.Now()
	job.CompletedAt = &now
	if err := c.db.UpdateCrawlJob(ctx, job); err != nil {
		log.Printf("[crawler] Failed to update crawl job: %v", err)
	}

	// Bust trending cache after every crawl
	c.cache.DeletePattern(ctx, "mcp:trending:*")
	c.cache.DeletePattern(ctx, "mcp:search:*")

	return job, nil
}

// runCrawl is the inner crawl logic.
func (c *Crawler) runCrawl(ctx context.Context, job *models.CrawlJob) error {
	var allResults []CodeSearchResult

	for _, query := range c.queries {
		log.Printf("[crawler] Searching GitHub: %s", query)
		results, err := c.github.SearchCode(ctx, query, c.maxItems/len(c.queries))
		if err != nil {
			// Log but continue with other queries
			log.Printf("[crawler] Search error for '%s': %v", query, err)
			continue
		}
		allResults = append(allResults, results...)
		job.GitHubQueries++

		// Respect GitHub search rate limit
		time.Sleep(2 * time.Second)
	}

	job.SkillsFound = len(allResults)
	log.Printf("[crawler] Found %d skills across %d queries", len(allResults), job.GitHubQueries)

	// De-duplicate by GitHub URL
	seen := map[string]bool{}
	var unique []CodeSearchResult
	for _, r := range allResults {
		if !seen[r.HTMLURL] {
			seen[r.HTMLURL] = true
			unique = append(unique, r)
		}
	}

	// Fetch content & build skill records
	var skills []models.Skill
	for i, result := range unique {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if i > 0 && i%10 == 0 {
			log.Printf("[crawler] Processing %d/%d skills...", i, len(unique))
		}

		skill, isNew, err := c.processResult(ctx, result)
		if err != nil {
			log.Printf("[crawler] Skip %s: %v", result.HTMLURL, err)
			continue
		}

		skills = append(skills, *skill)
		if isNew {
			job.SkillsNew++
		} else {
			job.SkillsUpdated++
		}

		// Small delay to avoid secondary rate limits
		time.Sleep(200 * time.Millisecond)
	}

	// Rescore all skills
	log.Printf("[crawler] Scoring %d skills...", len(skills))
	scored := c.scorer.ScoreAll(ctx, skills)

	// Persist to DB
	return c.scorer.PersistScores(ctx, scored)
}

// processResult fetches and parses a single GitHub code search result.
func (c *Crawler) processResult(ctx context.Context, r CodeSearchResult) (*models.Skill, bool, error) {
	content, lastUpdated, err := c.github.GetFileContent(ctx, r.RepoOwner, r.RepoName, r.FilePath)
	if err != nil {
		return nil, false, fmt.Errorf("fetch content: %w", err)
	}

	if len(strings.TrimSpace(content)) < 50 {
		return nil, false, fmt.Errorf("content too short, skipping")
	}

	parsed := ParseContent(content)

	skill := &models.Skill{
		ID:            uuid.New().String(),
		GitHubURL:     r.HTMLURL,
		RepoOwner:     r.RepoOwner,
		RepoName:      r.RepoName,
		FilePath:      r.FilePath,
		Content:       content,
		Title:         parsed.Title,
		Description:   parsed.Description,
		Tags:          parsed.Tags,
		Stars:         r.Stars,
		Forks:         r.Forks,
		Watchers:      r.Watchers,
		LastUpdatedAt: lastUpdated,
		IndexedAt:     time.Now(),
		IsActive:      true,
	}

	// Check if it's a new or existing skill
	isNew := true
	// The DB UpsertSkill handles deduplication by github_url

	return skill, isNew, nil
}
