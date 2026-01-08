package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vhvplatform/go-auth-service/internal/domain"
	"github.com/vhvplatform/go-auth-service/internal/repository"
	"github.com/vhvplatform/go-auth-service/internal/service"
	"github.com/vhvplatform/go-shared/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockCache is a mock implementation of cache.Cache
type MockCache struct {
	mock.Mock
	store map[string]interface{}
}

func NewMockCache() *MockCache {
	return &MockCache{
		store: make(map[string]interface{}),
	}
}

func (m *MockCache) Get(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	if stored, ok := m.store[key]; ok {
		// Simple copy for testing
		switch v := value.(type) {
		case *[]string:
			*v = stored.([]string)
		}
		return nil
	}
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	m.store[key] = value
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	delete(m.store, key)
	return args.Error(0)
}

func (m *MockCache) Clear(ctx context.Context) error {
	m.store = make(map[string]interface{})
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.store[key]
	return exists, nil
}

// MockUserTenantRepository
type MockUserTenantRepository struct {
	mock.Mock
}

func (m *MockUserTenantRepository) FindByUserAndTenant(ctx context.Context, userID, tenantID string) (*domain.UserTenant, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserTenant), args.Error(1)
}

// MockRoleRepository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetPermissionsForRoles(ctx context.Context, roles []string, tenantID string) ([]string, error) {
	args := m.Called(ctx, roles, tenantID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRoleRepository) FindByNames(ctx context.Context, names []string, tenantID string) ([]*domain.Role, error) {
	args := m.Called(ctx, names, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Role), args.Error(1)
}

func TestPermissionService_CheckPermission(t *testing.T) {
	// Setup
	log := logger.NewLogger()
	mockCache := NewMockCache()
	mockUserTenantRepo := &MockUserTenantRepository{}
	mockRoleRepo := &MockRoleRepository{}

	permService := service.NewPermissionService(
		nil, // UserRepository not needed for this test
		mockUserTenantRepo,
		mockRoleRepo,
		mockCache,
		log,
	)

	ctx := context.Background()
	userID := "user123"
	tenantID := "tenant123"

	t.Run("User has exact permission", func(t *testing.T) {
		// Setup mocks
		userTenant := &domain.UserTenant{
			ID:       primitive.NewObjectID(),
			UserID:   userID,
			TenantID: tenantID,
			Roles:    []string{"admin"},
			IsActive: true,
		}
		mockUserTenantRepo.On("FindByUserAndTenant", ctx, userID, tenantID).Return(userTenant, nil).Once()
		mockRoleRepo.On("GetPermissionsForRoles", ctx, []string{"admin"}, tenantID).
			Return([]string{"user.read", "user.write", "tenant.read"}, nil).Once()
		mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(nil).Once()
		mockCache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		// Test
		hasPermission, err := permService.CheckPermission(ctx, userID, tenantID, "user.read")

		// Assert
		assert.NoError(t, err)
		assert.True(t, hasPermission)
		mockUserTenantRepo.AssertExpectations(t)
		mockRoleRepo.AssertExpectations(t)
	})

	t.Run("User has wildcard permission", func(t *testing.T) {
		// Clear cache
		mockCache.Clear(ctx)

		userTenant := &domain.UserTenant{
			ID:       primitive.NewObjectID(),
			UserID:   userID,
			TenantID: tenantID,
			Roles:    []string{"super_admin"},
			IsActive: true,
		}
		mockUserTenantRepo.On("FindByUserAndTenant", ctx, userID, tenantID).Return(userTenant, nil).Once()
		mockRoleRepo.On("GetPermissionsForRoles", ctx, []string{"super_admin"}, tenantID).
			Return([]string{"*"}, nil).Once()
		mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(nil).Once()
		mockCache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		// Test - any permission should pass
		hasPermission, err := permService.CheckPermission(ctx, userID, tenantID, "anything.can.do")

		// Assert
		assert.NoError(t, err)
		assert.True(t, hasPermission)
	})

	t.Run("User has resource wildcard permission", func(t *testing.T) {
		mockCache.Clear(ctx)

		userTenant := &domain.UserTenant{
			ID:       primitive.NewObjectID(),
			UserID:   userID,
			TenantID: tenantID,
			Roles:    []string{"user_admin"},
			IsActive: true,
		}
		mockUserTenantRepo.On("FindByUserAndTenant", ctx, userID, tenantID).Return(userTenant, nil).Once()
		mockRoleRepo.On("GetPermissionsForRoles", ctx, []string{"user_admin"}, tenantID).
			Return([]string{"user.*"}, nil).Once()
		mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(nil).Once()
		mockCache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		// Test - should match user.read, user.write, etc.
		hasPermission, err := permService.CheckPermission(ctx, userID, tenantID, "user.delete")

		// Assert
		assert.NoError(t, err)
		assert.True(t, hasPermission)
	})

	t.Run("User does not have permission", func(t *testing.T) {
		mockCache.Clear(ctx)

		userTenant := &domain.UserTenant{
			ID:       primitive.NewObjectID(),
			UserID:   userID,
			TenantID: tenantID,
			Roles:    []string{"viewer"},
			IsActive: true,
		}
		mockUserTenantRepo.On("FindByUserAndTenant", ctx, userID, tenantID).Return(userTenant, nil).Once()
		mockRoleRepo.On("GetPermissionsForRoles", ctx, []string{"viewer"}, tenantID).
			Return([]string{"user.read", "tenant.read"}, nil).Once()
		mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(nil).Once()
		mockCache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		// Test - user doesn't have write permission
		hasPermission, err := permService.CheckPermission(ctx, userID, tenantID, "user.write")

		// Assert
		assert.NoError(t, err)
		assert.False(t, hasPermission)
	})

	t.Run("User not in tenant", func(t *testing.T) {
		mockCache.Clear(ctx)

		mockUserTenantRepo.On("FindByUserAndTenant", ctx, userID, tenantID).Return(nil, nil).Once()
		mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(nil).Once()

		// Test
		hasPermission, err := permService.CheckPermission(ctx, userID, tenantID, "user.read")

		// Assert
		assert.NoError(t, err)
		assert.False(t, hasPermission)
	})
}

