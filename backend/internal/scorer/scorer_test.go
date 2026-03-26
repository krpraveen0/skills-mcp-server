package scorer

import (
	"math"
	"testing"
	"time"

	"github.com/krpraveen0/skills-mcp-server/pkg/models"
)

func TestScoreSkill_HighStars(t *testing.T) {
	now := time.Now()
	s := &models.Skill{
		Stars:         10000,
		Forks:         2000,
		CommunityRefs: 100,
		LastUpdatedAt: &now,
	}
	score := ScoreSkill(s, math.Log(10001), math.Log(2101))
	if score < 50 {
		t.Errorf("expected high-star skill score >= 50, got %.2f", score)
	}
}

func TestScoreSkill_NewSkill_LowScore(t *testing.T) {
	s := &models.Skill{
		Stars:         0,
		Forks:         0,
		CommunityRefs: 0,
		LastUpdatedAt: nil,
	}
	score := ScoreSkill(s, 10, 10)
	if score > 40 {
		t.Errorf("expected new/unknown skill score < 40, got %.2f", score)
	}
}

func TestRecencyScore_Recent(t *testing.T) {
	now := time.Now()
	score := recencyScore(&now)
	if score < 0.99 {
		t.Errorf("expected recency score near 1 for today, got %.4f", score)
	}
}

func TestRecencyScore_OldFile(t *testing.T) {
	old := time.Now().AddDate(-2, 0, 0) // 2 years ago
	score := recencyScore(&old)
	if score > 0.2 {
		t.Errorf("expected low recency score for 2yr old file, got %.4f", score)
	}
}

func TestRecencyScore_NilDate(t *testing.T) {
	score := recencyScore(nil)
	if score <= 0 || score > 1 {
		t.Errorf("expected recency score in (0,1] for nil date, got %.4f", score)
	}
}

func TestNormalizeLog(t *testing.T) {
	if v := normalizeLog(100, math.Log(101)); v <= 0 || v > 1 {
		t.Errorf("expected normalized value in (0,1], got %.4f", v)
	}
	if v := normalizeLog(0, 10); v != 0 {
		t.Errorf("expected 0 for val=0, got %.4f", v)
	}
}
