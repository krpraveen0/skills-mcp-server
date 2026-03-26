# MASTER PLAN — skills-mcp-server
> **Living document.** Update this file before any architectural change, new feature, or sprint kickoff.
> Last updated: 2026-03-26 | Status: 🟡 Planning Complete → Implementation Starting

---

## 1. PROJECT EVALUATION

### 1.1 Is This Useful?

**Verdict: ✅ YES — High value, clear gap in the market.**

The SKILL.md ecosystem has exploded in 2025–2026 (Anthropic, Microsoft, VoltAgent, ComposioHQ all maintain large skill registries). However:

| Problem Today | Impact |
|---|---|
| No centralized discovery layer | Developers search GitHub manually, wasting hours |
| No quality signal | All skills.md files look equal — no way to know which ones are battle-tested |
| No MCP-native interface | You can't ask your AI agent "find me the best Docker deployment skill" |
| Fragmented registries | Anthropic repo, Microsoft repo, VoltAgent repo — no unified index |
| Existing solutions are incomplete | `skills-mcp/skills-mcp`, `K-Dense-AI/claude-skills-mcp` exist but lack scoring, production infra, and a UI |

**Our differentiator:** A production-grade, scored, cached, and UI-accessible skill discovery layer exposed as a standards-compliant MCP server.

---

### 1.2 What Else Could We Solve? (Agentic AI Industry Pain Points, 2026)

Below are the top unsolved problems in agentic AI workflows where the industry is still struggling. These are ranked by severity and opportunity:

| # | Problem | Severity | Solution Opportunity |
|---|---|---|---|
| 1 | **MCP Server Discovery & Health Monitoring** | 🔴 Critical | A registry of MCP servers with uptime checks, capability listing, and versioning |
| 2 | **Agent Workflow Observability** | 🔴 Critical | Distributed tracing for multi-agent pipelines (what each agent did, why, cost) |
| 3 | **Skill Compatibility Matrix** | 🟠 High | Does this skill.md work with Claude / Codex / Gemini CLI? Automated compatibility testing |
| 4 | **Agent Memory Management Layer** | 🟠 High | Persistent, scoped, and queryable memory across agent sessions |
| 5 | **Prompt + Skill Version Control** | 🟠 High | Git-style diffing and rollback for prompt chains and skill definitions |
| 6 | **Multi-Agent Coordination Bus** | 🟡 Medium | Pub/sub messaging between autonomous agents with conflict resolution |
| 7 | **Agent Cost Attribution** | 🟡 Medium | Which agent/skill/task is consuming what LLM token budget |

**Our current focus (Problem 0):** Skill discovery and quality ranking — this is a pre-requisite for many of the above.

---

## 2. ARCHITECTURE OVERVIEW

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                                │
│  ┌──────────────────┐          ┌───────────────────────────────┐   │
│  │  AI Agents        │          │  React 19 + Vite + MUI        │   │
│  │  (Claude, Codex,  │          │  Dashboard                    │   │
│  │   Gemini CLI)     │          │  ┌─────────────┐ ┌─────────┐ │   │
│  └────────┬─────────┘          │  │ Skill       │ │ Admin   │ │   │
│           │ MCP JSON-RPC 2.0   │  │ Explorer    │ │ Panel   │ │   │
│           │ over HTTP/SSE      │  └─────────────┘ └─────────┘ │   │
│           │                    └───────────────┬───────────────┘   │
└───────────┼────────────────────────────────────┼───────────────────┘
            │                                    │ REST API