func TestPermissionService_CheckPermissions(t *testing.T) {
	log := logger.NewLogger()
	mockCache := NewMockCache()
	mockUserTenantRepo := &MockUserTenantRepository{}
	mockRoleRepo := &MockRoleRepository{}

	permService := service.NewPermissionService(
		nil,
		mockUserTenantRepo,
		mockRoleRepo,
		mockCache,
		log,
	)

	ctx := context.Background()
	userID := "user123"
	tenantID := "tenant123"

	t.Run("User has all permissions", func(t *testing.T) {
		mockCache.Clear(ctx)

		userTenant := &domain.UserTenant{
			ID:       primitive.NewObjectID(),
			UserID:   userID,
			TenantID: tenantID,
			Roles:    []string{"admin"},
			IsActive: true,
		}
		mockUserTenantRepo.On("FindByUserAndTenant", ctx, userID, tenantID).Return(userTenant, nil).Once()
		mockRoleRepo.On("GetPermissionsForRoles", ctx, []string{"admin"}, tenantID).
			Return([]string{"user.read", "user.write", "tenant.read", "tenant.write"}, nil).Once()
		mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(nil).Once()
		mockCache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		// Test
		hasAll, missing, err := permService.CheckPermissions(ctx, userID, tenantID, []string{"user.read", "user.write"})

		// Assert
		assert.NoError(t, err)
		assert.True(t, hasAll)
		assert.Empty(t, missing)
	})

	t.Run("User missing some permissions", func(t *testing.T) {
		mockCache.Clear(ctx)

		userTenant := &domain.UserTenant{
			ID:       primitive.NewObjectID(),
			UserID:   userID,
			TenantID: tenantID,
			Roles:    []string{"viewer"},
			IsActive: true,
		}
		mockUserTenantRepo.On("FindByUserAndTenant", ctx, userID, tenantID).Return(userTenant, nil).Once()
		mockRoleRepo.On("GetPermissionsForRoles", ctx, []string{"viewer"}, tenantID).
			Return([]string{"user.read", "tenant.read"}, nil).Once()
		mockCache.On("Get", ctx, mock.Anything, mock.Anything).Return(nil).Once()
		mockCache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		// Test
		hasAll, missing, err := permService.CheckPermissions(ctx, userID, tenantID, 
			[]string{"user.read", "user.write", "user.delete"})

		// Assert
		assert.NoError(t, err)
		assert.False(t, hasAll)
		assert.Equal(t, []string{"user.write", "user.delete"}, missing)
	})
}

func TestPermissionService_GetUserPermissions_Cache(t *testing.T) {
	log := logger.NewLogger()
	mockCache := NewMockCache()
	mockUserTenantRepo := &MockUserTenantRepository{}
	mockRoleRepo := &MockRoleRepository{}

	permService := service.NewPermissionService(
		nil,
		mockUserTenantRepo,
		mockRoleRepo,
		mockCache,
		log,
	)

	ctx := context.Background()
	userID := "user123"
	tenantID := "tenant123"

	t.Run("Cache hit - should not query database", func(t *testing.T) {
		mockCache.Clear(ctx)

		// Pre-populate cache
		cachedPerms := []string{"user.read", "user.write"}
		cacheKey := "permissions:user123:tenant123"
		mockCache.store[cacheKey] = cachedPerms

		// Mock cache to return stored value
		mockCache.On("Get", ctx, cacheKey, mock.Anything).Run(func(args mock.Arguments) {
			// Simulate cache returning the value
			val := args.Get(2).(*[]string)
			*val = cachedPerms
		}).Return(nil).Once()

		// Test - should use cache, not call repository
		permissions, err := permService.GetUserPermissions(ctx, userID, tenantID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, cachedPerms, permissions)
		// Repositories should NOT be called
		mockUserTenantRepo.AssertNotCalled(t, "FindByUserAndTenant")
		mockRoleRepo.AssertNotCalled(t, "GetPermissionsForRoles")
	})
}
