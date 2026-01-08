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

// MultiTenantAuthService handles multi-tenant authentication business logic
type MultiTenantAuthService struct {
	userRepo              *repository.UserRepository
	userTenantRepo        *repository.UserTenantRepository
	tenantLoginConfigRepo *repository.TenantLoginConfigRepository
	refreshTokenRepo      *repository.RefreshTokenRepository
	roleRepo              *repository.RoleRepository
	jwtManager            *jwt.Manager
	redisCache            *redis.Cache
	logger                *logger.Logger
}

// NewMultiTenantAuthService creates a new multi-tenant auth service
func NewMultiTenantAuthService(
	userRepo *repository.UserRepository,
	userTenantRepo *repository.UserTenantRepository,
	tenantLoginConfigRepo *repository.TenantLoginConfigRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	roleRepo *repository.RoleRepository,
	jwtManager *jwt.Manager,
	redisClient *redis.Client,
	log *logger.Logger,
) *MultiTenantAuthService {
	var redisCache *redis.Cache
	if redisClient != nil {
		redisCache = redis.NewCache(redisClient, redis.CacheConfig{
			DefaultTTL: 24 * time.Hour,
			KeyPrefix:  "auth",
		})
	}

	return &MultiTenantAuthService{
		userRepo:              userRepo,
		userTenantRepo:        userTenantRepo,
		tenantLoginConfigRepo: tenantLoginConfigRepo,
		refreshTokenRepo:      refreshTokenRepo,
		roleRepo:              roleRepo,
		jwtManager:            jwtManager,
		redisCache:            redisCache,
		logger:                log,
	}
}

// Register registers a new user with initial tenant
func (s *MultiTenantAuthService) Register(ctx context.Context, email, username, phone, docNumber, password, firstName, lastName, tenantID string, roles []string) (*domain.User, error) {
	// 1. Validate tenant and check if registration is allowed
	loginConfig, err := s.tenantLoginConfigRepo.FindByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !loginConfig.AllowRegistration {
		return nil, errors.Forbidden("Registration is not allowed for this tenant")
	}

	// 2. Validate password requirements
	if err := s.validatePassword(password, loginConfig); err != nil {
		return nil, err
	}

	// 3. Check if user already exists (by any identifier)
	if email != "" {
		existingUser, _ := s.userRepo.FindByIdentifier(ctx, email)
		if existingUser != nil {
			return nil, errors.Conflict("Email already exists")
		}
	}
	if username != "" {
		existingUser, _ := s.userRepo.FindByIdentifier(ctx, username)
		if existingUser != nil {
			return nil, errors.Conflict("Username already exists")
		}
	}
	if phone != "" {
		existingUser, _ := s.userRepo.FindByIdentifier(ctx, phone)
		if existingUser != nil {
			return nil, errors.Conflict("Phone already exists")
		}
	}
	if docNumber != "" {
		existingUser, _ := s.userRepo.FindByIdentifier(ctx, docNumber)
		if existingUser != nil {
			return nil, errors.Conflict("Document number already exists")
		}
	}

	// 4. Hash password
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.Internal("Failed to hash password")
	}

	// 5. Create user
	user := &domain.User{
		Email:        email,
		Username:     username,
		Phone:        phone,
		DocNumber:    docNumber,
		PasswordHash: passwordHash,
		IsActive:     true,
		IsVerified:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 6. Create user-tenant relationship
	if roles == nil || len(roles) == 0 {
		roles = []string{"user"} // Default role
	}

	userTenant := &domain.UserTenant{
		UserID:   user.ID.Hex(),
		TenantID: tenantID,
		Roles:    roles,
		IsActive: true,
	}

	if err := s.userTenantRepo.Create(ctx, userTenant); err != nil {
		s.logger.Error("Failed to create user-tenant relationship", zap.Error(err))
		// User created but tenant relationship failed - log for manual intervention
	}

	s.logger.Info("User registered successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("tenant_id", tenantID),
		zap.String("email", email))

	return user, nil
}

