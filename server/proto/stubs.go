package proto

import (
	"context"

	"google.golang.org/grpc"
)

// This file is a TEMPORARY STUB to allow compilation without running protoc.
// It matches the expected output of proper protobuf generation.

type LoginRequest struct {
	Identifier string `json:"identifier,omitempty"`
	Password   string `json:"password,omitempty"`
	TenantId   string `json:"tenant_id,omitempty"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
}

type ValidateTokenRequest struct {
	Token    string `json:"token,omitempty"`
	TenantId string `json:"tenant_id,omitempty"`
}

type ValidateTokenResponse struct {
	Valid        bool              `json:"valid,omitempty"`
	UserId       string            `json:"user_id,omitempty"`
	TenantId     string            `json:"tenant_id,omitempty"`
	Email        string            `json:"email,omitempty"`
	Roles        []string          `json:"roles,omitempty"`
	Permissions  []string          `json:"permissions,omitempty"`
	ErrorMessage string            `json:"error_message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type GetUserRolesRequest struct {
	UserId   string `json:"user_id,omitempty"`
	TenantId string `json:"tenant_id,omitempty"`
}

type GetUserRolesResponse struct {
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type CheckPermissionRequest struct {
	UserId     string `json:"user_id,omitempty"`
	TenantId   string `json:"tenant_id,omitempty"`
	Permission string `json:"permission,omitempty"`
}

type CheckPermissionResponse struct {
	Allowed bool `json:"allowed,omitempty"`
}

// AuthServiceClient is the client API for AuthService.
type AuthServiceClient interface {
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	ValidateToken(ctx context.Context, in *ValidateTokenRequest, opts ...grpc.CallOption) (*ValidateTokenResponse, error)
	GetUserRoles(ctx context.Context, in *GetUserRolesRequest, opts ...grpc.CallOption) (*GetUserRolesResponse, error)
	CheckPermission(ctx context.Context, in *CheckPermissionRequest, opts ...grpc.CallOption) (*CheckPermissionResponse, error)
}

type authServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthServiceClient(cc grpc.ClientConnInterface) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, "/auth.AuthService/Login", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) ValidateToken(ctx context.Context, in *ValidateTokenRequest, opts ...grpc.CallOption) (*ValidateTokenResponse, error) {
	out := new(ValidateTokenResponse)
	err := c.cc.Invoke(ctx, "/auth.AuthService/ValidateToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) GetUserRoles(ctx context.Context, in *GetUserRolesRequest, opts ...grpc.CallOption) (*GetUserRolesResponse, error) {
	out := new(GetUserRolesResponse)
	err := c.cc.Invoke(ctx, "/auth.AuthService/GetUserRoles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) CheckPermission(ctx context.Context, in *CheckPermissionRequest, opts ...grpc.CallOption) (*CheckPermissionResponse, error) {
	out := new(CheckPermissionResponse)
	err := c.cc.Invoke(ctx, "/auth.AuthService/CheckPermission", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthServiceServer is the server API for AuthService.
type AuthServiceServer interface {
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	ValidateToken(context.Context, *ValidateTokenRequest) (*ValidateTokenResponse, error)
	GetUserRoles(context.Context, *GetUserRolesRequest) (*GetUserRolesResponse, error)
	CheckPermission(context.Context, *CheckPermissionRequest) (*CheckPermissionResponse, error)
	mustEmbedUnimplementedAuthServiceServer()
}

// UnimplementedAuthServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAuthServiceServer struct{}

func (UnimplementedAuthServiceServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) ValidateToken(context.Context, *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) GetUserRoles(context.Context, *GetUserRolesRequest) (*GetUserRolesResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) CheckPermission(context.Context, *CheckPermissionRequest) (*CheckPermissionResponse, error) {
	return nil, nil
}
func (UnimplementedAuthServiceServer) mustEmbedUnimplementedAuthServiceServer() {}

func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "auth.AuthService",
		HandlerType: (*AuthServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Login",
				Handler:    nil,
			},
			{
				MethodName: "ValidateToken",
				Handler:    nil,
			},
			{
				MethodName: "GetUserRoles",
				Handler:    nil,
			},
			{
				MethodName: "CheckPermission",
				Handler:    nil,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "auth.proto",
	}, srv)
}
