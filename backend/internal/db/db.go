package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// DB wraps a sql.DB connection with helper methods.
type DB struct {
	*sql.DB
}

// migration holds a numbered SQL statement to apply once.
type migration struct {
	version int
	sql     string
}

// migrations are applied in order; every statement uses IF NOT EXISTS so they
// are safe to re-run. A schema_migrations table tracks what has been applied.
var migrations = []migration{
	{1, `
CREATE TABLE IF NOT EXISTS skills (
    id              TEXT PRIMARY KEY,
    github_url      TEXT UNIQUE NOT NULL,
    repo_owner      TEXT NOT NULL,
    repo_name       TEXT NOT NULL,
    file_path       TEXT NOT NULL,
    content         TEXT NOT NULL,
    title           TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    tags            TEXT NOT NULL DEFAULT '[]',
    stars           INTEGER NOT NULL DEFAULT 0,
    forks           INTEGER NOT NULL DEFAULT 0,
    watchers        INTEGER NOT NULL DEFAULT 0,
    community_refs  INTEGER NOT NULL DEFAULT 0,
    last_updated_at TIMESTAMPTZ,
    indexed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    score           DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    score_breakdown TEXT NOT NULL DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_skills_score   ON skills(score DESC);
CREATE INDEX IF NOT EXISTS idx_skills_repo    ON skills(repo_owner, repo_name);
CREATE INDEX IF NOT EXISTS idx_skills_active  ON skills(is_active);
CREATE INDEX IF NOT EXISTS idx_skills_indexed ON skills(indexed_at DESC);
ALTER TABLE skills ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_skills_fts ON skills USING GIN(search_vector);
CREATE OR REPLACE FUNCTION skills_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(NEW.tags, '')), 'C');
    RETURN NEW;
END
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS skills_search_vector_trigger ON skills;
CREATE TRIGGER skills_search_vector_trigger
    BEFORE INSERT OR UPDATE ON skills
    FOR EACH ROW EXECUTE FUNCTION skills_search_vector_update();
`},
	{2, `
CREATE TABLE IF NOT EXISTS api_keys (
    id           TEXT PRIMARY KEY,
    key_hash     TEXT UNIQUE NOT NULL,
    key_prefix   TEXT NOT NULL,
    name         TEXT NOT NULL,
    user_email   TEXT NOT NULL DEFAULT '',
    rate_limit   INTEGER NOT NULL DEFAULT 1000,
    calls_today  INTEGER NOT NULL DEFAULT 0,
    total_calls  INTEGER NOT NULL DEFAULT 0,
    is_admin     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_api_keys_hash   ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active);
`},
	{3, `
CREATE TABLE IF NOT EXISTS crawl_jobs (
    id              TEXT PRIMARY KEY,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    status          TEXT NOT NULL DEFAULT 'pending',
    skills_found    INTEGER NOT NULL DEFAULT 0,
    skills_updated  INTEGER NOT NULL DEFAULT 0,
    skills_new      INTEGER NOT NULL DEFAULT 0,
    github_queries  INTEGER NOT NULL DEFAULT 0,
    error           TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_crawl_jobs_status     ON crawl_jobs(status);
CREATE INDEX IF NOT EXISTS idx_crawl_jobs_started_at ON crawl_jobs(started_at DESC);
`},
	{4, `
CREATE TABLE IF NOT EXISTS skill_submissions (
    id           TEXT PRIMARY KEY,
    github_url   TEXT NOT NULL,
    submitted_by TEXT NOT NULL DEFAULT '',
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status       TEXT NOT NULL DEFAULT 'pending',
    notes        TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_submissions_status ON skill_submissions(status);
CREATE INDEX IF NOT EXISTS idx_submissions_url    ON skill_submissions(github_url);
`},
}

// New opens a PostgreSQL database and runs all pending migrations.
func New(dsn string) (*DB, error) {
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	database := &DB{sqlDB}

	if err := database.runMigrations(); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	log.Printf("[db] Connected to PostgreSQL")
	return database, nil
}

// runMigrations applies any pending migrations tracked by schema_migrations.
func (d *DB) runMigrations() error {
	// Bootstrap the tracking table itself
	if _, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations {
		var count int
		if err := d.QueryRow(
			"SELECT COUNT(*) FROM schema_migrations WHERE version = $1", m.version,
		).Scan(&count); err != nil {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if count > 0 {
			continue // already applied
		}

		if _, err := d.Exec(m.sql); err != nil {
			return fmt.Errorf("apply migration %d: %w", m.version, err)
		}

		if _, err := d.Exec(
			"INSERT INTO schema_migrations (version) VALUES ($1)", m.version,
		); err != nil {
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}

		log.Printf("[db] Applied migration %d", m.version)
	}

	version := 0
	_ = d.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	log.Printf("[db] Schema at version %d", version)
	return nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.DB.Close()
}
