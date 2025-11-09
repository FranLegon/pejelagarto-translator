#!/usr/bin/env pwsh
# Production build script (unobfuscated) with frontend + ngrok_default tags
# Uses standard Go compiler optimization instead of garble
# This version works reliably with ngrok (garble breaks ngrok SDK)
# Output: Optimized server with WASM frontend and hardcoded ngrok credentials

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("windows", "linux", "darwin")]
    [string]$OS = "windows",
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("amd64", "arm64")]
    [string]$Arch = "amd64"
)

$ErrorActionPreference = "Stop"

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Production Build (Unobfuscated)" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Build Configuration:" -ForegroundColor Yellow
Write-Host "  - Tags: obfuscated, frontend, ngrok_default, downloadable" -ForegroundColor White
Write-Host "  - OS: $OS" -ForegroundColor White
Write-Host "  - Architecture: $Arch" -ForegroundColor White
Write-Host "  - Optimization: Standard Go (-ldflags='-s -w')" -ForegroundColor White
Write-Host "  - ngrok: Compatible ✓" -ForegroundColor Green
Write-Host ""

# Build WASM first
Write-Host "[1/5] Building WASM module..." -ForegroundColor Green
$wasmEnv = @{
    GOOS = "js"
    GOARCH = "wasm"
}

$wasmOutput = "bin/main.wasm"
Write-Host "  Output: $wasmOutput" -ForegroundColor White

& {
    $env:GOOS = $wasmEnv.GOOS
    $env:GOARCH = $wasmEnv.GOARCH
    go build -tags "frontend" -o $wasmOutput .
}

if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build WASM" -ForegroundColor Red
    exit 1
}

$wasmSize = (Get-Item $wasmOutput).Length / 1MB
Write-Host "✓ WASM built successfully ($([math]::Round($wasmSize, 2)) MB)" -ForegroundColor Green
Write-Host ""

# Copy wasm_exec.js
Write-Host "[2/5] Copying wasm_exec.js..." -ForegroundColor Green
$goroot = go env GOROOT
$wasmExecSrc = Join-Path $goroot "lib\wasm\wasm_exec.js"
$wasmExecDest = "bin\wasm_exec.js"

if (Test-Path $wasmExecSrc) {
    Copy-Item $wasmExecSrc $wasmExecDest -Force
    Write-Host "✓ wasm_exec.js copied" -ForegroundColor Green
} else {
    Write-Host "WARNING: wasm_exec.js not found at $wasmExecSrc" -ForegroundColor Yellow
}
Write-Host ""

# Determine output filename
$outputName = "piper-server"
if ($OS -eq "windows") {
    $outputName += ".exe"
}
$outputPath = "bin\$outputName"

# Build server (frontend server for WASM mode)
Write-Host "[3/5] Building optimized frontend server..." -ForegroundColor Green

Write-Host "  Tags: frontendserver,obfuscated,ngrok_default,downloadable" -ForegroundColor White
Write-Host "  Flags: -ldflags='-s -w' (strip symbols)" -ForegroundColor White
Write-Host "  Output: $outputPath" -ForegroundColor White

& {
    $env:GOOS = $OS
    $env:GOARCH = $Arch
    $env:CGO_ENABLED = "0"
    
    go build `
        -ldflags="-s -w -extldflags '-static'" `
        -tags "frontendserver,obfuscated,ngrok_default,downloadable" `
        -trimpath `
        -o $outputPath `
        .
}

if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build server" -ForegroundColor Red
    exit 1
}

$serverSize = (Get-Item $outputPath).Length / 1MB
Write-Host "✓ Server built successfully ($([math]::Round($serverSize, 2)) MB)" -ForegroundColor Green
Write-Host ""

# Generate checksums
Write-Host "[4/5] Generating checksums..." -ForegroundColor Green
$checksumFile = "bin\checksums-prod.txt"
$serverHash = (Get-FileHash $outputPath -Algorithm SHA256).Hash
$wasmHash = (Get-FileHash $wasmOutput -Algorithm SHA256).Hash

$checksumContent = @"
Production Build Checksums (Unobfuscated)
Generated: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
Build: $OS/$Arch with obfuscated+frontend+ngrok_default+downloadable
Optimization: Standard Go (-ldflags='-s -w')

Server ($outputName):
  SHA256: $serverHash
  Size: $([math]::Round($serverSize, 2)) MB

WASM (main.wasm):
  SHA256: $wasmHash
  Size: $([math]::Round($wasmSize, 2)) MB
"@

$checksumContent | Out-File -FilePath $checksumFile -Encoding UTF8
Write-Host "✓ Checksums saved to $checksumFile" -ForegroundColor Green
Write-Host ""

# Summary
Write-Host "[5/5] Build Summary" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Build Type: Production (Unobfuscated)" -ForegroundColor White
Write-Host "Features:" -ForegroundColor White
Write-Host "  ✓ Standard Go optimization (-s -w)" -ForegroundColor Green
Write-Host "  ✓ WASM frontend" -ForegroundColor Green
Write-Host "  ✓ Hardcoded ngrok credentials" -ForegroundColor Green
Write-Host "  ✓ Embedded binaries (downloadable)" -ForegroundColor Green
Write-Host "  ✓ ngrok SDK compatible" -ForegroundColor Green
Write-Host "  ✓ Windows Defender friendly" -ForegroundColor Green
Write-Host ""
Write-Host "Output Files:" -ForegroundColor White
Write-Host "  Server:  $outputPath ($([math]::Round($serverSize, 2)) MB)" -ForegroundColor Cyan
Write-Host "  WASM:    $wasmOutput ($([math]::Round($wasmSize, 2)) MB)" -ForegroundColor Cyan
Write-Host "  Runtime: bin\wasm_exec.js" -ForegroundColor Cyan
Write-Host ""
Write-Host "✓ Production build complete!" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Cyan
