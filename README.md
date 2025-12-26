# go-auth-service

> Authentication and Authorization Service - Part of the SaaS Framework

## Description

A comprehensive authentication and authorization service built with Go, providing secure user authentication, JWT token management, OAuth2 integration, and multi-factor authentication support. This service is designed for multi-tenant SaaS applications with enterprise-grade security features.

## Features

### Core Authentication
- **User Registration & Login**: Secure user registration with email verification and login with JWT tokens
- **Password Management**: Secure password hashing with bcrypt, password reset, and email verification flows
- **JWT Token Management**: Access and refresh token generation, validation, and revocation
- **Session Management**: Redis-based session storage with automatic expiration

### OAuth2 & SSO
- **OAuth2 Integration**: Support for Google, GitHub, and other OAuth2 providers
- **Social Login**: One-click authentication with social media accounts
- **SSO Support**: Enterprise Single Sign-On integration capabilities

### Security Features
- **Multi-Factor Authentication (MFA)**: TOTP-based two-factor authentication support
- **Token Refresh Mechanism**: Automatic token refresh with refresh tokens
- **Rate Limiting**: Brute force protection with configurable rate limits
- **Token Blacklist**: Secure logout with token revocation and blacklisting
- **Password Policies**: Configurable password complexity requirements

### Authorization
- **Role-Based Access Control (RBAC)**: Fine-grained permission management
- **Multi-tenancy**: Tenant isolation and tenant-specific role management
- **Permission System**: Granular resource and action-based permissions

## Prerequisites

- Go 1.25+
- MongoDB 4.4+ (if applicable)
- Redis 6.0+ (if applicable)
- RabbitMQ 3.9+ (if applicable)

## Installation

```bash
# Clone the repository
git clone https://github.com/vhvplatform/go-auth-service.git
cd go-auth-service

# Install dependencies
go mod download
```

## Configuration

Copy the example environment file and update with your values:

```bash
cp .env.example .env
```

See [DEPENDENCIES.md](docs/DEPENDENCIES.md) for a complete list of environment variables.

## Development

### Running Locally

```bash
# Run the service
make run

# Or with go run
go run cmd/main.go
```

### Running with Docker

```bash
# Build and run
make docker-build
make docker-run
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

### Linting

```bash
# Run linters
make lint

