@echo off
REM Windows Setup Script for go-auth-service
REM This script helps set up the development environment on Windows

echo ========================================
echo   go-auth-service Windows Setup
echo ========================================
echo.

REM Check if Go is installed
echo [1/6] Checking Go installation...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH
    echo Please install Go from https://go.dev/dl/
    exit /b 1
)
go version
echo [OK] Go is installed
echo.

REM Check if Git is installed
echo [2/6] Checking Git installation...
git --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARNING] Git is not installed or not in PATH
    echo Please install Git from https://git-scm.com/download/win
) else (
    git --version
    echo [OK] Git is installed
)
echo.

REM Download dependencies
echo [3/6] Downloading Go dependencies...
go mod download
if %errorlevel% neq 0 (
    echo [ERROR] Failed to download dependencies
    exit /b 1
)
echo [OK] Dependencies downloaded
echo.

REM Verify dependencies
echo [4/6] Verifying Go dependencies...
go mod verify
if %errorlevel% neq 0 (
    echo [ERROR] Failed to verify dependencies
    exit /b 1
)
echo [OK] Dependencies verified
echo.

REM Create .env file if it doesn't exist
echo [5/6] Setting up environment configuration...
if not exist .env (
    if exist .env.example (
        copy .env.example .env
        echo [OK] Created .env from .env.example
        echo [ACTION REQUIRED] Please edit .env file with your configuration
    ) else (
        echo # Auth Service Configuration > .env
        echo MONGODB_URI=mongodb://localhost:27017 >> .env
        echo MONGODB_DATABASE=auth_service >> .env
        echo REDIS_HOST=localhost >> .env
        echo REDIS_PORT=6379 >> .env
        echo REDIS_PASSWORD= >> .env
        echo REDIS_DB=0 >> .env
        echo JWT_SECRET=your-secret-key-change-in-production >> .env
        echo JWT_EXPIRATION=900 >> .env
        echo JWT_REFRESH_EXPIRATION=604800 >> .env
        echo AUTH_SERVICE_PORT=50051 >> .env
        echo AUTH_SERVICE_HTTP_PORT=8081 >> .env
        echo LOG_LEVEL=info >> .env
        echo ENVIRONMENT=development >> .env
        echo [OK] Created default .env file
        echo [ACTION REQUIRED] Please edit .env file with your configuration
    )
) else (
    echo [OK] .env file already exists
)
echo.

REM Build the application
echo [6/6] Building the application...
if not exist bin mkdir bin
go build -o bin\auth-service.exe .\cmd\main.go
if %errorlevel% neq 0 (
    echo [ERROR] Failed to build application
    exit /b 1
)
echo [OK] Application built successfully: bin\auth-service.exe
echo.

echo ========================================
echo   Setup Complete!
echo ========================================
echo.
echo Next steps:
echo 1. Edit .env file with your configuration
echo 2. Ensure MongoDB is running (locally or in Docker)
echo 3. Ensure Redis is running (locally or in Docker)
echo 4. Run the service with: run-dev.bat
echo 5. Or run tests with: test-windows.bat
echo.
echo For more information, see WINDOWS_SETUP.md
echo.

pause
