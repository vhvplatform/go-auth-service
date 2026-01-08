package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvplatform/go-auth-service/internal/domain"
	"github.com/vhvplatform/go-auth-service/internal/repository"
	"github.com/vhvplatform/go-shared/errors"
	"github.com/vhvplatform/go-shared/jwt"
	"github.com/vhvplatform/go-shared/logger"
	"github.com/vhvplatform/go-shared/redis"
	"github.com/vhvplatform/go-shared/utils"
	"go.uber.org/zap"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo         *repository.UserRepository
	tenantRepo       *repository.TenantRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	roleRepo         *repository.RoleRepository
	jwtManager       *jwt.Manager
	redisCache       *redis.Cache
	logger           *logger.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repository.UserRepository,
	tenantRepo *repository.TenantRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	roleRepo *repository.RoleRepository,
	jwtManager *jwt.Manager,
	redisClient *redis.Client,
	log *logger.Logger,
) *AuthService {
	var redisCache *redis.Cache
	if redisClient != nil {
		redisCache = redis.NewCache(redisClient, redis.CacheConfig{
			DefaultTTL: 24 * time.Hour,
			KeyPrefix:  "auth",
		})
	}

	return &AuthService{
		userRepo:         userRepo,
		tenantRepo:       tenantRepo,
		refreshTokenRepo: refreshTokenRepo,
		roleRepo:         roleRepo,
		jwtManager:       jwtManager,
		redisCache:       redisCache,
		logger:           log,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, userReq *domain.User) (*domain.User, error) {
	// Check if user already exists by email (or other identifier)
	if userReq.Email != "" {
		existingUser, err := s.userRepo.FindByIdentifier(ctx, userReq.Email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil {
			return nil, errors.Conflict("User already exists")
		}
	}

	// Hash password
	passwordHash, err := utils.HashPassword(userReq.PasswordHash) // Assume PasswordHash field temporarily holds plain password during creation
	if err != nil {
		return nil, errors.Internal("Failed to hash password")
	}
	userReq.PasswordHash = passwordHash

	if err := s.userRepo.Create(ctx, userReq); err != nil {
		return nil, err
	}

	return userReq, nil
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, identifier, password, tenantID string) (*domain.LoginResponse, error) {
	// 1. Find tenant to check allowed login methods
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil || tenant == nil {
		return nil, errors.NotFound("Tenant not found")
	}

	// 2. Find user by identifier
	user, err := s.userRepo.FindByIdentifier(ctx, identifier)
	if err != nil || user == nil {
		return nil, errors.Unauthorized("Invalid identifier or password")
	}

	// 3. Check if login method is allowed for this tenant
	method := s.detectLoginMethod(identifier, user)
	if !utils.Contains(tenant.LoginMethods, method) {
		return nil, errors.Forbidden(fmt.Sprintf("Login method %s not allowed for this tenant", method))
	}

	// 4. Check if user belongs to the requested tenant
	belongsToTenant := false
	for _, t := range user.Tenants {
		if t == tenantID {
			belongsToTenant = true
			break
		}
	}
	if !belongsToTenant {
		return nil, errors.Forbidden("User does not belong to this tenant")
	}

	// 5. Verify password
	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, errors.Unauthorized("Invalid identifier or password")
	}

	// 6. Generate tokens
	return s.generateTokens(ctx, user, tenantID)
}

func (s *AuthService) detectLoginMethod(identifier string, user *domain.User) string {
	if identifier == user.Email {
		return "email"
	}
	if identifier == user.Username {
		return "username"
	}
	if identifier == user.Phone {
		return "phone"
	}
	if identifier == user.DocNumber {
		return "document_number"
	}
	return "unknown"
}

// ValidateToken validates a token (JWT or Opaque)
func (s *AuthService) ValidateToken(ctx context.Context, token string, tenantID string) (*domain.ValidateTokenResponse, error) {
	var userID, email string
	var roles, permissions []string

	// 1. Try to validate as Opaque token from Redis
	if s.redisCache != nil {
		var session domain.Session
		err := s.redisCache.Get(ctx, fmt.Sprintf("session:%s", token), &session)
		if err == nil {
			userID = session.UserID
			tenantID = session.TenantID
			email = session.Email
			roles = session.Roles
		}
	}

	// 2. If not found in Redis, try as JWT (for backward compatibility or internal use)
	if userID == "" {
		claims, err := s.jwtManager.ValidateToken(token)
		if err == nil {
			userID = claims.UserID
			tenantID = claims.TenantID
			email = claims.Email
			roles = claims.Roles
			permissions = claims.Permissions
		}
	}

	if userID == "" {
		return nil, errors.Unauthorized("Invalid or expired token")
	}

	// 3. Verify user exists and belongs to tenant (unless already verified by session)
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.NotFound("User not found")
	}

	if tenantID != "" {
		belongs := false
		for _, t := range user.Tenants {
			if t == tenantID {
				belongs = true
				break
			}
		}
		if !belongs {
			return nil, errors.Forbidden("User does not belong to this tenant")
		}
	}

	// 4. Get permissions if not in session/claims
	if len(permissions) == 0 {
		_, permissions, err = s.GetUserRoles(ctx, userID, tenantID)
		if err != nil {
			return nil, err
		}
	}

	return &domain.ValidateTokenResponse{
		Valid:       true,
		UserID:      userID,
		TenantID:    tenantID,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		Metadata: map[string]string{
			"user_id":   userID,
			"tenant_id": tenantID,
		},
	}, nil
}

