#!/usr/bin/env pwsh
# Complete Android WebView APK Build Script
# Builds Go library with gomobile bind, then builds APK with Gradle

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Android WebView APK Builder" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Get paths
$rootDir = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$androidDir = Join-Path $rootDir "android"
$libsDir = Join-Path $androidDir "app\libs"
$binDir = Join-Path $rootDir "bin"

Write-Host "[1/6] Checking prerequisites..." -ForegroundColor Yellow

# Check gomobile
if (-not (Get-Command "gomobile" -ErrorAction SilentlyContinue)) {
    Write-Host "  Installing gomobile..." -ForegroundColor Gray
    go install golang.org/x/mobile/cmd/gomobile@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ Failed to install gomobile" -ForegroundColor Red
        exit 1
    }
}
Write-Host "  ✓ gomobile found" -ForegroundColor Green

# Check Java
$javaHome = $env:JAVA_HOME
if (-not $javaHome -or -not (Test-Path "$javaHome\bin\java.exe")) {
    $javaPath = & where.exe java 2>$null | Select-Object -First 1
    if ($javaPath) {
        $javaHome = Split-Path -Parent (Split-Path -Parent $javaPath)
        $env:JAVA_HOME = $javaHome
    } else {
        Write-Host "  ✗ Java JDK not found" -ForegroundColor Red
        Write-Host "  Install with: winget install Microsoft.OpenJDK.17" -ForegroundColor Yellow
        exit 1
    }
}
$javaVersion = & "$javaHome\bin\java.exe" -version 2>&1 | Select-Object -First 1
Write-Host "  ✓ Java: $javaVersion" -ForegroundColor Green

# Check Android SDK
$androidHome = $env:ANDROID_HOME
if (-not $androidHome) {
    $androidHome = "$env:LOCALAPPDATA\Android\sdk"
}
if (-not (Test-Path $androidHome)) {
    Write-Host "  ✗ Android SDK not found" -ForegroundColor Red
    Write-Host "  Run build-android-apk.ps1 first to set up SDK" -ForegroundColor Yellow
    exit 1
}
$env:ANDROID_HOME = $androidHome
Write-Host "  ✓ Android SDK: $androidHome" -ForegroundColor Green

# Check Android NDK
$ndkPath = Join-Path $androidHome "ndk\26.3.11579264"
if (-not (Test-Path $ndkPath)) {
    Write-Host "  ✗ Android NDK not found" -ForegroundColor Red
    Write-Host "  Run build-android-apk.ps1 first to install NDK" -ForegroundColor Yellow
    exit 1
}
$env:ANDROID_NDK_HOME = $ndkPath
Write-Host "  ✓ Android NDK: $ndkPath" -ForegroundColor Green

Write-Host ""
Write-Host "[2/6] Building Go library with gomobile bind..." -ForegroundColor Yellow
Write-Host "  This creates an Android AAR library from Go code" -ForegroundColor Gray
Write-Host "  Package: pejelagarto-translator/pkg/translator" -ForegroundColor Gray

Push-Location $rootDir
try {
    # Ensure libs directory exists
    if (-not (Test-Path $libsDir)) {
        New-Item -ItemType Directory -Path $libsDir -Force | Out-Null
    }
    
    # Initialize gomobile
    gomobile init 2>$null
    
    # Build AAR library
    $env:ANDROID_API = "24"
    gomobile bind -target=android -androidapi=24 -o "$libsDir\translator.aar" pejelagarto-translator/pkg/translator
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ Failed to build Go library" -ForegroundColor Red
        exit 1
    }
    
    if (-not (Test-Path "$libsDir\translator.aar")) {
        Write-Host "  ✗ AAR file not created" -ForegroundColor Red
        exit 1
    }
    
    $aarSize = [math]::Round((Get-Item "$libsDir\translator.aar").Length / 1MB, 2)
    Write-Host "  ✓ Go library built: translator.aar ($aarSize MB)" -ForegroundColor Green
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "[3/6] Verifying Android project structure..." -ForegroundColor Yellow

