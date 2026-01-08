package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vhvplatform/go-shared/jwt"
	"github.com/vhvplatform/go-shared/logger"
	"go.uber.org/zap"
)

// AuthClient interface for calling Auth Service
type AuthClient interface {
	ValidateToken(ctx context.Context, token, tenantID string) (*ValidateTokenResponse, error)
}

// ValidateTokenResponse matches the info needed from Auth Service
type ValidateTokenResponse struct {
	Valid       bool
	UserID      string
	TenantID    string
	Email       string
	Roles       []string
	Permissions []string
}

// AuthMiddleware handles authentication and tenant verification at the gateway
func AuthMiddleware(authClient AuthClient, cache *Cache, jwtManager *jwt.Manager, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c.Request)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			c.Abort()
			return
		}

		tenantID := c.GetHeader("X-Tenant-ID")

		// Check local cache
		cacheKey := fmt.Sprintf("token:%s:%s", token, tenantID)
		if val, ok := cache.Get(cacheKey); ok {
			claims := val.(*ValidateTokenResponse)
			injectHeaders(c, claims, jwtManager, log)
			c.Next()
			return
		}

		// Call Auth Service
		resp, err := authClient.ValidateToken(c.Request.Context(), token, tenantID)
		if err != nil || !resp.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Cache the result (e.g. for 5 minutes)
		cache.Set(cacheKey, resp, 5*time.Minute)

		injectHeaders(c, resp, jwtManager, log)
		c.Next()
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func injectHeaders(c *gin.Context, resp *ValidateTokenResponse, jwtManager *jwt.Manager, log *logger.Logger) {
	// Generate internal-token (JWT)
	// Note: In a real scenario, use a specific secret for internal communication
	internalToken, err := jwtManager.GenerateToken(resp.UserID, resp.TenantID, resp.Email, resp.Roles, resp.Permissions)
	if err != nil {
		log.Error("Failed to generate internal token", zap.Error(err))
		return
	}

	c.Set("tenant_id", resp.TenantID)
	c.Set("internal_token", internalToken)

	// These will be used by the Proxy to set outgoing headers
}

// Since I used fmt.Sprintf, I need to add fmt to imports
