package grpc

import (
	"context"

	"github.com/vhvplatform/go-auth-service/internal/domain"
	"github.com/vhvplatform/go-auth-service/internal/pb"
	"github.com/vhvplatform/go-auth-service/internal/service"
	"github.com/vhvplatform/go-shared/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MultiTenantAuthServer implements the gRPC auth service with multi-tenant support
type MultiTenantAuthServer struct {
	pb.UnimplementedAuthServiceServer
	authService       *service.MultiTenantAuthService
	permissionService *service.PermissionService
	logger            *logger.Logger
}

// NewMultiTenantAuthServer creates a new gRPC auth service server
func NewMultiTenantAuthServer(
	authService *service.MultiTenantAuthService,
	permissionService *service.PermissionService,
	log *logger.Logger,
) *MultiTenantAuthServer {
	return &MultiTenantAuthServer{
		authService:       authService,
		permissionService: permissionService,
		logger:            log,
	}
}

// Login authenticates a user
func (s *MultiTenantAuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	s.logger.Info("Login request received",
		zap.String("identifier", req.Identifier),
		zap.String("tenant_id", req.TenantId))

	// Validate request
	if req.Identifier == "" {
		return nil, status.Error(codes.InvalidArgument, "identifier is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	// Attempt login
	response, err := s.authService.Login(ctx, req.Identifier, req.Password, req.TenantId)
	if err != nil {
		s.logger.Warn("Login failed",
			zap.String("identifier", req.Identifier),
			zap.String("tenant_id", req.TenantId),
			zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &pb.LoginResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		TokenType:    response.TokenType,
		ExpiresIn:    response.ExpiresIn,
	}, nil
}

// Register registers a new user
func (s *MultiTenantAuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	s.logger.Info("Register request received",
		zap.String("email", req.Email),
		zap.String("tenant_id", req.TenantId))

	// Validate request
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if req.Email == "" && req.Username == "" && req.Phone == "" && req.DocumentNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "at least one identifier (email, username, phone, or document_number) is required")
	}

	// Default roles if not provided
	roles := []string{"user"}

	// Register user
	user, err := s.authService.Register(
		ctx,
		req.Email,
		req.Username,
		req.Phone,
		req.DocumentNumber,
		req.Password,
		req.FirstName,
		req.LastName,
		req.TenantId,
		roles,
	)
	if err != nil {
		s.logger.Warn("Registration failed",
			zap.String("email", req.Email),
			zap.String("tenant_id", req.TenantId),
			zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.RegisterResponse{
		UserId:  user.ID.Hex(),
		Message: "User registered successfully",
	}, nil
}

// RefreshToken refreshes an access token
func (s *MultiTenantAuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	s.logger.Info("Refresh token request received")

	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	response, err := s.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Warn("Refresh token failed", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		TokenType:    response.TokenType,
		ExpiresIn:    response.ExpiresIn,
	}, nil
}

// ValidateToken validates a token (legacy support)
func (s *MultiTenantAuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	s.logger.Debug("Validate token request received")

	if req.Token == "" {
		return &pb.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "token is required",
		}, nil
	}

	resp, err := s.authService.VerifyToken(ctx, req.Token)
	if err != nil {
		s.logger.Debug("Token validation failed", zap.Error(err))
		return &pb.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:       true,
		UserId:      resp.UserID,
		TenantId:    resp.TenantID,
		Email:       resp.Email,
		Roles:       resp.Roles,
		Permissions: resp.Permissions,
		Metadata:    resp.Metadata,
	}, nil
}

// VerifyToken verifies an opaque token (primary method for gateway)
func (s *MultiTenantAuthServer) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	s.logger.Debug("Verify token request received")

	if req.Token == "" {
		return &pb.VerifyTokenResponse{
			Valid: false,
		}, status.Error(codes.InvalidArgument, "token is required")
	}

	resp, err := s.authService.VerifyToken(ctx, req.Token)
	if err != nil {
		s.logger.Debug("Token verification failed", zap.Error(err))
		return &pb.VerifyTokenResponse{
			Valid: false,
		}, nil
	}

	return &pb.VerifyTokenResponse{
		Valid:       true,
		UserId:      resp.UserID,
		TenantId:    resp.TenantID,
		Email:       resp.Email,
		Roles:       resp.Roles,
		Permissions: resp.Permissions,
		Metadata:    resp.Metadata,
	}, nil
}

// GetUserRoles gets roles and permissions for a user (not implemented yet)
func (s *MultiTenantAuthServer) GetUserRoles(ctx context.Context, req *pb.GetUserRolesRequest) (*pb.GetUserRolesResponse, error) {
	s.logger.Info("Get user roles request received",
		zap.String("user_id", req.UserId),
		zap.String("tenant_id", req.TenantId))

	// TODO: Implement when role repository methods are available
	return &pb.GetUserRolesResponse{
		Roles:       []string{},
		Permissions: []string{},
	}, nil
}

// CheckPermission checks if a user has a specific permission (not implemented yet)
func (s *MultiTenantAuthServer) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	s.logger.Info("Check permission request received",
		zap.String("user_id", req.UserId),
		zap.String("tenant_id", req.TenantId),
		zap.String("permission", req.Permission))

	// TODO: Implement when permission checking is available
	return &pb.CheckPermissionResponse{
		Allowed: false,
	}, nil
}

// GetTenantLoginConfig returns the login configuration for a tenant
func (s *MultiTenantAuthServer) GetTenantLoginConfig(ctx context.Context, req *pb.GetTenantLoginConfigRequest) (*pb.GetTenantLoginConfigResponse, error) {
	s.logger.Info("Get tenant login config request received",
		zap.String("tenant_id", req.TenantId))

	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	config, err := s.authService.GetTenantLoginConfig(ctx, req.TenantId)
	if err != nil {
		s.logger.Error("Failed to get tenant login config", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetTenantLoginConfigResponse{
		AllowedIdentifiers:  config.AllowedIdentifiers,
		Require2Fa:          config.Require2FA,
		AllowRegistration:   config.AllowRegistration,
		CustomLogoUrl:       config.CustomLogoURL,
		CustomBackgroundUrl: config.CustomBackgroundURL,
		CustomFields:        config.CustomFields,
	}, nil
}

// Logout logs out a user
func (s *MultiTenantAuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	s.logger.Info("Logout request received", zap.String("tenant_id", req.TenantId))

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	err := s.authService.Logout(ctx, req.Token)
	if err != nil {
		s.logger.Error("Logout failed", zap.Error(err))
		return &pb.LogoutResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	s.logger.Info("Logout successful", zap.String("session_id", req.Token))

	return &pb.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}

// CheckPermission checks if a user has a specific permission
func (s *MultiTenantAuthServer) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	s.logger.Debug("CheckPermission request",
		zap.String("user_id", req.UserId),
		zap.String("tenant_id", req.TenantId),
		zap.String("permission", req.Permission))

	// Validate request
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	if req.Permission == "" {
		return nil, status.Error(codes.InvalidArgument, "permission is required")
	}

	// Check permission
	hasPermission, err := s.permissionService.CheckPermission(ctx, req.UserId, req.TenantId, req.Permission)
	if err != nil {
		s.logger.Error("Failed to check permission",
			zap.String("user_id", req.UserId),
			zap.String("tenant_id", req.TenantId),
			zap.String("permission", req.Permission),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to check permission")
	}

	return &pb.CheckPermissionResponse{
		HasPermission: hasPermission,
	}, nil
}

// GetUserRoles gets all roles for a user in a tenant
func (s *MultiTenantAuthServer) GetUserRoles(ctx context.Context, req *pb.GetUserRolesRequest) (*pb.GetUserRolesResponse, error) {
	s.logger.Debug("GetUserRoles request",
		zap.String("user_id", req.UserId),
		zap.String("tenant_id", req.TenantId))

	// Validate request
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	// Get roles
	roles, err := s.permissionService.GetUserRoles(ctx, req.UserId, req.TenantId)
	if err != nil {
		s.logger.Error("Failed to get user roles",
			zap.String("user_id", req.UserId),
			zap.String("tenant_id", req.TenantId),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user roles")
	}

	return &pb.GetUserRolesResponse{
		Roles: roles,
	}, nil
}

// Helper function to convert domain user to proto user
func convertUserToProto(user *domain.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		Id:         user.ID.Hex(),
		Email:      user.Email,
		Username:   user.Username,
		Phone:      user.Phone,
		DocNumber:  user.DocNumber,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
