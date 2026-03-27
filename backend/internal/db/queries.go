package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/krpraveen0/skills-mcp-server/pkg/models"
)

// ---- Skills ----

// sanitizeUTF8 strips any invalid UTF-8 bytes from s.
// PostgreSQL requires strict UTF-8; some GitHub files contain byte sequences
// from non-UTF-8 encodings (e.g. Shift-JIS) that would cause a
// "pq: invalid byte sequence" error without this guard.
func sanitizeUTF8(s string) string {
	return strings.ToValidUTF8(s, "")
}

// UpsertSkill inserts or updates a skill record.
func (d *DB) UpsertSkill(ctx context.Context, s *models.Skill) error {
	// Sanitize all free-text fields before hitting PostgreSQL's strict UTF-8
	s.Content = sanitizeUTF8(s.Content)
	s.Title = sanitizeUTF8(s.Title)
	s.Description = sanitizeUTF8(s.Description)
	s.FilePath = sanitizeUTF8(s.FilePath)
	for i, tag := range s.Tags {
		s.Tags[i] = sanitizeUTF8(tag)
	}

	tagsJSON, _ := json.Marshal(s.Tags)
	scoreJSON, _ := json.Marshal(s.ScoreBreakdown)

	_, err := d.ExecContext(ctx, `
		INSERT INTO skills (
			id, github_url, repo_owner, repo_name, file_path,
			content, title, description, tags,
			stars, forks, watchers, community_refs,
			last_updated_at, indexed_at, score, score_breakdown, is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		ON CONFLICT(github_url) DO UPDATE SET
			content        = excluded.content,
			title          = excluded.title,
			description    = excluded.description,
			tags           = excluded.tags,
			stars          = excluded.stars,
			forks          = excluded.forks,
			watchers       = excluded.watchers,
			community_refs = excluded.community_refs,
			last_updated_at= excluded.last_updated_at,
			indexed_at     = excluded.indexed_at,
			score          = excluded.score,
			score_breakdown= excluded.score_breakdown,
			is_active      = excluded.is_active`,
		s.ID, s.GitHubURL, s.RepoOwner, s.RepoName, s.FilePath,
		s.Content, s.Title, s.Description, string(tagsJSON),
		s.Stars, s.Forks, s.Watchers, s.CommunityRefs,
		s.LastUpdatedAt, s.IndexedAt, s.Score, string(scoreJSON), s.IsActive,
	)
	return err
}

// GetSkillByID returns a single skill by its UUID.
func (d *DB) GetSkillByID(ctx context.Context, id string) (*models.Skill, error) {
	row := d.DB.QueryRowContext(ctx, `
		SELECT id, github_url, repo_owner, repo_name, file_path,
		       content, title, description, tags,
		       stars, forks, watchers, community_refs,
		       last_updated_at, indexed_at, score, score_breakdown, is_active
		FROM skills WHERE id = $1 AND is_active = TRUE`, id)

	var s models.Skill
	var tagsJSON, scoreJSON string

	err := row.Scan(
		&s.ID, &s.GitHubURL, &s.RepoOwner, &s.RepoName, &s.FilePath,
		&s.Content, &s.Title, &s.Description, &tagsJSON,
		&s.Stars, &s.Forks, &s.Watchers, &s.CommunityRefs,
		&s.LastUpdatedAt, &s.IndexedAt, &s.Score, &scoreJSON, &s.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("skill not found: %s", id)
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &s.Tags); err != nil || s.Tags == nil {
		s.Tags = []string{}
	}
	if err := json.Unmarshal([]byte(scoreJSON), &s.ScoreBreakdown); err != nil {
		s.ScoreBreakdown = models.ScoreBreakdown{}
	}
	return &s, nil
}

