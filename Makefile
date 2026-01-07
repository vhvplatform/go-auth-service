.PHONY: help server-build server-test server-run client-build client-run flutter-build docker-build-all clean

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Server (Go Backend) targets
server-build: ## Build the Go backend server
	@echo "Building Go backend server..."
	@cd server && $(MAKE) build

server-test: ## Run Go backend tests
	@echo "Running Go backend tests..."
	@cd server && $(MAKE) test

server-run: ## Run the Go backend server locally
	@echo "Running Go backend server..."
	@cd server && $(MAKE) run

server-lint: ## Lint Go backend code
	@echo "Linting Go backend..."
	@cd server && $(MAKE) lint

server-clean: ## Clean Go backend build artifacts
	@echo "Cleaning Go backend..."
	@cd server && $(MAKE) clean

# Client (ReactJS Frontend) targets
client-install: ## Install client dependencies (placeholder)
	@echo "Client installation not yet implemented"
	@echo "Future: cd client && npm install"

client-build: ## Build the ReactJS frontend (placeholder)
	@echo "Client build not yet implemented"
	@echo "Future: cd client && npm run build"

client-run: ## Run the ReactJS frontend (placeholder)
	@echo "Client run not yet implemented"
	@echo "Future: cd client && npm start"

# Flutter (Mobile App) targets
flutter-get: ## Get Flutter dependencies (placeholder)
	@echo "Flutter setup not yet implemented"
	@echo "Future: cd flutter && flutter pub get"

flutter-build: ## Build the Flutter mobile app (placeholder)
	@echo "Flutter build not yet implemented"
	@echo "Future: cd flutter && flutter build apk"

flutter-run: ## Run the Flutter mobile app (placeholder)
	@echo "Flutter run not yet implemented"
	@echo "Future: cd flutter && flutter run"

# Docker targets
docker-build-server: ## Build Docker image for server
	@echo "Building Docker image for server..."
	@docker build -t vhv-auth-server:latest -f Dockerfile .

docker-run-server: ## Run server Docker container
	@echo "Running server Docker container..."
	@docker run --rm -p 50051:50051 -p 8081:8081 --name vhv-auth-server vhv-auth-server:latest

# Combined targets
build-all: server-build ## Build all components
	@echo "All components built successfully!"

test-all: server-test ## Run all tests
	@echo "All tests completed!"

clean: server-clean ## Clean all build artifacts
	@echo "All cleaned!"

# Installation targets
install: ## Install all dependencies
	@echo "Installing server dependencies..."
	@cd server && go mod download
	@echo "All dependencies installed!"

# Default target
.DEFAULT_GOAL := help
