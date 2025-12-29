@echo off
REM Windows Build Script for go-auth-service
REM This script builds the application for Windows

echo ========================================
echo   Building go-auth-service for Windows
echo ========================================
echo.

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH
    exit /b 1
)

REM Create bin directory if it doesn't exist
if not exist bin mkdir bin

REM Clean previous builds
echo [1/3] Cleaning previous builds...
if exist bin\auth-service.exe del bin\auth-service.exe
echo [OK] Clean complete
echo.

REM Build the application
echo [2/3] Building application...
echo Target: Windows AMD64
echo Output: bin\auth-service.exe
echo.

set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

go build -ldflags="-s -w" -o bin\auth-service.exe .\cmd\main.go

if %errorlevel% neq 0 (
    echo [ERROR] Build failed
    pause
    exit /b 1
)

echo [OK] Build successful
echo.

REM Show build info
echo [3/3] Build information:
for %%I in (bin\auth-service.exe) do echo Size: %%~zI bytes
echo Location: %CD%\bin\auth-service.exe
echo.

echo ========================================
echo   Build Complete!
echo ========================================
echo.
echo To run the service:
echo   .\bin\auth-service.exe
echo.
echo Or use the development script:
echo   run-dev.bat
echo.

pause
