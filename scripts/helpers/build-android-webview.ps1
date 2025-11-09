#!/usr/bin/env pwsh
# Build Android APK with gomobile bind + Gradle
# This creates a proper WebView-based Android app that calls Go translation functions

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Android APK Build (WebView + Gradle)" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Get paths
$rootDir = Split-Path -Parent (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
$androidDir = Join-Path $rootDir "android"
$libsDir = Join-Path $androidDir "app\libs"
$pkgDir = Join-Path $rootDir "pkg\translator"

# Ensure libs directory exists
if (-not (Test-Path $libsDir)) {
    New-Item -ItemType Directory -Path $libsDir -Force | Out-Null
}

Write-Host "[1/5] Checking gomobile installation..." -ForegroundColor Yellow
if (-not (Get-Command "gomobile" -ErrorAction SilentlyContinue)) {
    Write-Host "  Installing gomobile..." -ForegroundColor Gray
    go install golang.org/x/mobile/cmd/gomobile@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ Failed to install gomobile" -ForegroundColor Red
        exit 1
    }
}
Write-Host "  ✓ gomobile found" -ForegroundColor Green

Write-Host ""
Write-Host "[2/5] Checking Java JDK..." -ForegroundColor Yellow
$javaHome = $env:JAVA_HOME
if (-not $javaHome -or -not (Test-Path "$javaHome\bin\java.exe")) {
    Write-Host "  Detecting Java installation..." -ForegroundColor Gray
    
    # Try to find Java via winget
    $javaPath = & where.exe java 2>$null | Select-Object -First 1
    if ($javaPath) {
        $javaHome = Split-Path -Parent (Split-Path -Parent $javaPath)
    } else {
        Write-Host "  ✗ Java JDK not found. Please install Java JDK 17 or later." -ForegroundColor Red
        Write-Host "  You can install it with: winget install Microsoft.OpenJDK.17" -ForegroundColor Yellow
        exit 1
    }
}

$env:JAVA_HOME = $javaHome
$javaVersion = & "$javaHome\bin\java.exe" -version 2>&1 | Select-Object -First 1
Write-Host "  ✓ Java JDK found: $javaVersion" -ForegroundColor Green

Write-Host ""
Write-Host "[3/5] Checking Android SDK..." -ForegroundColor Yellow
$androidHome = $env:ANDROID_HOME
if (-not $androidHome) {
    $androidHome = "$env:LOCALAPPDATA\Android\sdk"
}
if (-not (Test-Path $androidHome)) {
    Write-Host "  ✗ Android SDK not found at: $androidHome" -ForegroundColor Red
    Write-Host "  Please run build-android-apk.ps1 first to set up Android SDK" -ForegroundColor Yellow
    exit 1
}
$env:ANDROID_HOME = $androidHome
Write-Host "  ✓ Android SDK found at: $androidHome" -ForegroundColor Green

Write-Host ""
Write-Host "[4/5] Building Go library with gomobile bind..." -ForegroundColor Yellow
Write-Host "  Package: pkg/translator" -ForegroundColor Gray
Write-Host "  Output: translator.aar" -ForegroundColor Gray
Write-Host "  Features: WebView with JavaScript bridge to Go functions" -ForegroundColor Gray
Write-Host "  This may take a few minutes..." -ForegroundColor Gray

Push-Location $pkgDir
try {
    # Set Android NDK home to use newer NDK version
    $ndkPath = Join-Path $androidHome "ndk\26.3.11579264"
    if (Test-Path $ndkPath) {
        $env:ANDROID_NDK_HOME = $ndkPath
        Write-Host "  Using NDK: $ndkPath" -ForegroundColor Gray
    }
    
    # Initialize gomobile if needed
    gomobile init 2>$null
    
    # Build AAR library for Android with API 24
    $env:ANDROID_API = "24"
    gomobile bind -target=android -androidapi=24 -o "$libsDir\translator.aar" .
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ Failed to build Go library" -ForegroundColor Red
        exit 1
    }
    
    if (-not (Test-Path "$libsDir\translator.aar")) {
        Write-Host "  ✗ AAR file not created" -ForegroundColor Red
        exit 1
    }
    
    $aarSize = [math]::Round((Get-Item "$libsDir\translator.aar").Length / 1MB, 2)
    Write-Host "  ✓ Go library built successfully ($aarSize MB)" -ForegroundColor Green
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "[5/5] Building APK with Gradle..." -ForegroundColor Yellow
Write-Host "  Building debug APK (WebView + JavaScript bridge)..." -ForegroundColor Gray
Write-Host "  This may take several minutes on first build..." -ForegroundColor Gray

Push-Location $androidDir
try {
    # Build debug APK (easier for testing and already signed)
    .\gradlew.bat assembleDebug
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ Gradle build failed" -ForegroundColor Red
        exit 1
    }
    
    $apkPath = "app\build\outputs\apk\debug\app-debug.apk"
    
    if (-not (Test-Path $apkPath)) {
        Write-Host "  ✗ APK file not found at: $apkPath" -ForegroundColor Red
        exit 1
    }
    
    # Copy to bin directory
    $binDir = Join-Path $rootDir "bin"
    if (-not (Test-Path $binDir)) {
        New-Item -ItemType Directory -Path $binDir -Force | Out-Null
    }
    Copy-Item $apkPath "$binDir\pejelagarto-translator-webview.apk" -Force
    
    $apkSize = [math]::Round((Get-Item $apkPath).Length / 1MB, 2)
    Write-Host "  ✓ APK built successfully ($apkSize MB)" -ForegroundColor Green
    
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "======================================" -ForegroundColor Green
Write-Host "  APK Build Complete!" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Green
Write-Host ""
Write-Host "APK Location: bin\pejelagarto-translator-webview.apk" -ForegroundColor Cyan
Write-Host "Type: WebView-based native Android app" -ForegroundColor Cyan
Write-Host "Features: Full UI with Go translation backend" -ForegroundColor Cyan
Write-Host ""
Write-Host "To install on device:" -ForegroundColor Yellow
Write-Host "  adb install bin\pejelagarto-translator-webview.apk" -ForegroundColor White
Write-Host ""