// GetUserRoles gets roles and permissions for a user
func (s *AuthService) GetUserRoles(ctx context.Context, userID, tenantID string) ([]string, []string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, errors.NotFound("User not found")
	}

	// Roles for this specific tenant
	tenantRoles := user.TenantRoles[tenantID]
	if len(tenantRoles) == 0 {
		// Fallback to global roles if applicable or return empty
		tenantRoles = user.Roles
	}

	permissions, err := s.roleRepo.GetPermissionsForRoles(ctx, tenantRoles, tenantID)
	if err != nil {
		return nil, nil, err
	}

	return tenantRoles, permissions, nil
}

// CheckPermission checks if a user has a specific permission
func (s *AuthService) CheckPermission(ctx context.Context, userID, tenantID, permission string) (bool, error) {
	_, permissions, err := s.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}

	return utils.Contains(permissions, permission), nil
}

// Logout logs out a user by revoking token/session
func (s *AuthService) Logout(ctx context.Context, userID, token string) error {
	// Revoke refresh token (if it's a refresh token)
	if token != "" {
		_ = s.refreshTokenRepo.Revoke(ctx, token)
	}

	// Remove session from Redis
	if s.redisCache != nil && token != "" {
		_ = s.redisCache.Delete(ctx, fmt.Sprintf("session:%s", token))
	}

	return nil
}

// RefreshToken refreshes an access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*domain.LoginResponse, error) {
	// Validate refresh token exists in DB
	token, err := s.refreshTokenRepo.FindByToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, errors.Internal("Failed to refresh token")
	}
	if token == nil {
		return nil, errors.Unauthorized("Invalid refresh token")
	}

	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		return nil, errors.Internal("Failed to refresh token")
	}
	if user == nil {
		return nil, errors.Unauthorized("User not found")
	}

	// Generate new tokens
	return s.generateTokens(ctx, user, token.TenantID)
}

// generateTokens generates access and refresh tokens
func (s *AuthService) generateTokens(ctx context.Context, user *domain.User, tenantID string) (*domain.LoginResponse, error) {
	userID := user.ID.Hex()

	// Generate Opaque Access Token
	accessToken, err := utils.GenerateRandomString(32)
	if err != nil {
		return nil, errors.Internal("Failed to generate access token")
	}

	// Prepare session
	session := domain.Session{
		UserID:    userID,
		TenantID:  tenantID,
		Email:     user.Email,
		Roles:     user.TenantRoles[tenantID],
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Store in Redis
	if s.redisCache != nil {
		if err := s.redisCache.Set(ctx, fmt.Sprintf("session:%s", accessToken), session, 24*time.Hour); err != nil {
			s.logger.Error("Failed to store session in Redis", zap.Error(err))
			// Fallback to JWT if Redis fails? User requested opaque, but we should handle failure.
			// For now, return error.
			return nil, errors.Internal("Failed to store session")
		}
	}

	// Generate JWT Refresh Token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(userID, tenantID)
	if err != nil {
		return nil, errors.Internal("Failed to generate refresh token")
	}

	// Store refresh token in DB
	refreshTokenDoc := &domain.RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		TenantID:  tenantID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.refreshTokenRepo.Create(ctx, refreshTokenDoc); err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400,
		User: domain.UserInfo{
			ID:       userID,
			Email:    user.Email,
			TenantID: tenantID,
			Roles:    user.TenantRoles[tenantID],
		},
	}, nil
}
