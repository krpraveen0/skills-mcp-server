package scorer

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/krpraveen0/skills-mcp-server/internal/db"
	"github.com/krpraveen0/skills-mcp-server/pkg/models"
)

// Weights for the composite score
const (
	weightStars    = 0.35
	weightAdoption = 0.35
	weightRecency  = 0.30

	// Recency half-life in days — score halves every 180 days
	recencyHalfLife = 180.0
)

// Engine computes composite scores for skills.
type Engine struct {
	db *db.DB
}

// New creates a new scoring engine.
func New(database *db.DB) *Engine {
	return &Engine{db: database}
}

// ScoreSkill computes the composite score for a single skill.
// It needs global maxima for normalization, so prefer ScoreAll for batch runs.
func ScoreSkill(s *models.Skill, maxLogStars, maxLogAdoption float64) float64 {
	star := normalizeLog(float64(s.Stars), maxLogStars)
	adopt := normalizeLog(float64(s.Forks+s.CommunityRefs), maxLogAdoption)
	rec := recencyScore(s.LastUpdatedAt)

	composite := (weightStars*star + weightAdoption*adopt + weightRecency*rec) * 100

	s.ScoreBreakdown = models.ScoreBreakdown{
		StarScore:      roundTo2(star * 100),
		AdoptionScore:  roundTo2(adopt * 100),
		RecencyScore:   roundTo2(rec * 100),
		CompositeScore: roundTo2(composite),
	}
	return roundTo2(composite)
}

// ScoreAll rescores all active skills in the database.
// It first computes global maxima for normalization, then updates each skill.
func (e *Engine) ScoreAll(ctx context.Context, skills []models.Skill) []models.Skill {
	if len(skills) == 0 {
		return skills
	}

	// Compute global maxima for log-normalization
	var maxLogStars, maxLogAdoption float64
	for _, s := range skills {
		if v := math.Log(float64(s.Stars) + 1); v > maxLogStars {
			maxLogStars = v
		}
		if v := math.Log(float64(s.Forks+s.CommunityRefs) + 1); v > maxLogAdoption {
			maxLogAdoption = v
		}
	}
	if maxLogStars == 0 {
		maxLogStars = 1
	}
	if maxLogAdoption == 0 {
		maxLogAdoption = 1
	}

	for i := range skills {
		skills[i].Score = ScoreSkill(&skills[i], maxLogStars, maxLogAdoption)
	}

	return skills
}

// PersistScores writes the updated scores back to the database.
func (e *Engine) PersistScores(ctx context.Context, skills []models.Skill) error {
	log.Printf("[scorer] Persisting scores for %d skills", len(skills))
	for i := range skills {
		if err := e.db.UpsertSkill(ctx, &skills[i]); err != nil {
			return err
		}
	}
	return nil
}

// --- Math helpers ---

// normalizeLog applies log(x+1) normalization against a global max.
func normalizeLog(val, globalMax float64) float64 {
	if globalMax <= 0 {
		return 0
	}
	return math.Log(val+1) / globalMax
}

// recencyScore returns an exponential decay score ∈ [0, 1]
// based on the number of days since the file was last updated.
func recencyScore(lastUpdated *time.Time) float64 {
	if lastUpdated == nil {
		// Unknown update time — treat as 1 year old
		return math.Exp(-365.0 / recencyHalfLife * math.Log(2))
	}
	days := time.Since(*lastUpdated).Hours() / 24
	if days < 0 {
		days = 0
	}
	// Exponential decay: score = e^(-λt) where λ = ln(2)/halfLife
	lambda := math.Log(2) / recencyHalfLife
	return math.Exp(-lambda * days)
}

func roundTo2(v float64) float64 {
	return math.Round(v*100) / 100
}
