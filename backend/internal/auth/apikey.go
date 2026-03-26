package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/krpraveen0/skills-mcp-server/internal/cache"
	"github.com/krpraveen0/skills-mcp-server/internal/db"
	"github.com/krpraveen0/skills-mcp-server/pkg/models"
)

const (
	rawKeyPrefix = "sk_live_"
	keyBytes     = 32
	cacheKeyTTL  = 5 * time.Minute
)

// Service handles API key creation and validation.
type Service struct {
	db    *db.DB
	cache *cache.Redis
}

// NewService creates a new auth service.
func NewService(database *db.DB, redisCache *cache.Redis) *Service {
	return &Service{db: database, cache: redisCache}
}

// GenerateKey creates a new cryptographically random API key.
func (s *Service) GenerateKey(ctx context.Context, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error) {
	rawBytes := make([]byte, keyBytes)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, fmt.Errorf("generate random key: %w", err)
	}
	rawKey := rawKeyPrefix + hex.EncodeToString(rawBytes)
	hash := hashKey(rawKey)
	prefix := rawKey[:len(rawKeyPrefix)+8]

	rateLimit := req.RateLimit
	if rateLimit <= 0 {
		rateLimit = 1000
	}

	key := &models.APIKey{
		ID:        uuid.New().String(),
		KeyHash:   hash,
		KeyPrefix: prefix,
		Name:      req.Name,
		UserEmail: req.Email,
		RateLimit: rateLimit,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := s.db.CreateAPIKey(ctx, key); err != nil {
		return nil, fmt.Errorf("store api key: %w", err)
	}

	return &models.CreateAPIKeyResponse{
		APIKey: *key,
		RawKey: rawKey,
	}, nil
}

// ValidateKey validates an API key and increments usage.
func (s *Service) ValidateKey(ctx context.Context, rawKey string) (*models.APIKey, error) {
	if rawKey == "" {
		return nil, errors.New("missing api key")
	}

	hash := hashKey(rawKey)
	cacheKey := "apikey:" + hash

	var cached models.APIKey
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		if !cached.IsActive {
			return nil, errors.New("api key revoked")
		}
		go func() {
			if err := s.db.IncrementAPIKeyUsage(context.Background(), cached.ID); err != nil {
				log.Printf("[auth] increment usage error: %v", err)
			}
		}()
		return &cached, nil
	}

	key, err := s.db.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid api key")
		}
		return nil, fmt.Errorf("lookup api key: %w", err)
	}

	if !key.IsActive {
		return nil, errors.New("api key revoked")
	}

	if key.CallsToday >= key.RateLimit {
		return nil, fmt.Errorf("rate limit exceeded (%d/%d calls today)", key.CallsToday, key.RateLimit)
	}

	s.cache.Set(ctx, cacheKey, key, cacheKeyTTL)

	if err := s.db.IncrementAPIKeyUsage(ctx, key.ID); err != nil {
		log.Printf("[auth] failed to increment usage: %v", err)
	}

	return key, nil
}

func hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
