param(
    [string]$OS = "windows"
)

# Determine output filename based on OS
$outputFile = if ($OS -eq "windows") {
    "bin/piper-server.exe"
} else {
    "bin/piper-server"
}

Write-Host "🔨 Building Obfuscated Version with Embedded Binaries" -ForegroundColor Cyan
Write-Host "Building obfuscated version for $OS..."
Write-Host "Output: $outputFile"

# Ensure bin directory exists
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build required embedded binaries first
Write-Host ""
Write-Host "📦 Building embedded binaries..." -ForegroundColor Yellow

# Build Windows binary
Write-Host "  Building Windows binary..." -ForegroundColor White
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o bin/pejelagarto-translator.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Windows binary build failed!" -ForegroundColor Red
    exit 1
}

# Build Linux binary
Write-Host "  Building Linux binary..." -ForegroundColor White
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o bin/pejelagarto-translator .
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Linux binary build failed!" -ForegroundColor Red
    exit 1
}

# Reset environment variables
Remove-Item Env:\GOOS
Remove-Item Env:\GOARCH

# Build Android APK if possible
Write-Host "  Building Android APK..." -ForegroundColor White
if (Test-Path ".\scripts\helpers\build-android-apk.ps1") {
    try {
        & ".\scripts\helpers\build-android-apk.ps1" 2>&1 | Out-Null
    } catch {
        Write-Host "⚠️  Android APK build failed, continuing..." -ForegroundColor Yellow
    }
}

# Build Android WebView APK if possible  
Write-Host "  Building Android WebView APK..." -ForegroundColor White
if (Test-Path ".\scripts\helpers\build-android-webview.ps1") {
    try {
        & ".\scripts\helpers\build-android-webview.ps1" 2>&1 | Out-Null
    } catch {
        Write-Host "⚠️  Android WebView APK build failed, continuing..." -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "📦 Building obfuscated server with embedded binaries..." -ForegroundColor Yellow

# Run garble with obfuscation flags
# Build the entire package (.) instead of main.go to properly handle build tags
garble -literals -tiny build -tags "obfuscated,downloadable" -o $outputFile .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Obfuscated build complete!" -ForegroundColor Green
    Write-Host "📁 Output: $outputFile" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Features included:" -ForegroundColor Yellow
    Write-Host "  ✓ Code obfuscation (garble)" -ForegroundColor Green
    Write-Host "  ✓ Embedded Windows/Linux binaries" -ForegroundColor Green
    Write-Host "  ✓ Embedded Android APKs (if built)" -ForegroundColor Green
    Write-Host "  ✓ Download buttons in web UI" -ForegroundColor Green
} else {
    Write-Host "❌ Build failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}
