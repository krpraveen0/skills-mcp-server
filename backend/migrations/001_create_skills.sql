-- UP
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
    last_updated_at DATETIME,
    indexed_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    score           REAL NOT NULL DEFAULT 0.0,
    score_breakdown TEXT NOT NULL DEFAULT '{}',
    is_active       INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_skills_score     ON skills(score DESC);
CREATE INDEX IF NOT EXISTS idx_skills_repo      ON skills(repo_owner, repo_name);
CREATE INDEX IF NOT EXISTS idx_skills_active    ON skills(is_active);
CREATE INDEX IF NOT EXISTS idx_skills_indexed   ON skills(indexed_at DESC);

-- Full-text search virtual table
CREATE VIRTUAL TABLE IF NOT EXISTS skills_fts USING fts5(
    id UNINDEXED,
    title,
    description,
    tags,
    content='skills',
    content_rowid='rowid'
);

-- Keep FTS in sync via triggers
CREATE TRIGGER IF NOT EXISTS skills_fts_insert AFTER INSERT ON skills BEGIN
    INSERT INTO skills_fts(rowid, id, title, description, tags)
    VALUES (new.rowid, new.id, new.title, new.description, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS skills_fts_update AFTER UPDATE ON skills BEGIN
    INSERT INTO skills_fts(skills_fts, rowid, id, title, description, tags)
    VALUES ('delete', old.rowid, old.id, old.title, old.description, old.tags);
    INSERT INTO skills_fts(rowid, id, title, description, tags)
    VALUES (new.rowid, new.id, new.title, new.description, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS skills_fts_delete AFTER DELETE ON skills BEGIN
    INSERT INTO skills_fts(skills_fts, rowid, id, title, description, tags)
    VALUES ('delete', old.rowid, old.id, old.title, old.description, old.tags);
END;
