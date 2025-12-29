@echo off
REM Windows Development Run Script for go-auth-service
REM This script runs the auth service in development mode

echo ========================================
echo   Starting go-auth-service
echo ========================================
echo.

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH
    echo Please run setup-windows.bat first
    exit /b 1
)

REM Check if .env file exists
if not exist .env (
    echo [WARNING] .env file not found
    echo Please run setup-windows.bat first or create .env file
    echo.
    pause
    exit /b 1
)

REM Display configuration
echo [INFO] Starting service with configuration from .env
echo [INFO] HTTP Port: Check AUTH_SERVICE_HTTP_PORT in .env (default: 8081)
echo [INFO] gRPC Port: Check AUTH_SERVICE_PORT in .env (default: 50051)
echo.
echo [INFO] Health check: http://localhost:8081/health
echo [INFO] Press Ctrl+C to stop the service
echo.
echo ----------------------------------------
echo.

REM Run the application
go run cmd\main.go

if %errorlevel% neq 0 (
    echo.
    echo [ERROR] Service failed to start
    echo.
    echo Common issues:
    echo - MongoDB is not running
    echo - Redis is not running
    echo - Port is already in use
    echo - Invalid configuration in .env
    echo.
    echo See WINDOWS_SETUP.md for troubleshooting
    pause
    exit /b 1
)
