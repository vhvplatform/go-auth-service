# go-auth-service

> Authentication and Authorization Service - Part of the SaaS Framework

## Description

A robust, production-ready authentication and authorization microservice built with Go. This service provides comprehensive authentication flows including JWT token management, OAuth2 integration, 2FA support, session management, and role-based access control (RBAC).

## Features

### Core Authentication
- 🔐 **User Registration & Login**: Secure user authentication with password hashing (bcrypt)
- 🎟️ **JWT Token Management**: Access and refresh token generation, validation, and rotation
- 🔄 **Token Refresh**: Seamless token refresh mechanism for continuous sessions
- 🚪 **Logout**: Secure token revocation and session cleanup

### OAuth2 Integration
- 🌐 **OAuth2 Support**: Ready for Google, GitHub, and custom OAuth providers
- 🔗 **Account Linking**: Link multiple OAuth accounts to a single user
- 🔓 **Social Login**: Simplified login flow with popular identity providers

### Security Features
- 🛡️ **Password Security**: bcrypt hashing with configurable cost factor
- 🔒 **Session Management**: Redis-based session storage with TTL
- 🎯 **Role-Based Access Control (RBAC)**: Fine-grained permissions system
- 🚦 **Rate Limiting**: Protection against brute force attacks
- 🔍 **Audit Logging**: Comprehensive authentication event logging
- 🛑 **Account Lockout**: Automatic account protection after failed attempts

### Multi-tenancy
- 🏢 **Tenant Isolation**: Complete data isolation per tenant
- 🔑 **Tenant-specific Roles**: Flexible role assignment per tenant
- ⚙️ **Tenant Configuration**: Custom auth policies per tenant

## Prerequisites

- Go 1.25+
- MongoDB 4.4+ (for user data, roles, and tokens storage)
- Redis 6.0+ (for session and cache management)
- Docker & Docker Compose (optional, for containerized deployment)

## Installation

```bash
# Clone the repository
git clone https://github.com/vhvcorp/go-auth-service.git
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

### REST API Endpoints

#### Authentication Endpoints

**POST `/api/v1/auth/register`**
```bash
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "tenant_id": "tenant_123"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 3600,
  "token_type": "Bearer",
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "email": "user@example.com",
    "tenant_id": "tenant_123",
    "roles": ["user"]
  }
}
```

**POST `/api/v1/auth/login`**
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "tenant_id": "tenant_123"
  }'
```

**POST `/api/v1/auth/refresh`**
```bash
curl -X POST http://localhost:8081/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGc..."
  }'
```

**POST `/api/v1/auth/logout`**
```bash
curl -X POST http://localhost:8081/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGc..."
  }'
```

#### Health Check Endpoints

**GET `/health`**
```bash
curl http://localhost:8081/health
```

**GET `/ready`**
```bash
curl http://localhost:8081/ready
```

### gRPC API

The service exposes gRPC endpoints on port 50051:

- `AuthService.Register`
- `AuthService.Login`
- `AuthService.RefreshToken`
- `AuthService.ValidateToken`
- `AuthService.Logout`
- `AuthService.GetUserRoles`
- `AuthService.CheckPermission`

See `proto/auth.proto` for complete API definitions.

## Deployment

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for deployment instructions.

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for architecture details.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## Related Repositories

- [go-shared](https://github.com/vhvcorp/go-shared) - Shared Go libraries
- [saas-framework-go](https://github.com/vhvcorp/saas-framework-go) - Original monorepo

## License

MIT License - see [LICENSE](LICENSE) for details

## Support

- Documentation: [Wiki](https://github.com/vhvcorp/go-auth-service/wiki)
- Issues: [GitHub Issues](https://github.com/vhvcorp/go-auth-service/issues)
- Discussions: [GitHub Discussions](https://github.com/vhvcorp/go-auth-service/discussions)
