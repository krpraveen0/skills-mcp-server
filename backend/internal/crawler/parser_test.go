package crawler

import (
	"testing"
)

func TestParseContent_WithTitle(t *testing.T) {
	content := `# Docker Deployment Skill

This skill automates Docker container deployment to any VPS.

Tags: docker, devops, deployment
`
	result := ParseContent(content)

	if result.Title != "Docker Deployment Skill" {
		t.Errorf("expected title 'Docker Deployment Skill', got '%s'", result.Title)
	}
	if result.Description == "" {
		t.Error("expected non-empty description")
	}
	found := false
	for _, tag := range result.Tags {
		if tag == "docker" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'docker' in tags, got %v", result.Tags)
	}
}

func TestParseContent_NoTitle(t *testing.T) {
	content := "Some content without a proper title"
	result := ParseContent(content)
	if result.Title != "Untitled Skill" {
		t.Errorf("expected 'Untitled Skill', got '%s'", result.Title)
	}
}

func TestParseContent_AutoTags(t *testing.T) {
	content := "# Test Skill\nThis skill helps with kubernetes deployments and testing."
	result := ParseContent(content)
	if len(result.Tags) == 0 {
		t.Error("expected auto-generated tags for kubernetes/testing content")
	}
}

func TestFtsQuery(t *testing.T) {
	q := ftsQuery("docker deploy")
	if q == "" {
		t.Error("expected non-empty FTS query")
	}
}
