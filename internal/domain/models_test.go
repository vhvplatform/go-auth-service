package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUser_Structure(t *testing.T) {
	user := &User{
		ID:           primitive.NewObjectID(),
		Email:        "test@example.com",
		PasswordHash: "$2a$12$hashedpassword",
		TenantID:     "tenant_001",
		Roles:        []string{"user", "admin"},
		IsActive:     true,
		IsVerified:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "tenant_001", user.TenantID)
	assert.Contains(t, user.Roles, "user")
	assert.Contains(t, user.Roles, "admin")
	assert.True(t, user.IsActive)
	assert.False(t, user.IsVerified)
}

func TestRefreshToken_Structure(t *testing.T) {
	now := time.Now()
	token := &RefreshToken{
		ID:        primitive.NewObjectID(),
		UserID:    "507f1f77bcf86cd799439011",
		Token:     "refresh-token-string",
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
	}

	assert.NotEmpty(t, token.ID)
	assert.Equal(t, "507f1f77bcf86cd799439011", token.UserID)
	assert.Equal(t, "refresh-token-string", token.Token)
	assert.True(t, token.ExpiresAt.After(now))
}

func TestRole_Structure(t *testing.T) {
	role := &Role{
		ID:          primitive.NewObjectID(),
		Name:        "admin",
		Description: "Administrator role",
		Permissions: []string{"read:all", "write:all", "delete:all"},
		TenantID:    "tenant_001",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.NotEmpty(t, role.ID)
	assert.Equal(t, "admin", role.Name)
	assert.Contains(t, role.Permissions, "read:all")
	assert.Equal(t, "tenant_001", role.TenantID)
}

func TestSession_Structure(t *testing.T) {
	now := time.Now()
	session := &Session{
		UserID:    "507f1f77bcf86cd799439011",
		TenantID:  "tenant_001",
		Email:     "user@example.com",
		Roles:     []string{"user"},
		CreatedAt: now,
		ExpiresAt: now.Add(1 * time.Hour),
	}

	assert.Equal(t, "507f1f77bcf86cd799439011", session.UserID)
	assert.Equal(t, "tenant_001", session.TenantID)
	assert.Equal(t, "user@example.com", session.Email)
	assert.True(t, session.ExpiresAt.After(now))
}

func TestOAuthProvider_Constants(t *testing.T) {
	assert.Equal(t, OAuthProvider("google"), OAuthProviderGoogle)
	assert.Equal(t, OAuthProvider("github"), OAuthProviderGitHub)
}

func TestOAuthAccount_Structure(t *testing.T) {
	account := &OAuthAccount{
		ID:         primitive.NewObjectID(),
		UserID:     "507f1f77bcf86cd799439011",
		Provider:   OAuthProviderGoogle,
		ProviderID: "google_12345",
		Email:      "user@gmail.com",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	assert.NotEmpty(t, account.ID)
	assert.Equal(t, "507f1f77bcf86cd799439011", account.UserID)
	assert.Equal(t, OAuthProviderGoogle, account.Provider)
	assert.Equal(t, "google_12345", account.ProviderID)
}
