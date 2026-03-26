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

// UpsertSkill inserts or updates a skill record.
func (d *DB) UpsertSkill(ctx context.Context, s *models.Skill) error {
	tagsJSON, _ := json.Marshal(s.Tags)
	scoreJSON, _ := json.Marshal(s.ScoreBreakdown)

	_, err := d.ExecContext(ctx, `
		INSERT INTO skills (
			id, github_url, repo_owner, repo_name, file_path,
			content, title, description, tags,
			stars, forks, watchers, community_refs,
			last_updated_at, indexed_at, score, score_breakdown, is_active
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
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
		FROM skills WHERE id = ? AND is_active = 1`, id)

	var s models.Skill
	var tagsJSON, scoreJSON string
	var isActive int

	err := row.Scan(
		&s.ID, &s.GitHubURL, &s.RepoOwner, &s.RepoName, &s.FilePath,
		&s.Content, &s.Title, &s.Description, &tagsJSON,
		&s.Stars, &s.Forks, &s.Watchers, &s.CommunityRefs,
		&s.LastUpdatedAt, &s.IndexedAt, &s.Score, &scoreJSON, &isActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("skill not found: %s", id)
		}
		return nil, err
	}

	s.IsActive = isActive == 1
	if err := json.Unmarshal([]byte(tagsJSON), &s.Tags); err != nil {
		s.Tags = []string{}
	}
	if err := json.Unmarshal([]byte(scoreJSON), &s.ScoreBreakdown); err != nil {
		s.ScoreBreakdown = models.ScoreBreakdown{}
	}
	return &s, nil
}

// SearchSkills performs full-text search via SQLite FTS5.
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
			SELECT s.id, s.github_url, s.repo_owner, s.repo_name, s.file_path,
			       '' as content, s.title, s.description, s.tags,
			       s.stars, s.forks, s.watchers, s.community_refs,
			       s.last_updated_at, s.indexed_at, s.score, s.score_breakdown, s.is_active
			FROM skills s
			JOIN skills_fts fts ON s.id = fts.id
			WHERE skills_fts MATCH ? AND s.is_active = 1
			ORDER BY s.score DESC
			LIMIT ? OFFSET ?`, ftsQuery(query), limit, offset)
	} else {
		rows, err = d.DB.QueryContext(ctx, `
			SELECT id, github_url, repo_owner, repo_name, file_path,
			       '' as content, title, description, tags,
			       stars, forks, watchers, community_refs,
			       last_updated_at, indexed_at, score, score_breakdown, is_active
			FROM skills WHERE is_active = 1
			ORDER BY score DESC
			LIMIT ? OFFSET ?`, limit, offset)
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
			`SELECT COUNT(*) FROM skills s JOIN skills_fts fts ON s.id = fts.id WHERE skills_fts MATCH ? AND s.is_active = 1`,
			ftsQuery(query)).Scan(&total)
	} else {
		_ = d.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM skills WHERE is_active = 1`).Scan(&total)
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
		FROM skills WHERE is_active = 1
		ORDER BY score DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSkills(rows)
}

// GetSkillStats returns aggregate stats about indexed skills.
func (d *DB) GetSkillStats(ctx context.Context) (total int, today int, err error) {
	_ = d.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM skills WHERE is_active = 1`).Scan(&total)
	_ = d.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM skills WHERE is_active = 1 AND date(indexed_at) = date('now')`).Scan(&today)
	return total, today, nil
}

// ---- API Keys ----

