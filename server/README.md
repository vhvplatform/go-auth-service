# Server - Go Backend Microservice

This directory contains the Go backend microservice for the VHV Platform authentication system.

## Overview

The authentication service is built with Go and provides:
- gRPC API for service-to-service communication
- REST API for client applications
- JWT-based authentication
- MongoDB for user data storage
- Redis for session management
- OAuth2 integration support

## Structure

```
server/
├── cmd/              # Application entry point
├── internal/         # Private application code
│   ├── domain/      # Business logic and entities
│   ├── grpc/        # gRPC server implementation
│   ├── handler/     # HTTP handlers
│   ├── repository/  # Data access layer
│   ├── service/     # Business logic services
│   └── tests/       # Tests
├── proto/           # Protocol Buffer definitions
├── go.mod           # Go module dependencies
└── Makefile         # Build automation
```

## Prerequisites

- Go 1.25.5 or higher
- MongoDB
- Redis (optional, for caching)
- Protocol Buffers compiler (for gRPC)

## Getting Started

### Installation

```bash
cd server
go mod download
```

### Configuration

Copy the example environment file and update with your settings:

```bash
cp .env.example .env
```

### Running the Service

```bash
# Development mode
go run ./cmd/main.go

# Or using make
make run
```

### Building

```bash
# Build for current platform
make build

# Build for Linux
make build-linux

# Build for Windows
make build-windows
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run Windows-specific tests
make test-windows
```

### Docker

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run
```

## API Endpoints

### gRPC (Port 50051)
- Authentication services
- User management
- Token validation

### REST API (Port 8081)
- Health check: `GET /health`
- Additional HTTP endpoints

## Development

### Code Formatting
```bash
make fmt
```

### Linting
```bash
make lint
```

### Generate Protocol Buffers
```bash
make proto
```

## Documentation

For more detailed documentation, see the `/docs` directory in the repository root.

## Environment Variables

See `.env.example` for all available configuration options.

## Windows Development

For Windows-specific setup and development instructions, see:
- WINDOWS_SETUP.md (in repository root)
- WINDOWS_QUICKSTART.md (in repository root)
- WINDOWS_COMPATIBILITY.md (in repository root)