// SearchSkills performs full-text search via PostgreSQL tsvector.
func (d *DB) SearchSkills(ctx context.Context, query string, _ []string, limit, offset int) ([]models.Skill, int, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	var (
		rows *sql.Rows
		err  error
	)

	if query != "" {
		rows, err = d.DB.QueryContext(ctx, `
			SELECT id, github_url, repo_owner, repo_name, file_path,
			       '' as content, title, description, tags,
			       stars, forks, watchers, community_refs,
			       last_updated_at, indexed_at, score, score_breakdown, is_active
			FROM skills
			WHERE search_vector @@ plainto_tsquery('english', $1) AND is_active = TRUE
			ORDER BY score DESC
			LIMIT $2 OFFSET $3`, query, limit, offset)
	} else {
		rows, err = d.DB.QueryContext(ctx, `
			SELECT id, github_url, repo_owner, repo_name, file_path,
			       '' as content, title, description, tags,
			       stars, forks, watchers, community_refs,
			       last_updated_at, indexed_at, score, score_breakdown, is_active
			FROM skills WHERE is_active = TRUE
			ORDER BY score DESC
			LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	skills, err := scanSkills(rows)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if query != "" {
		_ = d.DB.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM skills WHERE search_vector @@ plainto_tsquery('english', $1) AND is_active = TRUE`,
			query).Scan(&total)
	} else {
		_ = d.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM skills WHERE is_active = TRUE`).Scan(&total)
	}

	return skills, total, nil
}

// ListTrendingSkills returns the top-N skills by composite score.
func (d *DB) ListTrendingSkills(ctx context.Context, limit int, _ string) ([]models.Skill, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := d.DB.QueryContext(ctx, `
		SELECT id, github_url, repo_owner, repo_name, file_path,
		       '' as content, title, description, tags,
		       stars, forks, watchers, community_refs,
		       last_updated_at, indexed_at, score, score_breakdown, is_active
		FROM skills WHERE is_active = TRUE
		ORDER BY score DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSkills(rows)
}

// GetSkillStats returns aggregate stats about indexed skills.
func (d *DB) GetSkillStats(ctx context.Context) (total int, today int, err error) {
	_ = d.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM skills WHERE is_active = TRUE`).Scan(&total)
	_ = d.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM skills WHERE is_active = TRUE AND indexed_at::date = CURRENT_DATE`).Scan(&today)
	return total, today, nil
}

// ---- API Keys ----

// CreateAPIKey inserts a new API key record.
func (d *DB) CreateAPIKey(ctx context.Context, key *models.APIKey) error {
	_, err := d.ExecContext(ctx, `
		INSERT INTO api_keys (id, key_hash, key_prefix, name, user_email, rate_limit, is_admin, created_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		key.ID, key.KeyHash, key.KeyPrefix, key.Name, key.UserEmail,
		key.RateLimit, false, time.Now(), true,
	)
	return err
}

// GetAPIKeyByHash looks up an API key by its SHA-256 hash.
func (d *DB) GetAPIKeyByHash(ctx context.Context, hash string) (*models.APIKey, error) {
	row := d.DB.QueryRowContext(ctx, `
		SELECT id, key_hash, key_prefix, name, user_email,
		       rate_limit, calls_today, total_calls, created_at, last_used_at, is_active
		FROM api_keys WHERE key_hash = $1 AND is_active = TRUE`, hash)

	var k models.APIKey
	if err := row.Scan(&k.ID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.UserEmail,
		&k.RateLimit, &k.CallsToday, &k.TotalCalls, &k.CreatedAt, &k.LastUsedAt, &k.IsActive); err != nil {
		return nil, err
	}
	return &k, nil
}

// IncrementAPIKeyUsage bumps the usage counters for an API key.
func (d *DB) IncrementAPIKeyUsage(ctx context.Context, id string) error {
	_, err := d.ExecContext(ctx, `
		UPDATE api_keys
		SET calls_today = calls_today + 1,
		    total_calls = total_calls + 1,
		    last_used_at = NOW()
		WHERE id = $1`, id)
	return err
}

// ListAPIKeys returns all API keys for admin display.
func (d *DB) ListAPIKeys(ctx context.Context) ([]models.APIKey, error) {
	rows, err := d.DB.QueryContext(ctx, `
		SELECT id, key_hash, key_prefix, name, user_email,
		       rate_limit, calls_today, total_calls, created_at, last_used_at, is_active
		FROM api_keys ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.UserEmail,
			&k.RateLimit, &k.CallsToday, &k.TotalCalls, &k.CreatedAt, &k.LastUsedAt, &k.IsActive); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	if keys == nil {
		keys = []models.APIKey{}
	}
	return keys, nil
}

