# go-auth-service

> Part of the SaaS Framework - Extracted from monorepo

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

- [go-shared](https://github.com/vhvcorp/go-shared) - Shared Go libraries
- [saas-framework-go](https://github.com/vhvcorp/saas-framework-go) - Original monorepo

## License

MIT License - see [LICENSE](LICENSE) for details

## Support

- Documentation: [Wiki](https://github.com/vhvcorp/go-auth-service/wiki)
- Issues: [GitHub Issues](https://github.com/vhvcorp/go-auth-service/issues)
- Discussions: [GitHub Discussions](https://github.com/vhvcorp/go-auth-service/discussions)