$requiredFiles = @(
    "$androidDir\build.gradle",
    "$androidDir\settings.gradle",
    "$androidDir\app\build.gradle",
    "$androidDir\app\src\main\AndroidManifest.xml",
    "$androidDir\app\src\main\java\com\pejelagarto\translator\MainActivity.java",
    "$androidDir\gradle\wrapper\gradle-wrapper.jar",
    "$androidDir\gradle\wrapper\gradle-wrapper.properties",
    "$androidDir\gradlew.bat"
)

$missingFiles = @()
foreach ($file in $requiredFiles) {
    if (-not (Test-Path $file)) {
        $missingFiles += $file
    }
}

if ($missingFiles.Count -gt 0) {
    Write-Host "  ✗ Missing required files:" -ForegroundColor Red
    foreach ($file in $missingFiles) {
        Write-Host "    - $file" -ForegroundColor Gray
    }
    exit 1
}

Write-Host "  ✓ All required files present" -ForegroundColor Green

Write-Host ""
Write-Host "[4/6] Cleaning previous builds..." -ForegroundColor Yellow
Push-Location $androidDir
try {
    if (Test-Path "app\build") {
        Remove-Item -Recurse -Force "app\build" -ErrorAction SilentlyContinue
        Write-Host "  ✓ Cleaned build directory" -ForegroundColor Green
    } else {
        Write-Host "  ✓ No previous build to clean" -ForegroundColor Green
    }
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "[5/6] Building APK with Gradle..." -ForegroundColor Yellow
Write-Host "  This will download Gradle and dependencies on first run" -ForegroundColor Gray
Write-Host "  Build type: Debug (unsigned)" -ForegroundColor Gray
Write-Host "  This may take 5-10 minutes..." -ForegroundColor Gray

Push-Location $androidDir
try {
    # Build debug APK (easier for testing)
    & .\gradlew.bat assembleDebug
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ Gradle build failed" -ForegroundColor Red
        Write-Host "  Check the error messages above for details" -ForegroundColor Yellow
        exit 1
    }
    
    # Find the APK
    $apkPath = "app\build\outputs\apk\debug\app-debug.apk"
    
    if (-not (Test-Path $apkPath)) {
        Write-Host "  ✗ APK file not found at: $apkPath" -ForegroundColor Red
        exit 1
    }
    
    $apkSize = [math]::Round((Get-Item $apkPath).Length / 1MB, 2)
    Write-Host "  ✓ APK built successfully ($apkSize MB)" -ForegroundColor Green
    
    Write-Host ""
    Write-Host "[6/6] Copying APK to bin directory..." -ForegroundColor Yellow
    
    # Ensure bin directory exists
    if (-not (Test-Path $binDir)) {
        New-Item -ItemType Directory -Path $binDir -Force | Out-Null
    }
    
    # Copy to bin with a clear name
    $outputApk = Join-Path $binDir "pejelagarto-translator-webview.apk"
    Copy-Item $apkPath $outputApk -Force
    
    Write-Host "  ✓ APK copied to: $outputApk" -ForegroundColor Green
    
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "======================================" -ForegroundColor Green
Write-Host "  ✓ Build Complete!" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Green
Write-Host ""
Write-Host "📱 APK Details:" -ForegroundColor Cyan
Write-Host "  Location: bin\pejelagarto-translator-webview.apk" -ForegroundColor White
Write-Host "  Type: Native Android app with WebView UI" -ForegroundColor White
Write-Host "  Features: Full translation with embedded Go library" -ForegroundColor White
Write-Host "  Size: $apkSize MB" -ForegroundColor White
Write-Host ""
Write-Host "📲 Install on device:" -ForegroundColor Yellow
Write-Host "  adb install -r bin\pejelagarto-translator-webview.apk" -ForegroundColor White
Write-Host ""
Write-Host "🎉 The app will have a proper UI with native Android WebView!" -ForegroundColor Green
Write-Host ""
