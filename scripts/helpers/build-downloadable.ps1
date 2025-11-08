#!/usr/bin/env pwsh
# Build script for downloadable version with embedded binaries
# This script compiles binaries for Windows and Linux/Mac, then builds
# the main application with the downloadable tag to embed them

Write-Host "🔨 Building Pejelagarto Translator - Downloadable Version" -ForegroundColor Cyan
Write-Host ""

# Ensure bin directory exists
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build Windows binary
Write-Host "📦 Building Windows binary..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o bin/pejelagarto-translator.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Windows build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Windows binary created: bin/pejelagarto-translator.exe" -ForegroundColor Green

# Build Linux/Mac binary
Write-Host "📦 Building Linux/Mac binary..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o bin/pejelagarto-translator .
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Linux/Mac build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Linux/Mac binary created: bin/pejelagarto-translator" -ForegroundColor Green

# Reset environment variables
Remove-Item Env:\GOOS
Remove-Item Env:\GOARCH

# Build downloadable version with embedded binaries
Write-Host ""
Write-Host "📦 Building downloadable version with embedded binaries..." -ForegroundColor Yellow
go build -tags downloadable -o bin/pejelagarto-translator-downloadable.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Downloadable build failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✅ Downloadable build complete!" -ForegroundColor Green
Write-Host "📁 Output: bin/pejelagarto-translator-downloadable.exe" -ForegroundColor Cyan
Write-Host ""
Write-Host "To run the downloadable version:" -ForegroundColor Yellow
Write-Host "  .\bin\pejelagarto-translator-downloadable.exe" -ForegroundColor White
Write-Host ""
Write-Host "The download buttons will appear at the bottom of the web UI." -ForegroundColor Yellow
