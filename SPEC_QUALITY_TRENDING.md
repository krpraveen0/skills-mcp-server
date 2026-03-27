# Feature Spec: Quality Filtering + Trending Repositories

## 1 · Quality Filtering

### Problem Statement
The Explorer currently shows all 600+ indexed skills with zero quality gate. Most low-star, unmaintained repos add noise. Developers browsing for production-ready skills expect the same quality bar as npm top packages or Smithery — only widely-adopted, actively maintained work.

### Goals
- Default view shows only skills from repos with ≥ 100 GitHub stars
- Quality score cross-verifies adoption via stars + forks (composite guard)
- Users can toggle the quality gate off to browse the full catalogue
- Search respects the active quality filter

### Non-Goals
- No manual curation or editorial review (v1)
- No per-skill star history graph (v2)

### Requirements (P0)
- `GET /api/v1/skills?min_stars=100` param (default: 0; frontend passes 100)
- Explorer defaults `qualityFilter = true`; toggle chip disables it
- Skills with `stars >= 100 AND forks >= 5` pass the composite gate
- Badge "⭐ 100+" shown on skill cards in quality mode

### Success Metrics
- Explorer p50 skill quality score improves (baseline vs. post)
- Bounce rate from Explorer drops (fewer "what is this?" results)

---

## 2 · Trending Repositories Page

### Problem Statement
Users have no way to discover which GitHub repos are hot *right now*. Competitors (Smithery, LobeHub, mcp.so) all surface trending repos on their home page. Our Explorer only searches static indexed data.

### Goals
- New `/trending` page shows top-N repos ranked by stars, filterable by period
- Data refreshed daily via the existing crawl pipeline
- Each repo card links to a detail page showing all skills from that repo
- Period filters: Today (indexed in last 24h), This Week, This Month, All Time

### Non-Goals
- Real-time GitHub star growth tracking (we approximate with crawl timestamps)
- Repos with zero SKILL.md files (only repos in our DB are shown)
- Language breakdown charts (v2)

### User Stories
- "As a developer, I want to see which repos are trending today so I can discover new skills quickly"
- "As a developer, I want to filter by time period so I can see what was hot this week vs all time"
- "As a developer, I want to click a repo and see all its skills + stats"

### Requirements (P0)
- `GET /api/v1/repos/trending?period=week&min_stars=100&limit=10`
  - Groups skills by `(repo_owner, repo_name)`
  - Period filter: `today` → indexed_at ≥ 24h ago; `week` → last 7d; `month` → last 30d; `all` → no filter
  - Returns: owner, name, stars, forks, watchers, skill_count, top_score, description, tags
- `GET /api/v1/repos/:owner/:repo` — all skills for a repo + repo metadata
- Cached 6 hours; cache invalidated on admin cache flush
- `/trending` page with period toggle (Today / Week / Month / All) and min_stars filter
- Repo cards show: name, star count, fork count, skill count, top quality score
- `/repos/:owner/:repo` detail page lists all skills with click-through to SkillDetailPage

### Success Metrics
- Trending page gets > 30% of Explorer page views within 2 weeks of launch
- Time-on-site increases (users explore deeper via repo → skill drill-down)

### Open Questions
- Q (engineering): Should trending cache be busted daily at midnight or on crawl completion? → Use crawl-completion invalidation (already have cache flush endpoint)
