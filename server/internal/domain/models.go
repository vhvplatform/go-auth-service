package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents the authentication data for a user
type User struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Email        string              `bson:"email,omitempty" json:"email,omitempty"`
	Username     string              `bson:"username,omitempty" json:"username,omitempty"`
	Phone        string              `bson:"phone,omitempty" json:"phone,omitempty"`
	DocNumber    string              `bson:"docNumber,omitempty" json:"doc_number,omitempty"`
	PasswordHash string              `bson:"passwordHash" json:"-"`
	Tenants      []string            `bson:"tenants" json:"tenants"`
	Roles        []string            `bson:"roles" json:"roles"`              // Global roles? Usually roles are per tenant.
	TenantRoles  map[string][]string `bson:"tenantRoles" json:"tenant_roles"` // tenantId -> roles
	IsActive     bool                `bson:"isActive" json:"is_active"`
	IsVerified   bool                `bson:"isVerified" json:"is_verified"`
	LastLoginAt  *time.Time          `bson:"lastLoginAt,omitempty" json:"last_login_at,omitempty"`
	CreatedAt    time.Time           `bson:"createdAt" json:"created_at"`
	UpdatedAt    time.Time           `bson:"updatedAt" json:"updated_at"`
}

// Tenant represents a tenant's configuration
type Tenant struct {
	ID           string    `bson:"_id" json:"id"`
	Name         string    `bson:"name" json:"name"`
	LoginMethods []string  `bson:"loginMethods" json:"login_methods"` // e.g. ["email", "username", "phone"]
	IsActive     bool      `bson:"isActive" json:"is_active"`
	CreatedAt    time.Time `bson:"createdAt" json:"created_at"`
	UpdatedAt    time.Time `bson:"updatedAt" json:"updated_at"`
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"userId" json:"user_id"`
	Token     string             `bson:"token" json:"token"`
	TenantID  string             `bson:"tenantId" json:"tenant_id"`
	ExpiresAt time.Time          `bson:"expiresAt" json:"expires_at"`
	CreatedAt time.Time          `bson:"createdAt" json:"created_at"`
	RevokedAt *time.Time         `bson:"revokedAt,omitempty" json:"revoked_at,omitempty"`
}

// Role represents a role in the system
type Role struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Permissions []string           `bson:"permissions" json:"permissions"`
	TenantID    string             `bson:"tenantId,omitempty" json:"tenant_id,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updated_at"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Resource    string             `bson:"resource" json:"resource"`
	Action      string             `bson:"action" json:"action"`
	CreatedAt   time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updated_at"`
}

// Session represents a user session stored in Redis
type Session struct {
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OAuthProvider represents OAuth provider types
type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderGitHub OAuthProvider = "github"
)

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     string             `bson:"userId" json:"user_id"`
	Provider   OAuthProvider      `bson:"provider" json:"provider"`
	ProviderID string             `bson:"providerId" json:"provider_id"`
	Email      string             `bson:"email" json:"email"`
	CreatedAt  time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updated_at"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserInfo `json:"user"`
}

// UserInfo represents brief user information in login response
type UserInfo struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	TenantID string   `json:"tenant_id"`
	Roles    []string `json:"roles"`
}

// ValidateTokenResponse represents the result of token validation
type ValidateTokenResponse struct {
	Valid        bool              `json:"valid"`
	UserID       string            `json:"user_id"`
	TenantID     string            `json:"tenant_id"`
	Email        string            `json:"email"`
	Roles        []string          `json:"roles"`
	Permissions  []string          `json:"permissions"`
	ErrorMessage string            `json:"error_message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}
