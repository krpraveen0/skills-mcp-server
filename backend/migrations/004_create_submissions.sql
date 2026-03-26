-- UP
CREATE TABLE IF NOT EXISTS skill_submissions (
    id           TEXT PRIMARY KEY,
    github_url   TEXT NOT NULL,
    submitted_by TEXT NOT NULL DEFAULT '',
    submitted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status       TEXT NOT NULL DEFAULT 'pending',
    notes        TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_submissions_status ON skill_submissions(status);
CREATE INDEX IF NOT EXISTS idx_submissions_url    ON skill_submissions(github_url);
