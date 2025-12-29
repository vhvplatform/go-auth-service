@echo off
REM Windows Test Script for go-auth-service
REM This script runs all tests on Windows

echo ========================================
echo   Running go-auth-service Tests
echo ========================================
echo.

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH
    echo Please run setup-windows.bat first
    exit /b 1
)

echo [INFO] Go version:
go version
echo.

REM Run tests
echo [1/3] Running all tests...
echo ----------------------------------------
go test -v ./...
if %errorlevel% neq 0 (
    echo.
    echo [ERROR] Some tests failed
    pause
    exit /b 1
)
echo.

REM Run tests with race detector
echo [2/3] Running tests with race detector...
echo ----------------------------------------
go test -race ./...
if %errorlevel% neq 0 (
    echo.
    echo [WARNING] Race conditions detected
    pause
)
echo.

REM Run tests with coverage
echo [3/3] Generating coverage report...
echo ----------------------------------------
go test -coverprofile=coverage.txt -covermode=atomic ./...
if %errorlevel% equ 0 (
    go tool cover -html=coverage.txt -o coverage.html
    echo [OK] Coverage report generated: coverage.html
    echo [INFO] Opening coverage report in browser...
    start coverage.html
) else (
    echo [WARNING] Could not generate coverage report
)
echo.

echo ========================================
echo   All Tests Completed Successfully!
echo ========================================
echo.

pause
