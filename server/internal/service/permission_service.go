package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vhvplatform/go-auth-service/internal/repository"
	"github.com/vhvplatform/go-shared/auth"
	"github.com/vhvplatform/go-shared/cache"
	"github.com/vhvplatform/go-shared/logger"
	"go.uber.org/zap"
)

// PermissionService handles permission checking and role management
type PermissionService struct {
	userRepo       *repository.UserRepository
	userTenantRepo *repository.UserTenantRepository
	roleRepo       *repository.RoleRepository
	cache          cache.Cache
	logger         *logger.Logger
}

// NewPermissionService creates a new permission service
func NewPermissionService(
	userRepo *repository.UserRepository,
	userTenantRepo *repository.UserTenantRepository,
	roleRepo *repository.RoleRepository,
	cacheClient cache.Cache,
	log *logger.Logger,
) *PermissionService {
	return &PermissionService{
		userRepo:       userRepo,
		userTenantRepo: userTenantRepo,
		roleRepo:       roleRepo,
		cache:          cacheClient,
		logger:         log,
	}
}

// GetUserPermissions gets all permissions for a user in a tenant
// Uses 2-level caching (L1 local, L2 Redis)
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]string, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("permissions:%s:%s", userID, tenantID)
	var cachedPermissions []string

	if s.cache != nil {
		err := s.cache.Get(ctx, cacheKey, &cachedPermissions)
		if err == nil && len(cachedPermissions) > 0 {
			s.logger.Debug("Permission cache hit",
				zap.String("user_id", userID),
				zap.String("tenant_id", tenantID))
			return cachedPermissions, nil
		}
	}

	// Cache miss, fetch from database
	s.logger.Debug("Permission cache miss, fetching from DB",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	// Get user-tenant relationship to get roles
	userTenant, err := s.userTenantRepo.FindByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user-tenant relationship: %w", err)
	}
	if userTenant == nil || !userTenant.IsActive {
		return []string{}, nil // No permissions if not in tenant
	}

	// Get permissions for all roles
	permissions, err := s.roleRepo.GetPermissionsForRoles(ctx, userTenant.Roles, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	// Remove duplicates
	permissions = removeDuplicates(permissions)

	// Cache the result (5 minutes TTL)
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, permissions, 5*time.Minute)
	}

	return permissions, nil
}

// CheckPermission checks if a user has a specific permission
func (s *PermissionService) CheckPermission(ctx context.Context, userID, tenantID, permission string) (bool, error) {
	permissions, err := s.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}

	// Check for wildcard permission (super admin)
	for _, perm := range permissions {
		if perm == "*" {
			return true, nil
		}
		if perm == permission {
			return true, nil
		}
	}

	// Check for wildcard patterns (e.g., "user.*" matches "user.read")
	permObj, err := auth.ParsePermission(permission)
	if err != nil {
		return false, nil
	}

	for _, perm := range permissions {
		userPerm, err := auth.ParsePermission(perm)
		if err != nil {
			continue
		}
		if userPerm.Matches(permObj) {
			return true, nil
		}
	}

	return false, nil
}

// CheckPermissions checks if user has all specified permissions
func (s *PermissionService) CheckPermissions(ctx context.Context, userID, tenantID string, requiredPermissions []string) (bool, []string, error) {
	permissions, err := s.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return false, nil, err
	}

	permSet, err := auth.NewPermissionSet(permissions)
	if err != nil {
		return false, nil, err
	}

	missingPermissions := []string{}
	for _, required := range requiredPermissions {
		if !permSet.Has(required) {
			missingPermissions = append(missingPermissions, required)
		}
	}

	hasAll := len(missingPermissions) == 0
	return hasAll, missingPermissions, nil
}

// CheckAnyPermission checks if user has any of the specified permissions
func (s *PermissionService) CheckAnyPermission(ctx context.Context, userID, tenantID string, requiredPermissions []string) (bool, error) {
	permissions, err := s.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}

	permSet, err := auth.NewPermissionSet(permissions)
	if err != nil {
		return false, err
	}

	return permSet.HasAny(requiredPermissions...), nil
}

// GetUserRoles gets roles for a user in a tenant
func (s *PermissionService) GetUserRoles(ctx context.Context, userID, tenantID string) ([]string, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("roles:%s:%s", userID, tenantID)
	var cachedRoles []string

	if s.cache != nil {
		err := s.cache.Get(ctx, cacheKey, &cachedRoles)
		if err == nil && len(cachedRoles) > 0 {
			return cachedRoles, nil
		}
	}

	// Cache miss, fetch from database
	userTenant, err := s.userTenantRepo.FindByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user-tenant relationship: %w", err)
	}
	if userTenant == nil || !userTenant.IsActive {
		return []string{}, nil
	}

	roles := userTenant.Roles
	if roles == nil {
		roles = []string{}
	}

	// Cache the result (5 minutes TTL)
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, roles, 5*time.Minute)
	}

	return roles, nil
}

// HasRole checks if user has a specific role
func (s *PermissionService) HasRole(ctx context.Context, userID, tenantID, role string) (bool, error) {
	roles, err := s.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}

	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}

	return false, nil
}

// InvalidateUserPermissionCache invalidates permission cache for a user
func (s *PermissionService) InvalidateUserPermissionCache(ctx context.Context, userID, tenantID string) error {
	if s.cache == nil {
		return nil
	}

	permCacheKey := fmt.Sprintf("permissions:%s:%s", userID, tenantID)
	roleCacheKey := fmt.Sprintf("roles:%s:%s", userID, tenantID)

	_ = s.cache.Delete(ctx, permCacheKey)
	_ = s.cache.Delete(ctx, roleCacheKey)

	s.logger.Info("Invalidated permission cache",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	return nil
}

// InvalidateTenantPermissionCache invalidates all permission caches for a tenant
// Called when roles/permissions are updated
func (s *PermissionService) InvalidateTenantPermissionCache(ctx context.Context, tenantID string) error {
	// This is a simplified version - in production, you'd want to track all cached keys
	// or use cache tagging/grouping
	s.logger.Info("Tenant permission cache invalidation requested",
		zap.String("tenant_id", tenantID))

	// Note: Redis/cache backend should support pattern-based deletion
	// For now, we log it and rely on TTL expiration

	return nil
}

// CreateRBACChecker creates an RBAC checker for a user
func (s *PermissionService) CreateRBACChecker(ctx context.Context, userID, tenantID string) (*auth.RBACChecker, error) {
	roles, err := s.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	return auth.NewRBACChecker(roles, permissions)
}

// Helper functions

func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