// Login authenticates a user with multi-tenant support
func (s *MultiTenantAuthService) Login(ctx context.Context, identifier, password, tenantID string) (*domain.LoginResponse, error) {
	// 1. Get tenant login configuration
	loginConfig, err := s.tenantLoginConfigRepo.FindByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 2. Find user by identifier
	user, err := s.userRepo.FindByIdentifier(ctx, identifier)
	if err != nil {
		return nil, errors.Unauthorized("Invalid credentials")
	}
	if user == nil {
		return nil, errors.Unauthorized("Invalid credentials")
	}

	// 3. Detect and validate login method
	identifierType := domain.DetectIdentifierType(identifier, user)
	if identifierType == "" {
		return nil, errors.Unauthorized("Invalid credentials")
	}

	// Check if this identifier type is allowed for this tenant
	allowed := false
	for _, allowedType := range loginConfig.AllowedIdentifiers {
		if allowedType == string(identifierType) {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, errors.Forbidden(fmt.Sprintf("Login with %s is not allowed for this tenant", identifierType))
	}

	// 4. Check if user belongs to the tenant
	userTenant, err := s.userTenantRepo.FindByUserAndTenant(ctx, user.ID.Hex(), tenantID)
	if err != nil {
		return nil, err
	}
	if userTenant == nil {
		return nil, errors.Forbidden("User does not have access to this tenant")
	}
	if !userTenant.IsActive {
		return nil, errors.Forbidden("User access to this tenant is deactivated")
	}

	// 5. Check if user is active
	if !user.IsActive {
		return nil, errors.Forbidden("User account is deactivated")
	}

	// 6. Verify password
	if !utils.CheckPassword(password, user.PasswordHash) {
		// TODO: Track failed login attempts
		return nil, errors.Unauthorized("Invalid credentials")
	}

	// 7. Get user roles and permissions for this tenant
	roles := userTenant.Roles
	permissions, err := s.roleRepo.GetPermissionsForRoles(ctx, roles, tenantID)
	if err != nil {
		s.logger.Error("Failed to get permissions", zap.Error(err))
		permissions = []string{} // Continue with empty permissions
	}

	// 8. Generate tokens
	response, err := s.generateTokens(ctx, user, tenantID, roles, permissions)
	if err != nil {
		return nil, err
	}

	// 9. Update last login time
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID.Hex())

	s.logger.Info("User logged in successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("tenant_id", tenantID),
		zap.String("identifier_type", string(identifierType)))

	return response, nil
}

// VerifyToken verifies an opaque token and returns user information
func (s *MultiTenantAuthService) VerifyToken(ctx context.Context, token string) (*domain.ValidateTokenResponse, error) {
	if s.redisCache == nil {
		return nil, errors.Internal("Session store not available")
	}

	// Try to get session from Redis
	var session domain.Session
	err := s.redisCache.Get(ctx, fmt.Sprintf("session:%s", token), &session)
	if err != nil {
		return nil, errors.Unauthorized("Invalid or expired token")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		_ = s.redisCache.Delete(ctx, fmt.Sprintf("session:%s", token))
		return nil, errors.Unauthorized("Token expired")
	}

	// Get full user information to ensure user still exists and is active
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil || user == nil {
		return nil, errors.Unauthorized("User not found")
	}

	if !user.IsActive {
		return nil, errors.Forbidden("User account is deactivated")
	}

	// Verify user still has access to tenant
	userTenant, err := s.userTenantRepo.FindByUserAndTenant(ctx, session.UserID, session.TenantID)
	if err != nil || userTenant == nil || !userTenant.IsActive {
		return nil, errors.Forbidden("User does not have access to this tenant")
	}

	// Get permissions
	permissions, err := s.roleRepo.GetPermissionsForRoles(ctx, session.Roles, session.TenantID)
	if err != nil {
		permissions = []string{}
	}

	return &domain.ValidateTokenResponse{
		Valid:       true,
		UserID:      session.UserID,
		TenantID:    session.TenantID,
		Email:       session.Email,
		Roles:       session.Roles,
		Permissions: permissions,
		Metadata: map[string]string{
			"user_id":   session.UserID,
			"tenant_id": session.TenantID,
		},
	}, nil
}

// GetTenantLoginConfig returns the login configuration for a tenant
func (s *MultiTenantAuthService) GetTenantLoginConfig(ctx context.Context, tenantID string) (*domain.TenantLoginConfig, error) {
	config, err := s.tenantLoginConfigRepo.FindByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetUserTenants returns all tenants a user belongs to
func (s *MultiTenantAuthService) GetUserTenants(ctx context.Context, userID string) ([]*domain.UserTenant, error) {
	userTenants, err := s.userTenantRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return userTenants, nil
}

// AddUserToTenant adds a user to a tenant with specified roles
func (s *MultiTenantAuthService) AddUserToTenant(ctx context.Context, userID, tenantID string, roles []string) error {
	// Check if relationship already exists
	existing, err := s.userTenantRepo.FindByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return err
	}

	if existing != nil {
		// Already exists, update roles
		return s.userTenantRepo.UpdateRoles(ctx, userID, tenantID, roles)
	}

	// Create new relationship
	userTenant := &domain.UserTenant{
		UserID:   userID,
		TenantID: tenantID,
		Roles:    roles,
		IsActive: true,
	}

	return s.userTenantRepo.Create(ctx, userTenant)
}

