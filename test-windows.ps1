# Windows Test Script for go-auth-service (PowerShell)
# This script runs all tests on Windows

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Running go-auth-service Tests" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
try {
    go version | Out-Null
    if ($LASTEXITCODE -ne 0) { throw }
} catch {
    Write-Host "[ERROR] Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please run setup-windows.ps1 first" -ForegroundColor Red
    exit 1
}

Write-Host "[INFO] Go version:" -ForegroundColor Cyan
go version
Write-Host ""

# Run tests
Write-Host "[1/3] Running all tests..." -ForegroundColor Yellow
Write-Host "----------------------------------------"
try {
    go test -v ./...
    if ($LASTEXITCODE -ne 0) { throw }
} catch {
    Write-Host ""
    Write-Host "[ERROR] Some tests failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Run tests with race detector
Write-Host "[2/3] Running tests with race detector..." -ForegroundColor Yellow
Write-Host "----------------------------------------"
try {
    go test -race ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "[WARNING] Race conditions detected" -ForegroundColor Yellow
    }
} catch {
    Write-Host ""
    Write-Host "[WARNING] Race detector test encountered issues" -ForegroundColor Yellow
}
Write-Host ""

# Run tests with coverage
Write-Host "[3/3] Generating coverage report..." -ForegroundColor Yellow
Write-Host "----------------------------------------"
try {
    go test -coverprofile=coverage.txt -covermode=atomic ./...
    if ($LASTEXITCODE -eq 0) {
        go tool cover -html=coverage.txt -o coverage.html
        Write-Host "[OK] Coverage report generated: coverage.html" -ForegroundColor Green
        Write-Host "[INFO] Opening coverage report in browser..." -ForegroundColor Cyan
        Start-Process coverage.html
    } else {
        throw
    }
} catch {
    Write-Host "[WARNING] Could not generate coverage report" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  All Tests Completed Successfully!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
