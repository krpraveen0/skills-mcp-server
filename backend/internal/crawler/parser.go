package crawler

import (
	"regexp"
	"strings"
)

var (
	h1Pattern   = regexp.MustCompile(`(?m)^#\s+(.+)$`)
	tagPattern  = regexp.MustCompile(`(?i)(?:tags?|categories?):\s*(.+)`)
	codePattern = regexp.MustCompile("```[\\s\\S]*?```")
)

// ParsedSkill holds extracted metadata from a SKILL.md file.
type ParsedSkill struct {
	Title       string
	Description string
	Tags        []string
}

// ParseContent extracts structured metadata from raw SKILL.md content.
func ParseContent(content string) ParsedSkill {
	result := ParsedSkill{
		Tags: []string{},
	}

	lines := strings.Split(content, "\n")
	titleFound := false
	descLines := []string{}

	for i, line := range lines {
		line = strings.TrimRight(line, "\r")

		// Extract title from first H1
		if !titleFound && strings.HasPrefix(line, "# ") {
			result.Title = strings.TrimSpace(line[2:])
			titleFound = true
			continue
		}

		// Extract description from lines after title until next heading or blank line block
		if titleFound && len(descLines) < 5 {
			if strings.HasPrefix(line, "#") {
				break
			}
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "```") {
				descLines = append(descLines, trimmed)
			}
		}

		// Extract tags from "Tags:" or "Categories:" line
		if m := tagPattern.FindStringSubmatch(line); m != nil {
			result.Tags = parseTags(m[1])
		}

		_ = i
	}

	result.Description = strings.Join(descLines, " ")
	if len(result.Description) > 500 {
		result.Description = result.Description[:500] + "..."
	}

	// If no explicit title found, use a generic one
	if result.Title == "" {
		result.Title = "Untitled Skill"
	}

	// Auto-generate tags from title/description if none found
	if len(result.Tags) == 0 {
		result.Tags = autoTags(result.Title + " " + result.Description)
	}

	return result
}

// parseTags splits a comma/space-separated tag string.
func parseTags(raw string) []string {
	// Handle comma or space separated
	raw = strings.ReplaceAll(raw, ",", " ")
	parts := strings.Fields(raw)
	var tags []string
	for _, p := range parts {
		p = strings.Trim(p, `"'[]()`)
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" && len(p) < 50 {
			tags = append(tags, p)
		}
	}
	return tags
}

// autoTags extracts likely tags from skill text using keyword matching.
func autoTags(text string) []string {
	text = strings.ToLower(text)
	keywords := map[string]string{
		"docker":      "devops",
		"kubernetes":  "devops",
		"k8s":         "devops",
		"ci/cd":       "devops",
		"github":      "devops",
		"test":        "testing",
		"lint":        "quality",
		"review":      "code-review",
		"document":    "documentation",
		"database":    "database",
		"sql":         "database",
		"migration":   "database",
		"deploy":      "deployment",
		"build":       "build",
		"react":       "frontend",
		"vue":         "frontend",
		"api":         "api",
		"rest":        "api",
		"graphql":     "api",
		"typescript":  "typescript",
		"python":      "python",
		"golang":      "golang",
		"go ":         "golang",
		"rust":        "rust",
		"security":    "security",
		"auth":        "security",
		"performance": "performance",
		"monitor":     "monitoring",
		"observ":      "monitoring",
	}

	seen := map[string]bool{}
	var tags []string
	for keyword, tag := range keywords {
		if strings.Contains(text, keyword) && !seen[tag] {
			tags = append(tags, tag)
			seen[tag] = true
		}
	}
	return tags
}
