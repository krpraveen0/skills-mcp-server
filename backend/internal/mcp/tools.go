package mcp

// toolDefinitions returns all tool schemas exposed by this MCP server.
func toolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "search_skills",
			Description: "Search indexed SKILL.md files from GitHub by keyword, category, or description. Returns ranked results ordered by composite score (stars + adoption + recency).",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]SchemaField{
					"query": {
						Type:        "string",
						Description: "Search terms — keywords describing the skill you need (e.g. 'docker deployment', 'code review', 'database migration')",
					},
					"tags": {
						Type:        "array",
						Description: "Filter results by specific tags",
						Items:       &SchemaField{Type: "string"},
					},
					"limit": {
						Type:        "integer",
						Description: "Maximum number of results to return (1-50)",
						Default:     10,
						Maximum:     50,
					},
					"offset": {
						Type:        "integer",
						Description: "Pagination offset",
						Default:     0,
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "get_skill_detail",
			Description: "Retrieve the full SKILL.md content and metadata for a specific skill by its ID. Use this after search_skills to get the complete skill instructions.",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]SchemaField{
					"id": {
						Type:        "string",
						Description: "Skill UUID from search_skills or list_trending_skills results",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "list_trending_skills",
			Description: "Return the top-ranked SKILL.md files sorted by composite score combining GitHub stars, community adoption, and recency. Great for discovering high-quality skills without a specific query.",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]SchemaField{
					"limit": {
						Type:        "integer",
						Description: "Number of trending skills to return (1-100)",
						Default:     20,
						Maximum:     100,
					},
					"category": {
						Type:        "string",
						Description: "Optional category filter (e.g. 'devops', 'testing', 'documentation')",
					},
				},
			},
		},
		{
			Name:        "submit_skill",
			Description: "Submit a GitHub URL pointing to a SKILL.md file or repository for indexing. Submitted URLs are queued for crawling and ranking.",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]SchemaField{
					"github_url": {
						Type:        "string",
						Description: "Full GitHub URL to the SKILL.md file or repository (e.g. https://github.com/owner/repo/blob/main/SKILL.md)",
					},
					"notes": {
						Type:        "string",
						Description: "Optional notes about this skill — why it's useful, use cases, etc.",
					},
				},
				Required: []string{"github_url"},
			},
		},
	}
}