// CreateAPIKey inserts a new API key record.
func (d *DB) CreateAPIKey(ctx context.Context, key *models.APIKey) error {
	_, err := d.ExecContext(ctx, `
		INSERT INTO api_keys (id, key_hash, key_prefix, name, user_email, rate_limit, is_admin, created_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		key.ID, key.KeyHash, key.KeyPrefix, key.Name, key.UserEmail,
		key.RateLimit, 0, time.Now(), 1,
	)
	return err
}

// GetAPIKeyByHash looks up an API key by its SHA-256 hash.
func (d *DB) GetAPIKeyByHash(ctx context.Context, hash string) (*models.APIKey, error) {
	row := d.DB.QueryRowContext(ctx, `
		SELECT id, key_hash, key_prefix, name, user_email,
		       rate_limit, calls_today, total_calls, created_at, last_used_at, is_active
		FROM api_keys WHERE key_hash = ? AND is_active = 1`, hash)

	var k models.APIKey
	var isActive int
	if err := row.Scan(&k.ID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.UserEmail,
		&k.RateLimit, &k.CallsToday, &k.TotalCalls, &k.CreatedAt, &k.LastUsedAt, &isActive); err != nil {
		return nil, err
	}
	k.IsActive = isActive == 1
	return &k, nil
}

// IncrementAPIKeyUsage bumps the usage counters for an API key.
func (d *DB) IncrementAPIKeyUsage(ctx context.Context, id string) error {
	_, err := d.ExecContext(ctx, `
		UPDATE api_keys
		SET calls_today = calls_today + 1,
		    total_calls = total_calls + 1,
		    last_used_at = CURRENT_TIMESTAMP
		WHERE id = ?`, id)
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
		var isActive int
		if err := rows.Scan(&k.ID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.UserEmail,
			&k.RateLimit, &k.CallsToday, &k.TotalCalls, &k.CreatedAt, &k.LastUsedAt, &isActive); err != nil {
			return nil, err
		}
		k.IsActive = isActive == 1
		keys = append(keys, k)
	}
	if keys == nil {
		keys = []models.APIKey{}
	}
	return keys, nil
}

// RevokeAPIKey marks an API key as inactive.
func (d *DB) RevokeAPIKey(ctx context.Context, id string) error {
	_, err := d.ExecContext(ctx, `UPDATE api_keys SET is_active = 0 WHERE id = ?`, id)
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
		VALUES (?, ?, ?)`, job.ID, job.StartedAt, job.Status)
	return err
}

// UpdateCrawlJob updates a crawl job's progress/result.
func (d *DB) UpdateCrawlJob(ctx context.Context, job *models.CrawlJob) error {
	_, err := d.ExecContext(ctx, `
		UPDATE crawl_jobs SET
			completed_at   = ?,
			status         = ?,
			skills_found   = ?,
			skills_updated = ?,
			skills_new     = ?,
			github_queries = ?,
			error          = ?
		WHERE id = ?`,
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
		FROM crawl_jobs ORDER BY started_at DESC LIMIT ?`, limit)
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
		VALUES (?, ?, ?, ?, ?, ?)`,
		sub.ID, sub.GitHubURL, sub.SubmittedBy, sub.SubmittedAt, sub.Status, sub.Notes)
	return err
}

// ---- Helpers ----

func scanSkills(rows *sql.Rows) ([]models.Skill, error) {
	var skills []models.Skill
	for rows.Next() {
		var s models.Skill
		var tagsJSON, scoreJSON string
		var isActive int
		if err := rows.Scan(
			&s.ID, &s.GitHubURL, &s.RepoOwner, &s.RepoName, &s.FilePath,
			&s.Content, &s.Title, &s.Description, &tagsJSON,
			&s.Stars, &s.Forks, &s.Watchers, &s.CommunityRefs,
			&s.LastUpdatedAt, &s.IndexedAt, &s.Score, &scoreJSON, &isActive,
		); err != nil {
			return nil, err
		}
		s.IsActive = isActive == 1
		if err := json.Unmarshal([]byte(tagsJSON), &s.Tags); err != nil {
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

// ftsQuery sanitizes a user query for SQLite FTS5.
func ftsQuery(q string) string {
	terms := strings.Fields(q)
	quoted := make([]string, len(terms))
	for i, t := range terms {
		t = strings.ReplaceAll(t, `"`, `""`)
		quoted[i] = `"` + t + `"`
	}
	return strings.Join(quoted, " ")
}
