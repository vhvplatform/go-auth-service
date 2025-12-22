package grpc

import (
	"github.com/longvhv/saas-framework-go/pkg/logger"
	"github.com/longvhv/saas-framework-go/services/auth-service/internal/service"
	// pb "github.com/longvhv/saas-framework-go/services/auth-service/proto"
)

// AuthServiceServer implements the gRPC auth service
type AuthServiceServer struct {
	// pb.UnimplementedAuthServiceServer
	authService *service.AuthService
	logger      *logger.Logger
}

// NewAuthServiceServer creates a new gRPC auth service server
func NewAuthServiceServer(authService *service.AuthService, log *logger.Logger) *AuthServiceServer {
	return &AuthServiceServer{
		authService: authService,
		logger:      log,
	}
}

// Note: gRPC methods are commented out until protobuf code is generated
// Run `make proto` to generate the protobuf code, then uncomment the methods below

/*
// ValidateToken validates a JWT token
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := s.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		s.logger.Warn("Token validation failed", zap.Error(err))
		return &pb.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		}, nil
	}
	
	return &pb.ValidateTokenResponse{
		Valid:    true,
		UserId:   claims.UserID,
		TenantId: claims.TenantID,
		Email:    claims.Email,
		Roles:    claims.Roles,
	}, nil
}

// GetUserRoles gets roles and permissions for a user
func (s *AuthServiceServer) GetUserRoles(ctx context.Context, req *pb.GetUserRolesRequest) (*pb.GetUserRolesResponse, error) {
	roles, permissions, err := s.authService.GetUserRoles(ctx, req.UserId, req.TenantId)
	if err != nil {
		s.logger.Error("Failed to get user roles", zap.Error(err))
		return nil, err
	}
	
	return &pb.GetUserRolesResponse{
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

// CheckPermission checks if a user has a specific permission
func (s *AuthServiceServer) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	allowed, err := s.authService.CheckPermission(ctx, req.UserId, req.TenantId, req.Permission)
	if err != nil {
		s.logger.Error("Failed to check permission", zap.Error(err))
		return nil, err
	}
	
	return &pb.CheckPermissionResponse{
		Allowed: allowed,
	}, nil
}
*/
