# 🧠 skills-mcp-server

> A production-ready MCP server for discovering and ranking the best SKILL.md files from GitHub.

[![CI](https://github.com/skills-mcp/skills-mcp-server/actions/workflows/ci.yml/badge.svg)](https://github.com/skills-mcp/skills-mcp-server/actions)

## What It Does

Indexes and ranks every `SKILL.md` file on GitHub using a composite score:
- **35%** — Repository stars
- **35%** — Community adoption (forks + references)
- **30%** — Recency (exponential decay, 6-month half-life)

Exposes 4 MCP tools that any AI agent (Claude, Codex, Gemini CLI) can call:

| Tool | Description |
|------|-------------|
| `search_skills` | Search by keyword or tag |
| `get_skill_detail` | Fetch full SKILL.md content |
| `list_trending_skills` | Top N by composite score |
| `submit_skill` | Submit a GitHub URL for indexing |

## Quick Start

```bash
# 1. Clone & configure
git clone https://github.com/your-org/skills-mcp-server
cd skills-mcp-server
cp .env.example .env
# Edit .env with your ADMIN_API_KEY, API_KEY_SALT, GITHUB_TOKEN

# 2. Start everything
docker-compose up -d

# 3. Access
# Dashboard:  http://localhost
# MCP server: http://localhost:8080/mcp
# API:        http://localhost:8080/api/v1
```

## Using the MCP Server

Add to your Claude Code config (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "skills": {
      "type": "http",
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Bearer sk_live_YOUR_KEY"
      }
    }
  }
}
```

Then in any Claude session:
```
Find me the best Docker deployment skill
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22 + Gin |
| Database | SQLite (FTS5) |
| Cache | Redis 7 |
| Frontend | React 19 + Vite 6 + MUI v6 |
| Deploy | Docker + GitHub Actions → VPS |

## Required GitHub Secrets (for deployment)

| Secret | Description |
|--------|-------------|
| `VPS_HOST` | Your VPS IP or hostname |
| `VPS_USER` | SSH username |
| `VPS_SSH_KEY` | SSH private key (no passphrase) |

## Architecture

See [docs/MASTER_PLAN.md](docs/MASTER_PLAN.md) for full architecture documentation.

## License

MIT