┌───────────▼────────────────────────────────────▼───────────────────┐
│                      GO BACKEND                                     │
│                                                                     │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │ MCP Handler │  │ REST API     │  │ Background Worker        │  │
│  │             │  │ (Gin)        │  │ (daily crawl cron)       │  │
│  │ search_skills│  │ /api/v1/*   │  │ GitHub Search API        │  │
│  │ get_skill   │  │             │  │ Scorer Engine            │  │
│  │ list_trending│  │             │  │ Index updater            │  │
│  │ submit_skill│  │             │  │                          │  │
│  └──────┬──────┘  └──────┬──────┘  └────────────┬─────────────┘  │
│         │                │                        │                │
│  ┌──────▼────────────────▼────────────────────────▼─────────────┐ │
│  │                    SERVICE LAYER                              │ │
│  │  SkillService │ CrawlerService │ ScorerService │ AuthService  │ │
│  └──────┬─────────────────────────────────────────────┬─────────┘ │
│         │                                             │            │
│  ┌──────▼────────────┐              ┌─────────────────▼──────────┐ │
│  │  Redis Cache      │              │  SQLite (via modernc)      │ │
│  │  - search results │              │  - skills table            │ │
│  │  - trending list  │              │  - api_keys table          │ │
│  │  - skill content  │              │  - crawl_jobs table        │ │
│  │  TTL: 1hr / 24hr  │              │  - submissions table       │ │
│  └───────────────────┘              └────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
            │
┌───────────▼─────────────────────────┐
│         GITHUB API                  │
│  Search Code API (skills.md files)  │
│  Repos API (stars, forks, metadata) │
│  Rate limit: 5000/hr authenticated  │
└─────────────────────────────────────┘
```

---

## 3. TECH STACK DECISIONS

| Layer | Technology | Rationale |
|---|---|---|
| Backend Language | **Go 1.22+** | Excellent concurrency for crawlers, low memory, single binary deployment |
| HTTP Router | **Gin** | Fast, mature, middleware ecosystem |
| Database | **SQLite** (via `modernc.org/sqlite`) | Zero-ops, sufficient for this workload, file-based backups trivial |
| DB Migrations | **golang-migrate** | SQL-based, version-controlled schema |
| Cache | **Redis 7** | MCP response caching, rate-limit counters, trending leaderboard |
| Redis Client | **go-redis/v9** | Most actively maintained Go Redis client |
| GitHub Client | Custom HTTP + `google/go-github` | Rate-limit aware, retry logic |
| Frontend | **React 19 + Vite 6** | Latest React concurrent features, fast dev server |
| UI Library | **Material UI (MUI) v6** | Comprehensive components, theming, accessibility |
| State | **Zustand** | Lightweight, React 19 compatible |
| Data Fetching | **TanStack Query v5** | Server state, caching, background refetch |
| Routing | **React Router v7** | File-based routing |
| HTTP Client | **Axios** | Request/response interceptors for API key injection |
| Containerization | **Docker + Docker Compose** | Single command startup |
| CI/CD | **GitHub Actions** | Test → Build → Push → Deploy pipeline |
| Reverse Proxy | **Nginx** (in docker-compose) | Static file serving + API proxying |

---

## 4. SCORING ALGORITHM

Every skill.md receives a composite score ∈ [0, 100]:

```
CompositeScore = (0.35 × StarScore) + (0.35 × AdoptionScore) + (0.30 × RecencyScore)

StarScore     = normalize(log(stars + 1), global_max_log_stars) × 100
AdoptionScore = normalize(log(forks + community_refs + 1), global_max) × 100
RecencyScore  = e^(-days_since_last_update / 180) × 100   ← exponential decay, half-life 6 months
```

Re-scored every time a crawl job completes (daily). Scores cached in Redis for 24h.

---

## 5. DATABASE SCHEMA

### 5.1 skills
```sql
CREATE TABLE skills (
    id            TEXT PRIMARY KEY,          -- UUID v4
    github_url    TEXT UNIQUE NOT NULL,      -- https://github.com/owner/repo/blob/main/path/SKILL.md
    repo_owner    TEXT NOT NULL,
    repo_name     TEXT NOT NULL,
    file_path     TEXT NOT NULL,             -- relative path within repo
    content       TEXT NOT NULL,             -- raw SKILL.md content
    title         TEXT,                      -- extracted from first H1
    description   TEXT,                      -- extracted from first paragraph
    tags          TEXT DEFAULT '[]',         -- JSON string array
    stars         INTEGER DEFAULT 0,
    forks         INTEGER DEFAULT 0,
    watchers      INTEGER DEFAULT 0,
    community_refs INTEGER DEFAULT 0,        -- # times referenced in other repos
    last_updated_at DATETIME,               -- last commit date on the file
    indexed_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    score         REAL DEFAULT 0.0,          -- composite score [0-100]
    score_breakdown TEXT DEFAULT '{}',       -- JSON: {star_score, adoption_score, recency_score}
    is_active     BOOLEAN DEFAULT TRUE
);
CREATE INDEX idx_skills_score ON skills(score DESC);
CREATE INDEX idx_skills_tags ON skills(tags);
CREATE INDEX idx_skills_repo ON skills(repo_owner, repo_name);
```

### 5.2 api_keys
```sql
CREATE TABLE api_keys (
    id            TEXT PRIMARY KEY,
    key_hash      TEXT UNIQUE NOT NULL,      -- SHA-256 of raw key, never store raw
    key_prefix    TEXT NOT NULL,             -- first 8 chars for display (e.g. "sk_live_")
    name          TEXT NOT NULL,
    user_email    TEXT,
    rate_limit    INTEGER DEFAULT 1000,      -- requests/day
    calls_today   INTEGER DEFAULT 0,
    total_calls   INTEGER DEFAULT 0,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at  DATETIME,
    is_active     BOOLEAN DEFAULT TRUE
);
```

### 5.3 crawl_jobs
```sql
CREATE TABLE crawl_jobs (
    id              TEXT PRIMARY KEY,
    started_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at    DATETIME,
    status          TEXT DEFAULT 'pending',  -- pending|running|completed|failed
    skills_found    INTEGER DEFAULT 0,
    skills_updated  INTEGER DEFAULT 0,
    skills_new      INTEGER DEFAULT 0,
    github_queries  INTEGER DEFAULT 0,
    error           TEXT
);
```

### 5.4 skill_submissions
```sql
CREATE TABLE skill_submissions (
    id           TEXT PRIMARY KEY,
    github_url   TEXT NOT NULL,
    submitted_by TEXT,                       -- API key prefix
    submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status       TEXT DEFAULT 'pending',     -- pending|indexed|rejected|duplicate
    notes        TEXT
);
```

---

## 6. MCP PROTOCOL SPECIFICATION

The server implements **MCP over HTTP with Server-Sent Events (SSE)** per the MCP 1.0 spec.

### Endpoint
```
POST /mcp
Authorization: Bearer sk_live_XXXX
Content-Type: application/json
```

### Tool Definitions

#### `search_skills`
```json
{
  "name": "search_skills",
  "description": "Search indexed SKILL.md files from GitHub by keyword, tag, or description. Returns ranked results.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": { "type": "string", "description": "Search terms" },
      "tags": { "type": "array", "items": {"type": "string"}, "description": "Filter by tags" },
      "limit": { "type": "integer", "default": 10, "maximum": 50 },
      "offset": { "type": "integer", "default": 0 }
    },
    "required": ["query"]
  }
}
```

#### `get_skill_detail`
```json
{
  "name": "get_skill_detail",
  "description": "Retrieve the full SKILL.md content and metadata for a specific skill by ID.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "id": { "type": "string", "description": "Skill UUID from search results" }
    },
    "required": ["id"]
  }
}
```

#### `list_trending_skills`
```json
{
  "name": "list_trending_skills",
  "description": "Return the top-ranked SKILL.md files sorted by composite score (stars + adoption + recency).",
  "inputSchema": {
    "type": "object",
    "properties": {
      "limit": { "type": "integer", "default": 20, "maximum": 100 },
      "category": { "type": "string", "description": "Optional category filter" }
    }
  }
}
```

#### `submit_skill`
```json
{
  "name": "submit_skill",
  "description": "Submit a GitHub URL containing a SKILL.md file for indexing and ranking.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "github_url": { "type": "string", "description": "Full GitHub URL to the SKILL.md file or repo" },
      "notes": { "type": "string", "description": "Optional notes about this skill" }
    },
    "required": ["github_url"]
  }
}
```

---

## 7. REST API ENDPOINTS

```
# Public (requires API Key)
GET  /api/v1/skills              ?q=&tags=&limit=&offset=    Search skills
GET  /api/v1/skills/:id                                      Get skill detail
GET  /api/v1/skills/trending     ?limit=&category=          Trending skills

# Submission
POST /api/v1/skills/submit       { github_url, notes }       Submit URL

# Admin (requires Admin API Key)
GET  /api/v1/admin/stats                                     Dashboard stats
GET  /api/v1/admin/keys                                      List API keys
POST /api/v1/admin/keys          { name, email, rate_limit } Create API key
PUT  /api/v1/admin/keys/:id      { is_active, rate_limit }   Update key
DELETE /api/v1/admin/keys/:id                                Revoke key
GET  /api/v1/admin/crawl/jobs    ?limit=                     Crawl history
POST /api/v1/admin/crawl/trigger                             Manual crawl

# Health
GET  /health                                                 Health check
GET  /metrics                                                Prometheus metrics
```

---

## 8. FOLDER STRUCTURE

```
skills-mcp-server/
├── backend/
│   ├── cmd/
│   │   ├── server/
│   │   │   └── main.go                 # HTTP server entrypoint
│   │   └── worker/
│   │       └── main.go                 # Background worker entrypoint
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler_skills.go       # Skills REST handlers
│   │   │   ├── handler_admin.go        # Admin REST handlers
│   │   │   ├── router.go               # Gin router setup
│   │   │   └── middleware.go           # Auth, CORS, rate limit
│   │   ├── mcp/
│   │   │   ├── server.go               # MCP protocol server
│   │   │   ├── tools.go                # Tool definitions & dispatch
│   │   │   └── types.go                # MCP JSON-RPC types
│   │   ├── crawler/
│   │   │   ├── crawler.go              # GitHub crawler orchestrator
│   │   │   ├── github_client.go        # GitHub API client w/ rate limiting
│   │   │   └── parser.go               # SKILL.md content parser
│   │   ├── scorer/
│   │   │   └── scorer.go               # Composite scoring engine
│   │   ├── cache/
│   │   │   └── redis.go                # Redis cache abstraction
│   │   ├── db/
│   │   │   ├── db.go                   # SQLite connection & queries
│   │   │   └── queries.go              # SQL query functions
│   │   ├── auth/
│   │   │   └── apikey.go               # API key generation & validation
│   │   └── config/
│   │       └── config.go               # Env-based config
│   ├── pkg/
│   │   └── models/
│   │       └── models.go               # Shared structs
│   ├── migrations/
│   │   ├── 001_create_skills.sql
│   │   ├── 002_create_api_keys.sql
│   │   ├── 003_create_crawl_jobs.sql
│   │   └── 004_create_submissions.sql
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   ├── App.tsx
│   │   │   └── router.tsx
│   │   ├── pages/
│   │   │   ├── explorer/
│   │   │   │   ├── ExplorerPage.tsx    # Public skill search UI
│   │   │   │   └── SkillDetailPage.tsx
│   │   │   ├── admin/
│   │   │   │   ├── AdminDashboard.tsx  # Stats overview
│   │   │   │   ├── ApiKeysPage.tsx     # API key management
│   │   │   │   └── CrawlJobsPage.tsx   # Crawl history
│   │   │   └── auth/
│   │   │       └── LoginPage.tsx
│   │   ├── components/
│   │   │   ├── common/
│   │   │   │   ├── Navbar.tsx
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   └── LoadingSpinner.tsx
│   │   │   ├── skills/
│   │   │   │   ├── SkillCard.tsx
│   │   │   │   ├── SkillSearchBar.tsx
│   │   │   │   └── ScoreBadge.tsx
│   │   │   └── admin/
│   │   │       ├── StatsCard.tsx
│   │   │       └── CrawlJobRow.tsx
│   │   ├── hooks/
│   │   │   ├── useSkills.ts
│   │   │   └── useAdmin.ts
│   │   ├── services/
│   │   │   ├── api.ts                  # Axios instance + interceptors
│   │   │   ├── skills.service.ts
│   │   │   └── admin.service.ts
│   │   ├── store/
│   │   │   └── useAppStore.ts          # Zustand global store
│   │   └── theme/
│   │       └── theme.ts                # MUI theme config
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
│
├── docker/
│   ├── Dockerfile.backend
│   ├── Dockerfile.frontend
│   └── nginx.conf
│
├── docker-compose.yml
├── docker-compose.prod.yml
│
├── .github/
│   └── workflows/
│       ├── ci.yml                      # PR: lint + test
│       └── deploy.yml                  # Main: build + push + deploy
│
├── docs/
│   └── MASTER_PLAN.md                  # ← this file
│
└── README.md
```

---

## 9. ENVIRONMENT CONFIGURATION

```env
# Server
PORT=8080
ENV=production                           # development | production

# Database
SQLITE_PATH=/data/skills.db

# Redis
REDIS_URL=redis://redis:6379
REDIS_PASSWORD=

# GitHub
GITHUB_TOKEN=                            # Optional — increases rate limit 5000/hr
GITHUB_CRAWL_QUERIES=filename:SKILL.md,filename:skills.md

# Auth
ADMIN_API_KEY=                           # Master admin key (hashed on startup)
API_KEY_SALT=                            # Random salt for HMAC

# Cache TTLs (seconds)
CACHE_TTL_SEARCH=3600                    # 1 hour
CACHE_TTL_TRENDING=86400                 # 24 hours
CACHE_TTL_SKILL=3600

# Crawler
CRAWL_SCHEDULE=0 2 * * *               # 2am daily cron
CRAWL_MAX_RESULTS=1000                   # Max skills per crawl run
```

---

## 10. GITHUB ACTIONS CI/CD PIPELINE

### CI (`.github/workflows/ci.yml`) — runs on every PR
```
Trigger: pull_request → main
Steps:
  1. go test ./...                        (backend unit tests)
  2. go vet + staticcheck                 (linting)
  3. npm run build                        (frontend build check)
  4. npm run test                         (frontend unit tests)
```

### Deploy (`.github/workflows/deploy.yml`) — runs on merge to main
```
Trigger: push → main
Steps:
  1. Build backend Docker image → push to GHCR
  2. Build frontend Docker image → push to GHCR
  3. SSH into VPS
  4. docker-compose pull
  5. docker-compose up -d --no-build
  6. Health check /health endpoint
  7. Slack/webhook notification
```

Required GitHub Secrets:
- `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`
- `GHCR_TOKEN` (GitHub Container Registry)
- `GITHUB_TOKEN` (auto-provided)

---

## 11. PHASED ROADMAP

### Phase 1 — MVP (Current Sprint)
- [x] Project structure
- [ ] Go backend: SQLite + Redis + Gin router
- [ ] GitHub crawler + daily cron worker
- [ ] Composite scoring engine
- [ ] 4 MCP tools (search, get, trending, submit)
- [ ] API key auth middleware
- [ ] React dashboard: Explorer + Admin
- [ ] Docker + GitHub Actions

### Phase 2 — Enhancement
- [ ] Full-text search with SQLite FTS5
- [ ] Skill content quality scoring (structural analysis)
- [ ] Webhook: re-index on GitHub push events
- [ ] Skills compatibility tags (Claude / Codex / Gemini)
- [ ] Email notifications for submission status

### Phase 3 — Scale
- [ ] Swap SQLite → PostgreSQL for multi-instance
- [ ] OpenTelemetry tracing for MCP calls
- [ ] Public API docs (Swagger/OpenAPI)
- [ ] Skill collections / bookmarks (user accounts)
- [ ] Premium tier with higher rate limits

---

## 12. CHANGE LOG

| Date | Change | Author |
|---|---|---|
| 2026-03-26 | Initial master plan created | Rohit / Claude |

---

## 13. DECISION LOG

| Decision | Rationale | Date |
|---|---|---|
| SQLite over PostgreSQL | Zero-ops for MVP, modernc driver = no CGO, trivial backups | 2026-03-26 |
| MCP over HTTP (not stdio) | HTTP allows web clients + agents to use same endpoint | 2026-03-26 |
| Gin over Chi/Echo | Mature middleware ecosystem, familiar to most Go devs | 2026-03-26 |
| Composite score (no AI) | Reproducible, explainable, no LLM cost on every crawl | 2026-03-26 |
| API key (not OAuth) | Simpler for agent-to-server auth; OAuth Phase 2 | 2026-03-26 |
| React 19 + Vite | Concurrent features needed for real-time crawl status | 2026-03-26 |
