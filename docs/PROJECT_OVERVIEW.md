# VHV Platform - Project Overview

## Repository Structure

This repository contains the complete VHV Platform Authentication Service, organized into multiple microservices and applications:

```
vhv-platform-auth/
├── server/          # Go backend microservice
│   ├── cmd/        # Application entry point
│   ├── internal/   # Private application code
│   ├── proto/      # Protocol Buffer definitions
│   └── ...
├── client/          # ReactJS frontend (future)
├── flutter/         # Flutter mobile app (future)
├── docs/            # Project documentation
└── README.md        # Main project documentation
```

## Components

### 1. Server (Backend Microservice)

**Technology:** Go 1.25.5+  
**Location:** `/server`  
**Status:** ✅ Active Development

The authentication backend service provides:
- gRPC API for service-to-service communication
- REST API for client applications
- JWT token generation and validation
- OAuth2 integration (Google, GitHub, etc.)
- User management and authentication
- Session management with Redis
- MongoDB for data persistence

**Key Features:**
- Multi-tenant architecture
- Role-based access control (RBAC)
- Token refresh mechanism
- Password security (bcrypt hashing)
- Rate limiting and brute force protection
- MFA (Multi-Factor Authentication) support

**APIs:**
- gRPC port: 50051
- HTTP/REST port: 8081

### 2. Client (Web Frontend)

**Technology:** ReactJS (planned)  
**Location:** `/client`  
**Status:** 🚧 Planned

The web frontend will provide:
- Modern, responsive user interface
- Login and registration forms
- OAuth2 social login integration
- User profile management
- Admin dashboard
- Real-time session management

**Planned Features:**
- TypeScript for type safety
- Redux Toolkit for state management
- Material-UI or Ant Design for UI components
- Axios for API communication
- React Router for navigation
- Jest and React Testing Library for testing

### 3. Flutter (Mobile Application)

**Technology:** Flutter (planned)  
**Location:** `/flutter`  
**Status:** 🚧 Planned

The mobile application will provide:
- Native iOS and Android experience
- Biometric authentication
- Push notifications
- Offline mode with local caching
- Deep linking support
- Face ID / Touch ID integration

**Planned Features:**
- Dart programming language
- Provider or Riverpod for state management
- Dio for HTTP requests
- flutter_secure_storage for secure token storage
- Firebase Cloud Messaging for notifications
- SQLite for local data storage

## Architecture

### Communication Flow

```
Mobile App (Flutter) ──┐
                       │
Web App (React) ───────┼──> API Gateway ──> Backend (Go) ──> MongoDB
                       │                                  └──> Redis
Other Services ────────┘
```

### Authentication Flow

1. **Registration:**
   - User submits registration data
   - Backend validates and creates user account
   - Email verification sent (optional)
   - JWT tokens generated and returned

2. **Login:**
   - User submits credentials
   - Backend validates credentials
   - JWT access token and refresh token generated
   - Session stored in Redis

3. **Token Refresh:**
   - Client sends refresh token
   - Backend validates and generates new tokens
   - Old refresh token revoked

4. **OAuth2 Flow:**
   - User initiates OAuth login
   - Redirects to provider (Google, GitHub, etc.)
   - User authorizes application
   - Backend exchanges code for tokens
   - User profile fetched and account created/linked
   - Internal JWT tokens generated

## Development Workflow

### Server Development

```bash
cd server
go mod download
cp .env.example .env
# Edit .env with your configuration
make run
```

### Testing

```bash
cd server
make test
make test-coverage
```

### Building for Production

```bash
# Server
cd server
make build-linux

# Docker
docker build -t vhv-auth-server:latest -f Dockerfile .
```

## Documentation

### Available Documentation

- **[DATABASE_DESIGN.md](DATABASE_DESIGN.md)** - Database schema and design decisions
- **[DEPENDENCIES.md](DEPENDENCIES.md)** - API documentation and dependencies
- **[OAUTH2_INTEGRATION.md](OAUTH2_INTEGRATION.md)** - OAuth2 setup and configuration
- **[PERFORMANCE.md](PERFORMANCE.md)** - Performance optimization guidelines
- **[SECURITY_HARDENING.md](SECURITY_HARDENING.md)** - Security best practices
- **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - Common issues and solutions

### Root Documentation

- **[README.md](../README.md)** - Main project overview
- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - Contribution guidelines
- **[SECURITY.md](../SECURITY.md)** - Security policies
- **[CHANGELOG.md](../CHANGELOG.md)** - Version history
- **[WINDOWS_SETUP.md](../WINDOWS_SETUP.md)** - Windows development setup

## Technology Stack Summary

| Component | Technology | Status |
|-----------|------------|--------|
| Backend | Go 1.25.5 | Active |
| Frontend | ReactJS | Planned |
| Mobile | Flutter | Planned |
| Database | MongoDB | Active |
| Cache | Redis | Active |
| API | gRPC + REST | Active |
| Auth | JWT + OAuth2 | Active |

## Getting Started

For detailed setup instructions for each component:

1. **Backend Server:** See [/server/README.md](../server/README.md)
2. **Frontend Client:** See [/client/README.md](../client/README.md)
3. **Mobile App:** See [/flutter/README.md](../flutter/README.md)

## Deployment

### Server Deployment

The backend can be deployed using:
- Docker containers
- Kubernetes
- Traditional VM deployment
- Cloud platforms (AWS, GCP, Azure)

### Client Deployment

The React frontend can be deployed to:
- Static hosting (Netlify, Vercel)
- CDN (CloudFront, Cloudflare)
- Traditional web servers (Nginx, Apache)

### Mobile Deployment

The Flutter app will be distributed through:
- Apple App Store (iOS)
- Google Play Store (Android)

## CI/CD

Continuous Integration and Deployment pipelines are configured in `.github/workflows/`:
- Automated testing on pull requests
- Docker image building and publishing
- Code quality checks and linting
- Security vulnerability scanning

## Support and Contribution

- **Issues:** [GitHub Issues](https://github.com/vhvplatform/go-auth-service/issues)
- **Discussions:** [GitHub Discussions](https://github.com/vhvplatform/go-auth-service/discussions)
- **Contributing:** See [CONTRIBUTING.md](../CONTRIBUTING.md)

## License

MIT License - See [LICENSE](../LICENSE) for details