// RevokeAPIKey marks an API key as inactive.
func (d *DB) RevokeAPIKey(ctx context.Context, id string) error {
	_, err := d.ExecContext(ctx, `UPDATE api_keys SET is_active = FALSE WHERE id = $1`, id)
	return err
}

// ResetDailyCallCounts zeroes calls_today for all keys.
func (d *DB) ResetDailyCallCounts(ctx context.Context) error {
	_, err := d.ExecContext(ctx, `UPDATE api_keys SET calls_today = 0`)
	return err
}

// ---- Crawl Jobs ----

// CreateCrawlJob inserts a new crawl job record.
func (d *DB) CreateCrawlJob(ctx context.Context, job *models.CrawlJob) error {
	_, err := d.ExecContext(ctx, `
		INSERT INTO crawl_jobs (id, started_at, status)
		VALUES ($1, $2, $3)`, job.ID, job.StartedAt, job.Status)
	return err
}

// UpdateCrawlJob updates a crawl job's progress/result.
func (d *DB) UpdateCrawlJob(ctx context.Context, job *models.CrawlJob) error {
	_, err := d.ExecContext(ctx, `
		UPDATE crawl_jobs SET
			completed_at   = $1,
			status         = $2,
			skills_found   = $3,
			skills_updated = $4,
			skills_new     = $5,
			github_queries = $6,
			error          = $7
		WHERE id = $8`,
		job.CompletedAt, job.Status,
		job.SkillsFound, job.SkillsUpdated, job.SkillsNew,
		job.GitHubQueries, job.Error, job.ID,
	)
	return err
}

// ListCrawlJobs returns the most recent crawl jobs.
func (d *DB) ListCrawlJobs(ctx context.Context, limit int) ([]models.CrawlJob, error) {
	rows, err := d.DB.QueryContext(ctx, `
		SELECT id, started_at, completed_at, status,
		       skills_found, skills_updated, skills_new, github_queries, error
		FROM crawl_jobs ORDER BY started_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.CrawlJob
	for rows.Next() {
		var j models.CrawlJob
		if err := rows.Scan(&j.ID, &j.StartedAt, &j.CompletedAt, &j.Status,
			&j.SkillsFound, &j.SkillsUpdated, &j.SkillsNew, &j.GitHubQueries, &j.Error); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	if jobs == nil {
		jobs = []models.CrawlJob{}
	}
	return jobs, nil
}

// ---- Submissions ----

// CreateSubmission stores a user-submitted skill URL.
func (d *DB) CreateSubmission(ctx context.Context, sub *models.SkillSubmission) error {
	_, err := d.ExecContext(ctx, `
		INSERT INTO skill_submissions (id, github_url, submitted_by, submitted_at, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		sub.ID, sub.GitHubURL, sub.SubmittedBy, sub.SubmittedAt, sub.Status, sub.Notes)
	return err
}

// ---- Helpers ----

func scanSkills(rows *sql.Rows) ([]models.Skill, error) {
	var skills []models.Skill
	for rows.Next() {
		var s models.Skill
		var tagsJSON, scoreJSON string
		if err := rows.Scan(
			&s.ID, &s.GitHubURL, &s.RepoOwner, &s.RepoName, &s.FilePath,
			&s.Content, &s.Title, &s.Description, &tagsJSON,
			&s.Stars, &s.Forks, &s.Watchers, &s.CommunityRefs,
			&s.LastUpdatedAt, &s.IndexedAt, &s.Score, &scoreJSON, &s.IsActive,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(tagsJSON), &s.Tags); err != nil || s.Tags == nil {
			s.Tags = []string{}
		}
		if err := json.Unmarshal([]byte(scoreJSON), &s.ScoreBreakdown); err != nil {
			s.ScoreBreakdown = models.ScoreBreakdown{}
		}
		skills = append(skills, s)
	}
	if skills == nil {
		skills = []models.Skill{}
	}
	return skills, rows.Err()
}
