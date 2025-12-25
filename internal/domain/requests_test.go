package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterRequest_Structure(t *testing.T) {
	req := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "SecurePass123!",
		TenantID:  "tenant_001",
		FirstName: "John",
		LastName:  "Doe",
	}

	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "SecurePass123!", req.Password)
	assert.Equal(t, "tenant_001", req.TenantID)
	assert.Equal(t, "John", req.FirstName)
	assert.Equal(t, "Doe", req.LastName)
}

func TestLoginRequest_Structure(t *testing.T) {
	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "SecurePass123!",
		TenantID: "tenant_001",
	}

	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "SecurePass123!", req.Password)
	assert.Equal(t, "tenant_001", req.TenantID)
}

func TestLoginResponse_Structure(t *testing.T) {
	resp := &LoginResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		RefreshToken: "refresh_token_string",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		User: UserInfo{
			ID:       "507f1f77bcf86cd799439011",
			Email:    "test@example.com",
			TenantID: "tenant_001",
			Roles:    []string{"user"},
		},
	}

	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, 3600, resp.ExpiresIn)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, "507f1f77bcf86cd799439011", resp.User.ID)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Contains(t, resp.User.Roles, "user")
}

func TestUserInfo_Structure(t *testing.T) {
	userInfo := &UserInfo{
		ID:       "507f1f77bcf86cd799439011",
		Email:    "test@example.com",
		TenantID: "tenant_001",
		Roles:    []string{"user", "admin"},
	}

	assert.Equal(t, "507f1f77bcf86cd799439011", userInfo.ID)
	assert.Equal(t, "test@example.com", userInfo.Email)
	assert.Equal(t, "tenant_001", userInfo.TenantID)
	assert.Len(t, userInfo.Roles, 2)
	assert.Contains(t, userInfo.Roles, "user")
	assert.Contains(t, userInfo.Roles, "admin")
}

func TestRefreshTokenRequest_Structure(t *testing.T) {
	req := &RefreshTokenRequest{
		RefreshToken: "refresh_token_string",
	}

	assert.Equal(t, "refresh_token_string", req.RefreshToken)
	assert.NotEmpty(t, req.RefreshToken)
}

func TestChangePasswordRequest_Structure(t *testing.T) {
	req := &ChangePasswordRequest{
		OldPassword: "OldPassword123!",
		NewPassword: "NewPassword456!",
	}

	assert.Equal(t, "OldPassword123!", req.OldPassword)
	assert.Equal(t, "NewPassword456!", req.NewPassword)
	assert.NotEqual(t, req.OldPassword, req.NewPassword)
}

func TestResetPasswordRequest_Structure(t *testing.T) {
	req := &ResetPasswordRequest{
		Email: "test@example.com",
	}

	assert.Equal(t, "test@example.com", req.Email)
	assert.NotEmpty(t, req.Email)
}

func TestOAuthCallbackRequest_Structure(t *testing.T) {
	req := &OAuthCallbackRequest{
		Code:     "authorization_code_123",
		State:    "csrf_state_token",
		Provider: "google",
	}

	assert.Equal(t, "authorization_code_123", req.Code)
	assert.Equal(t, "csrf_state_token", req.State)
	assert.Equal(t, "google", req.Provider)
}

func TestLoginResponse_BearerToken(t *testing.T) {
	resp := &LoginResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}

	// Verify token type is Bearer
	assert.Equal(t, "Bearer", resp.TokenType)
	
	// Verify expiration is positive
	assert.Greater(t, resp.ExpiresIn, 0)
	
	// Verify tokens are not empty
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestUserInfo_MultipleRoles(t *testing.T) {
	// Test with no roles
	userNoRoles := &UserInfo{
		ID:       "user1",
		Email:    "user1@example.com",
		TenantID: "tenant_001",
		Roles:    []string{},
	}
	assert.Empty(t, userNoRoles.Roles)

	// Test with single role
	userOneRole := &UserInfo{
		ID:       "user2",
		Email:    "user2@example.com",
		TenantID: "tenant_001",
		Roles:    []string{"user"},
	}
	assert.Len(t, userOneRole.Roles, 1)

	// Test with multiple roles
	userMultipleRoles := &UserInfo{
		ID:       "user3",
		Email:    "user3@example.com",
		TenantID: "tenant_001",
		Roles:    []string{"user", "admin", "moderator"},
	}
	assert.Len(t, userMultipleRoles.Roles, 3)
}
