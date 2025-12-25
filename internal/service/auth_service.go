package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/longvhv/saas-shared-go/errors"
	"github.com/longvhv/saas-shared-go/jwt"
	"github.com/longvhv/saas-shared-go/logger"
	"github.com/longvhv/saas-shared-go/redis"
	"github.com/longvhv/saas-shared-go/utils"
	"github.com/vhvcorp/go-auth-service/internal/domain"
	"github.com/vhvcorp/go-auth-service/internal/repository"
	"go.uber.org/zap"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	roleRepo         *repository.RoleRepository
	jwtManager       *jwt.Manager
	redisClient      *redis.Client
	logger           *logger.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repository.UserRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	roleRepo *repository.RoleRepository,
	jwtManager *jwt.Manager,
	redisClient *redis.Client,
	log *logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		roleRepo:         roleRepo,
		jwtManager:       jwtManager,
		redisClient:      redisClient,
		logger:           log,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.LoginResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmailAndTenant(ctx, req.Email, req.TenantID)
	if err != nil {
		s.logger.Error("Failed to check existing user", zap.Error(err))
		return nil, errors.Internal("Failed to register user")
	}
	if existingUser != nil {
		return nil, errors.Conflict("User already exists with this email")
	}
	
	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, errors.Internal("Failed to register user")
	}
	
	// Create user
	user := &domain.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		TenantID:     req.TenantID,
		Roles:        []string{"user"}, // Default role
		IsActive:     true,
		IsVerified:   false,
	}
	
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, errors.Internal("Failed to register user")
	}
	
	s.logger.Info("User registered successfully", 
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email),
	)
	
	// Generate tokens
	return s.generateTokens(ctx, user)
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// Find user
	var user *domain.User
	var err error
	
	if req.TenantID != "" {
		user, err = s.userRepo.FindByEmailAndTenant(ctx, req.Email, req.TenantID)
	} else {
		user, err = s.userRepo.FindByEmail(ctx, req.Email)
	}
	
	if err != nil {
		s.logger.Error("Failed to find user", zap.Error(err))
		return nil, errors.Internal("Failed to login")
	}
	if user == nil {
		return nil, errors.Unauthorized("Invalid email or password")
	}
	
	// Check if user is active
	if !user.IsActive {
		return nil, errors.Forbidden("User account is deactivated")
	}
	
	// Verify password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.Unauthorized("Invalid email or password")
	}
	
	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID.Hex()); err != nil {
		s.logger.Warn("Failed to update last login", zap.Error(err))
	}
	
	s.logger.Info("User logged in successfully", 
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email),
	)
	
	// Generate tokens
	return s.generateTokens(ctx, user)
}

// Logout logs out a user by revoking refresh token
func (s *AuthService) Logout(ctx context.Context, userID, refreshToken string) error {
	// Revoke refresh token
	if refreshToken != "" {
		if err := s.refreshTokenRepo.Revoke(ctx, refreshToken); err != nil {
			s.logger.Error("Failed to revoke refresh token", zap.Error(err))
		}
	}
	
	// Remove session from Redis
	sessionKey := fmt.Sprintf("session:%s", userID)
	if err := s.redisClient.Delete(ctx, sessionKey); err != nil {
		s.logger.Error("Failed to delete session", zap.Error(err))
	}
	
	s.logger.Info("User logged out", zap.String("user_id", userID))
	return nil
}

// RefreshToken refreshes an access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*domain.LoginResponse, error) {
	// Validate refresh token exists in DB
	token, err := s.refreshTokenRepo.FindByToken(ctx, refreshTokenStr)
	if err != nil {
		s.logger.Error("Failed to find refresh token", zap.Error(err))
		return nil, errors.Internal("Failed to refresh token")
	}
	if token == nil {
		return nil, errors.Unauthorized("Invalid refresh token")
	}
	
	// Get user
	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		s.logger.Error("Failed to find user", zap.Error(err))
		return nil, errors.Internal("Failed to refresh token")
	}
	if user == nil {
		return nil, errors.Unauthorized("User not found")
	}
	
	// Generate new tokens
	return s.generateTokens(ctx, user)
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*jwt.Claims, error) {
	claims, err := s.jwtManager.ValidateToken(tokenStr)
	if err != nil {
		return nil, errors.Unauthorized("Invalid or expired token")
	}
	return claims, nil
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
	
	permissions, err := s.roleRepo.GetPermissionsForRoles(ctx, user.Roles, tenantID)
	if err != nil {
		return nil, nil, err
	}
	
	return user.Roles, permissions, nil
}

// CheckPermission checks if a user has a specific permission
func (s *AuthService) CheckPermission(ctx context.Context, userID, tenantID, permission string) (bool, error) {
	_, permissions, err := s.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}
	
	return utils.Contains(permissions, permission), nil
}

// generateTokens generates access and refresh tokens
func (s *AuthService) generateTokens(ctx context.Context, user *domain.User) (*domain.LoginResponse, error) {
	userID := user.ID.Hex()
	
	// Generate access token
	accessToken, err := s.jwtManager.GenerateToken(userID, user.TenantID, user.Email, user.Roles)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, errors.Internal("Failed to generate tokens")
	}
	
	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(userID, user.TenantID)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, errors.Internal("Failed to generate tokens")
	}
	
	// Store refresh token in database
	refreshTokenDoc := &domain.RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}
	if err := s.refreshTokenRepo.Create(ctx, refreshTokenDoc); err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err))
	}
	
	// Store session in Redis
	session := &domain.Session{
		UserID:    userID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		Roles:     user.Roles,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	sessionData, _ := json.Marshal(session)
	sessionKey := fmt.Sprintf("session:%s", userID)
	if err := s.redisClient.Set(ctx, sessionKey, sessionData, 1*time.Hour); err != nil {
		s.logger.Warn("Failed to store session in Redis", zap.Error(err))
	}
	
	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
		TokenType:    "Bearer",
		User: domain.UserInfo{
			ID:       userID,
			Email:    user.Email,
			TenantID: user.TenantID,
			Roles:    user.Roles,
		},
	}, nil
}
