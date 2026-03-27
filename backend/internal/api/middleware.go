package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/krpraveen0/skills-mcp-server/internal/auth"
	"github.com/krpraveen0/skills-mcp-server/pkg/models"
)

const apiKeyContextKey = "api_key"

// AuthMiddleware validates Bearer API keys for all protected routes.
// If adminKey is non-empty and the request presents that exact key, the DB
// lookup is skipped — this allows the first admin to log in and create real
// API keys before any exist in the database (bootstrap flow).
func AuthMiddleware(authSvc *auth.Service, adminKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := extractBearerToken(c)
		if rawKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Code:    401,
				Message: "Missing Authorization header. Use: Authorization: Bearer <api_key>",
			})
			return
		}

		// Admin bootstrap: ADMIN_API_KEY env var bypasses the DB entirely.
		if adminKey != "" && rawKey == adminKey {
			c.Set(apiKeyContextKey, &models.APIKey{
				ID:        "admin-env",
				Name:      "Admin (env bootstrap)",
				IsActive:  true,
				IsAdmin:   true,
				RateLimit: 999999,
			})
			c.Next()
			return
		}

		key, err := authSvc.ValidateKey(c.Request.Context(), rawKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Code:    401,
				Message: err.Error(),
			})
			return
		}

		c.Set(apiKeyContextKey, key)
		c.Next()
	}
}

// OptionalAuthMiddleware extracts the API key if present but does NOT require
// it. Public routes use this so anonymous users can browse freely, while
// authenticated users still get usage tracking and higher rate limits.
func OptionalAuthMiddleware(authSvc *auth.Service, adminKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := extractBearerToken(c)
		if rawKey == "" {
			c.Next()
			return
		}

		// Admin bootstrap
		if adminKey != "" && rawKey == adminKey {
			c.Set(apiKeyContextKey, &models.APIKey{
				ID:        "admin-env",
				Name:      "Admin (env bootstrap)",
				IsActive:  true,
				IsAdmin:   true,
				RateLimit: 999999,
			})
			c.Next()
			return
		}

		if key, err := authSvc.ValidateKey(c.Request.Context(), rawKey); err == nil {
			c.Set(apiKeyContextKey, key)
		}
		c.Next()
	}
}

// AdminMiddleware checks that the authenticated user has admin privileges.
// Must be used after AuthMiddleware (which sets the apiKeyContextKey).
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get(apiKeyContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Code:    403,
				Message: "Admin access required",
			})
			return
		}
		key, ok := raw.(*models.APIKey)
		if !ok || !key.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Code:    403,
				Message: "Admin access required",
			})
			return
		}
		c.Next()
	}
}

// RequestIDMiddleware injects a unique request ID header.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// CORSMiddleware configures CORS headers.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return c.Query("api_key")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return parts[1]
	}
	return ""
}

func generateRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "req_" + hex.EncodeToString(b)
}
