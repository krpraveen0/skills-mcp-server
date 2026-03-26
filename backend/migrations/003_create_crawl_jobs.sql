-- UP
CREATE TABLE IF NOT EXISTS crawl_jobs (
    id              TEXT PRIMARY KEY,
    started_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at    DATETIME,
    status          TEXT NOT NULL DEFAULT 'pending',
    skills_found    INTEGER NOT NULL DEFAULT 0,
    skills_updated  INTEGER NOT NULL DEFAULT 0,
    skills_new      INTEGER NOT NULL DEFAULT 0,
    github_queries  INTEGER NOT NULL DEFAULT 0,
    error           TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_crawl_jobs_status     ON crawl_jobs(status);
CREATE INDEX IF NOT EXISTS idx_crawl_jobs_started_at ON crawl_jobs(started_at DESC);
