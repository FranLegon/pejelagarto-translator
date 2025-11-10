#!/usr/bin/env pwsh
# Build script for ngrok_default version with embedded binaries and hardcoded ngrok config
# This script compiles binaries for Windows and Linux/Mac, then builds
# the main application with both downloadable and ngrok_default tags

Write-Host "🔨 Building Pejelagarto Translator - Ngrok Default Version" -ForegroundColor Cyan
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

# Build Android APK if possible
Write-Host "📦 Building Android APK..." -ForegroundColor Yellow
if (Test-Path ".\scripts\helpers\build-android-apk.ps1") {
    try {
        & ".\scripts\helpers\build-android-apk.ps1" 2>&1 | Out-Null
        if (Test-Path "bin\pejelagarto-translator.apk") {
            Write-Host "✅ Android APK created: bin/pejelagarto-translator.apk" -ForegroundColor Green
        }
    } catch {
        Write-Host "⚠️  Android APK build failed, continuing without it..." -ForegroundColor Yellow
    }
} else {
    Write-Host "⚠️  Android APK build script not found, skipping..." -ForegroundColor Yellow
}

# Build Android WebView APK if possible
Write-Host "📦 Building Android WebView APK..." -ForegroundColor Yellow
if (Test-Path ".\scripts\helpers\build-android-webview.ps1") {
    try {
        & ".\scripts\helpers\build-android-webview.ps1" 2>&1 | Out-Null
        if (Test-Path "bin\pejelagarto-translator-webview.apk") {
            Write-Host "✅ Android WebView APK created: bin/pejelagarto-translator-webview.apk" -ForegroundColor Green
        }
    } catch {
        Write-Host "⚠️  Android WebView APK build failed, continuing without it..." -ForegroundColor Yellow
    }
} else {
    Write-Host "⚠️  Android WebView APK build script not found, skipping..." -ForegroundColor Yellow
}

# Build ngrok_default version with embedded binaries and hardcoded ngrok
Write-Host ""
Write-Host "📦 Building ngrok_default version with embedded binaries..." -ForegroundColor Yellow
go build -tags "ngrok_default,downloadable" -o bin/pejelagarto-translator-ngrok.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Ngrok default build failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✅ Ngrok default build complete!" -ForegroundColor Green
Write-Host "📁 Output: bin/pejelagarto-translator-ngrok.exe" -ForegroundColor Cyan
Write-Host ""
Write-Host "To run the ngrok default version:" -ForegroundColor Yellow
Write-Host "  .\bin\pejelagarto-translator-ngrok.exe" -ForegroundColor White
Write-Host ""
Write-Host "This version includes:" -ForegroundColor Yellow
Write-Host "  • Hardcoded ngrok token and domain" -ForegroundColor White
Write-Host "  • Download buttons for embedded binaries" -ForegroundColor White
Write-Host "  • Embedded Windows/Linux binaries and Android APKs" -ForegroundColor White
Write-Host "  • No need to pass ngrok flags" -ForegroundColor White
