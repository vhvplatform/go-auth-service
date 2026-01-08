package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserTenant represents the relationship between a user and a tenant
type UserTenant struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"userId" json:"user_id"`
	TenantID  string             `bson:"tenantId" json:"tenant_id"`
	Roles     []string           `bson:"roles" json:"roles"`
	IsActive  bool               `bson:"isActive" json:"is_active"`
	JoinedAt  time.Time          `bson:"joinedAt" json:"joined_at"`
	CreatedAt time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updated_at"`
}

// TenantLoginConfig represents login configuration for a tenant
type TenantLoginConfig struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID             string             `bson:"tenantId" json:"tenant_id"`
	AllowedIdentifiers   []string           `bson:"allowedIdentifiers" json:"allowed_identifiers"` // ["email", "phone", "username", "document_number"]
	Require2FA           bool               `bson:"require2FA" json:"require_2fa"`
	AllowRegistration    bool               `bson:"allowRegistration" json:"allow_registration"`
	CustomLogoURL        string             `bson:"customLogoUrl,omitempty" json:"custom_logo_url,omitempty"`
	CustomBackgroundURL  string             `bson:"customBackgroundUrl,omitempty" json:"custom_background_url,omitempty"`
	CustomFields         map[string]string  `bson:"customFields,omitempty" json:"custom_fields,omitempty"`
	PasswordMinLength    int                `bson:"passwordMinLength" json:"password_min_length"`
	PasswordRequireUpper bool               `bson:"passwordRequireUpper" json:"password_require_upper"`
	PasswordRequireLower bool               `bson:"passwordRequireLower" json:"password_require_lower"`
	PasswordRequireDigit bool               `bson:"passwordRequireDigit" json:"password_require_digit"`
	PasswordRequireSpec  bool               `bson:"passwordRequireSpec" json:"password_require_spec"`
	SessionTimeout       int                `bson:"sessionTimeout" json:"session_timeout"` // in minutes
	MaxLoginAttempts     int                `bson:"maxLoginAttempts" json:"max_login_attempts"`
	LockoutDuration      int                `bson:"lockoutDuration" json:"lockout_duration"` // in minutes
	CreatedAt            time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt            time.Time          `bson:"updatedAt" json:"updated_at"`
}

// IdentifierType represents the type of identifier used for login
type IdentifierType string

const (
	IdentifierTypeEmail          IdentifierType = "email"
	IdentifierTypeUsername       IdentifierType = "username"
	IdentifierTypePhone          IdentifierType = "phone"
	IdentifierTypeDocumentNumber IdentifierType = "document_number"
)

// ValidIdentifierTypes returns all valid identifier types
func ValidIdentifierTypes() []IdentifierType {
	return []IdentifierType{
		IdentifierTypeEmail,
		IdentifierTypeUsername,
		IdentifierTypePhone,
		IdentifierTypeDocumentNumber,
	}
}

// IsValidIdentifierType checks if an identifier type is valid
func IsValidIdentifierType(t string) bool {
	for _, valid := range ValidIdentifierTypes() {
		if string(valid) == t {
			return true
		}
	}
	return false
}

// DetectIdentifierType attempts to detect the type of identifier
func DetectIdentifierType(identifier string, user *User) IdentifierType {
	if user.Email == identifier {
		return IdentifierTypeEmail
	}
	if user.Username == identifier {
		return IdentifierTypeUsername
	}
	if user.Phone == identifier {
		return IdentifierTypePhone
	}
	if user.DocNumber == identifier {
		return IdentifierTypeDocumentNumber
	}
	return ""
}

// LoginAttempt tracks login attempts for rate limiting
type LoginAttempt struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Identifier string             `bson:"identifier" json:"identifier"`
	TenantID   string             `bson:"tenantId" json:"tenant_id"`
	IPAddress  string             `bson:"ipAddress" json:"ip_address"`
	Success    bool               `bson:"success" json:"success"`
	AttemptAt  time.Time          `bson:"attemptAt" json:"attempt_at"`
}

// UserLockout tracks user lockouts due to failed login attempts
type UserLockout struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     string             `bson:"userId" json:"user_id"`
	TenantID   string             `bson:"tenantId" json:"tenant_id"`
	LockedAt   time.Time          `bson:"lockedAt" json:"locked_at"`
	UnlockAt   time.Time          `bson:"unlockAt" json:"unlock_at"`
	Reason     string             `bson:"reason" json:"reason"`
	IsActive   bool               `bson:"isActive" json:"is_active"`
	CreatedAt  time.Time          `bson:"createdAt" json:"created_at"`
	ReleasedAt *time.Time         `bson:"releasedAt,omitempty" json:"released_at,omitempty"`
}