// RemoveUserFromTenant removes a user from a tenant
func (s *MultiTenantAuthService) RemoveUserFromTenant(ctx context.Context, userID, tenantID string) error {
	return s.userTenantRepo.Deactivate(ctx, userID, tenantID)
}

// RefreshToken refreshes an access token using a refresh token
func (s *MultiTenantAuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*domain.LoginResponse, error) {
	// Validate refresh token exists in DB
	refreshToken, err := s.refreshTokenRepo.FindByToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, errors.Internal("Failed to refresh token")
	}
	if refreshToken == nil || refreshToken.RevokedAt != nil {
		return nil, errors.Unauthorized("Invalid refresh token")
	}

	// Check expiration
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, errors.Unauthorized("Refresh token expired")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, refreshToken.UserID)
	if err != nil || user == nil {
		return nil, errors.Unauthorized("User not found")
	}

	// Get user-tenant relationship
	userTenant, err := s.userTenantRepo.FindByUserAndTenant(ctx, refreshToken.UserID, refreshToken.TenantID)
	if err != nil || userTenant == nil || !userTenant.IsActive {
		return nil, errors.Forbidden("User does not have access to this tenant")
	}

	// Get permissions
	permissions, err := s.roleRepo.GetPermissionsForRoles(ctx, userTenant.Roles, refreshToken.TenantID)
	if err != nil {
		permissions = []string{}
	}

	// Generate new tokens
	return s.generateTokens(ctx, user, refreshToken.TenantID, userTenant.Roles, permissions)
}

// Logout invalidates a token
func (s *MultiTenantAuthService) Logout(ctx context.Context, token string) error {
	if s.redisCache != nil {
		_ = s.redisCache.Delete(ctx, fmt.Sprintf("session:%s", token))
	}
	return nil
}

// generateTokens generates opaque access token and JWT refresh token
func (s *MultiTenantAuthService) generateTokens(ctx context.Context, user *domain.User, tenantID string, roles, permissions []string) (*domain.LoginResponse, error) {
	userID := user.ID.Hex()

	// Generate Opaque Access Token (random string)
	accessToken, err := utils.GenerateRandomString(32)
	if err != nil {
		return nil, errors.Internal("Failed to generate access token")
	}

	// Create session
	session := domain.Session{
		UserID:    userID,
		TenantID:  tenantID,
		Email:     user.Email,
		Roles:     roles,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Store session in Redis
	if s.redisCache != nil {
		if err := s.redisCache.Set(ctx, fmt.Sprintf("session:%s", accessToken), session, 24*time.Hour); err != nil {
			s.logger.Error("Failed to store session in Redis", zap.Error(err))
			return nil, errors.Internal("Failed to create session")
		}
	}

	// Generate JWT Refresh Token
	refreshTokenStr, err := s.jwtManager.GenerateToken(userID, tenantID, user.Email, roles, permissions)
	if err != nil {
		return nil, errors.Internal("Failed to generate refresh token")
	}

	// Store refresh token in DB
	refreshToken := &domain.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenStr,
		TenantID:  tenantID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err))
		// Continue anyway, user can re-login
	}

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24 hours
		User: domain.UserInfo{
			ID:       userID,
			Email:    user.Email,
			TenantID: tenantID,
			Roles:    roles,
		},
	}, nil
}

// validatePassword validates password against tenant requirements
func (s *MultiTenantAuthService) validatePassword(password string, config *domain.TenantLoginConfig) error {
	if len(password) < config.PasswordMinLength {
		return errors.BadRequest(fmt.Sprintf("Password must be at least %d characters long", config.PasswordMinLength))
	}

	if config.PasswordRequireUpper && !utils.ContainsUppercase(password) {
		return errors.BadRequest("Password must contain at least one uppercase letter")
	}

	if config.PasswordRequireLower && !utils.ContainsLowercase(password) {
		return errors.BadRequest("Password must contain at least one lowercase letter")
	}

	if config.PasswordRequireDigit && !utils.ContainsDigit(password) {
		return errors.BadRequest("Password must contain at least one digit")
	}

	if config.PasswordRequireSpec && !utils.ContainsSpecialChar(password) {
		return errors.BadRequest("Password must contain at least one special character")
	}

	return nil
}
