# Migration Guide - Repository Restructure

## Overview

The repository has been restructured to support a full-stack platform architecture with separate directories for backend, frontend, and mobile applications.

## What Changed?

### Before (Old Structure)
```
go-auth-service/
├── cmd/
├── internal/
├── proto/
├── docs/
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── ...
```

### After (New Structure)
```
vhv-platform-auth/
├── server/          # Go backend (all previous code moved here)
│   ├── cmd/
│   ├── internal/
│   ├── proto/
│   ├── go.mod
│   ├── go.sum
│   ├── Makefile
│   └── Dockerfile
├── client/          # ReactJS frontend (placeholder)
├── flutter/         # Flutter mobile app (placeholder)
├── docs/            # Project documentation
├── Makefile         # Root Makefile for all components
└── Dockerfile       # Root Dockerfile
```

## For Developers

### Checking Out the New Structure

If you already have the repository cloned:

```bash
# Switch to the new structure branch
git fetch origin
git checkout copilot/restructure-repository-structure
```

If you're cloning fresh:

```bash
# Clone the repository
git clone https://github.com/vhvplatform/go-auth-service.git
cd go-auth-service

# Checkout the restructure branch
git checkout copilot/restructure-repository-structure
```

### Working with the Backend (Server)

All Go backend work now happens in the `server/` directory:

```bash
# Navigate to server directory
cd server

# Install dependencies
go mod download

# Build the project
make build

# Run tests
make test

# Run the service
make run
```

### IDE Configuration Updates

#### VS Code

Update your `launch.json` and `tasks.json` to use the `server/` directory:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/server/cmd/main.go",
      "cwd": "${workspaceFolder}/server"
    }
  ]
}
```

#### GoLand / IntelliJ IDEA

1. Open the project
2. Right-click on `server/` directory
3. Select "Mark Directory as" → "Sources Root"
4. Update run configurations to use `server/cmd/main.go` as the entry point

### Docker Usage

#### Building from Server Directory

```bash
cd server
docker build -t vhv-auth-server:latest .
```

#### Building from Root Directory

```bash
# Uses the root Dockerfile which references server/
docker build -t vhv-auth-server:latest .
```

### Environment Files

Your `.env` file should now be in the `server/` directory:

```bash
cd server
cp .env.example .env
# Edit .env with your configuration
```

### Import Paths

Import paths **have NOT changed**. The Go module path remains:
```go
import "github.com/vhvplatform/go-auth-service/internal/..."
```

This is because `go.mod` in the `server/` directory still declares:
```
module github.com/vhvplatform/go-auth-service
```

### Makefile Commands

From repository root:
```bash
make server-build    # Build the server
make server-test     # Test the server
make server-run      # Run the server
make help           # Show all available commands
```

From server directory:
```bash
cd server
make build          # Build
make test           # Test
make run            # Run
```

## For CI/CD

### GitHub Actions

All workflow files have been updated to work with the new structure:
- `working-directory: ./server` added to relevant steps
- Cache keys updated to use `server/go.sum`
- Build paths updated to reference `server/`

### Existing Pipelines

If you have custom CI/CD pipelines, update them to:

1. **Navigate to server directory** before running Go commands:
   ```yaml
   - run: cd server && go build ./cmd/main.go
   ```

2. **Update go.sum paths** in cache configurations:
   ```yaml
   key: ${{ runner.os }}-go-${{ hashFiles('server/go.sum') }}
   ```

3. **Update test commands**:
   ```yaml
   - run: cd server && go test ./...
   ```

## For Deployment

### Kubernetes / Docker Compose

Update your deployment manifests to use the new Dockerfile:

```yaml
# Build context remains at repository root
build:
  context: .
  dockerfile: Dockerfile
```

Or build from the server directory:

```yaml
build:
  context: ./server
  dockerfile: Dockerfile
```

### Existing Containers

Rebuild your containers with the new structure:

```bash
docker build -t vhv-auth-server:latest .
docker tag vhv-auth-server:latest your-registry/vhv-auth-server:latest
docker push your-registry/vhv-auth-server:latest
```

## Future Work

### Client (ReactJS Frontend)

When the frontend is implemented:
```bash
cd client
npm install
npm start
```

### Flutter (Mobile App)

When the mobile app is implemented:
```bash
cd flutter
flutter pub get
flutter run
```

## Troubleshooting

### "Cannot find go.mod"

Make sure you're in the `server/` directory:
```bash
cd server
go mod download
```

### "Build failed"

1. Clean old build artifacts:
   ```bash
   cd server
   make clean
   ```

2. Re-download dependencies:
   ```bash
   go mod download
   go mod verify
   ```

3. Rebuild:
   ```bash
   make build
   ```

### IDE Not Recognizing Imports

1. Mark `server/` as the project root in your IDE
2. Restart the Go language server
3. Run `go mod download` in the server directory

## Questions?

If you encounter any issues with the migration:
1. Check this migration guide
2. Review the updated README.md in the repository root
3. Check component-specific README files in each directory
4. Open an issue on GitHub

## Git Commands Summary

```bash
# For existing clones - switch to new structure
git fetch origin
git checkout copilot/restructure-repository-structure

# For fresh clones
git clone https://github.com/vhvplatform/go-auth-service.git
cd go-auth-service
git checkout copilot/restructure-repository-structure
```