# Format code
make fmt
```

## Authentication Flows

### 1. User Registration Flow
```
1. Client submits registration request (email, password, tenant_id)
2. Service validates input and checks for existing user
3. Password is hashed using bcrypt
4. User record created in MongoDB with default roles
5. Access and refresh tokens generated
6. Session created in Redis
7. Tokens returned to client
```

### 2. Login Flow
```
1. Client submits credentials (email, password, tenant_id)
2. Service finds user by email and tenant
3. Password verified against stored hash
4. User account status checked (active, verified)
5. Last login timestamp updated
6. New JWT tokens generated
7. Session stored in Redis with TTL
8. Tokens and user info returned
```

### 3. Token Refresh Flow
```
1. Client sends expired access token + refresh token
2. Service validates refresh token from database
3. Checks token expiration and revocation status
4. Retrieves associated user information
5. Generates new access and refresh tokens
6. Old refresh token revoked
7. New tokens returned to client
```

### 4. Logout Flow
```
1. Client sends logout request with tokens
2. Refresh token revoked in database
3. Session removed from Redis cache
4. User logged out successfully
5. Client clears local tokens
```

### 5. OAuth2 Flow (Authorization Code)
```
1. Client redirects to OAuth provider
2. User authenticates with provider
3. Provider redirects back with authorization code
4. Service exchanges code for access token
5. User profile fetched from provider
6. User created or linked in database
7. Internal JWT tokens generated
8. User authenticated in application
```

### 6. Token Validation Flow
```
1. Client includes JWT in Authorization header
2. Middleware extracts and validates token signature
3. Token expiration checked
4. Claims extracted (user_id, tenant_id, roles)
5. Optional: Session verified in Redis
6. User context added to request
7. Request proceeds to handler
```

## JWT Token Structure

### Access Token Claims
```json
{
  "sub": "user_id",
  "tenant_id": "tenant_123",
  "email": "user@example.com",
  "roles": ["user", "admin"],
  "exp": 1640000000,
  "iat": 1639996400
}
```

### Refresh Token Claims
```json
{
  "sub": "user_id",
  "tenant_id": "tenant_123",
  "type": "refresh",
  "exp": 1640604400,
  "iat": 1639996400
}
```

## Session Management

Sessions are stored in Redis with the following structure:

```json
{
  "user_id": "507f1f77bcf86cd799439011",
  "tenant_id": "tenant_123",
  "email": "user@example.com",
  "roles": ["user"],
  "created_at": "2023-12-20T10:00:00Z",
  "expires_at": "2023-12-20T11:00:00Z"
}
```

**Session Key Pattern**: `session:{user_id}`  
**Default TTL**: 1 hour  
**Refresh**: Automatic on token refresh

## Password Security

### Password Requirements
- Minimum length: 8 characters
- Must contain: uppercase, lowercase, number, special character
- Maximum length: 128 characters
- Common passwords blocked

### Password Hashing
- Algorithm: bcrypt
- Cost factor: 12 (configurable)
- Salted automatically
- Timing-attack resistant comparison

## Rate Limiting

Protection against brute force and abuse:

| Endpoint | Limit | Window |
|----------|-------|--------|
| `/auth/login` | 5 attempts | 1 minute |
| `/auth/register` | 3 attempts | 1 hour |
| `/auth/refresh` | 10 attempts | 1 minute |
| `/auth/oauth/*` | 5 attempts | 1 minute |

## Security Best Practices

### For Developers
1. **Always use HTTPS** in production
2. **Store JWT secret securely** (environment variables, secrets manager)
3. **Rotate secrets regularly** (at least every 90 days)
4. **Implement token rotation** on refresh
5. **Use short access token expiry** (15-60 minutes)
6. **Validate all input** before processing
7. **Log authentication events** for audit trails
8. **Enable account lockout** after failed attempts

### For Deployment
1. **Use strong JWT secrets** (minimum 256 bits)
2. **Enable Redis password** authentication
3. **Restrict MongoDB access** with authentication
4. **Use TLS for database connections**
5. **Implement network policies** (firewall rules)
6. **Regular security updates** for dependencies
7. **Monitor for suspicious activity**
8. **Backup authentication data** regularly

## API Documentation

The service provides both gRPC and HTTP/REST APIs for maximum flexibility.

### Authentication Flows

#### 1. User Registration Flow
```
Client -> POST /api/v1/auth/register
       <- 201 Created (with verification email sent)
Client -> Click verification link
       -> GET /api/v1/auth/verify?token=xxx
       <- 200 OK (email verified)
```

#### 2. Login Flow
```
Client -> POST /api/v1/auth/login {email, password}
       <- 200 OK {access_token, refresh_token, expires_in}
Client -> Store tokens securely
```

#### 3. Token Refresh Flow
```
Client -> POST /api/v1/auth/refresh {refresh_token}
       <- 200 OK {access_token, refresh_token, expires_in}
```

#### 4. OAuth2 Flow
```
Client -> GET /api/v1/auth/oauth/google
       <- 302 Redirect to Google
User   -> Authorize on Google
       <- 302 Redirect to callback
Client -> GET /api/v1/auth/oauth/google/callback?code=xxx
       <- 200 OK {access_token, refresh_token, expires_in}
```

#### 5. MFA Flow
```
Client -> POST /api/v1/auth/mfa/enable
       <- 200 OK {qr_code, secret}
Client -> Scan QR code with authenticator app
       -> POST /api/v1/auth/mfa/verify {code}
       <- 200 OK (MFA enabled)
       
Login with MFA:
Client -> POST /api/v1/auth/login {email, password}
       <- 200 OK {mfa_token, requires_mfa: true}
       -> POST /api/v1/auth/mfa/authenticate {mfa_token, code}
       <- 200 OK {access_token, refresh_token}
```

#### 6. Password Reset Flow
```
Client -> POST /api/v1/auth/forgot-password {email}
       <- 200 OK (reset email sent)
Client -> Click reset link
       -> POST /api/v1/auth/reset-password {token, new_password}
       <- 200 OK (password updated)
```

### API Endpoints

See [docs/DEPENDENCIES.md](docs/DEPENDENCIES.md) for complete API documentation including gRPC endpoints and HTTP/REST endpoints.

### Security Best Practices

1. **Token Storage**: Store access tokens in memory and refresh tokens in secure HTTP-only cookies
2. **HTTPS Only**: Always use HTTPS in production to prevent token interception
3. **Token Expiration**: Access tokens expire in 15 minutes, refresh tokens in 7 days
4. **Rate Limiting**: Implement client-side rate limiting to avoid being blocked
5. **MFA**: Enable MFA for sensitive accounts
6. **Password Policy**: Enforce strong passwords (min 8 chars, uppercase, lowercase, numbers, special chars)
7. **Token Revocation**: Implement logout to revoke tokens on the server side
8. **CORS**: Configure CORS properly to prevent unauthorized domains from accessing your API

## Deployment

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for deployment instructions.

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for architecture details.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## Related Repositories

- [go-shared](https://github.com/vhvplatform/go-shared) - Shared Go libraries
- [saas-framework-go](https://github.com/vhvplatform/saas-framework-go) - Original monorepo

## License

MIT License - see [LICENSE](LICENSE) for details

## Support

- Documentation: [Wiki](https://github.com/vhvplatform/go-auth-service/wiki)
- Issues: [GitHub Issues](https://github.com/vhvplatform/go-auth-service/issues)
- Discussions: [GitHub Discussions](https://github.com/vhvplatform/go-auth-service/discussions)
