# build.ps1
# Automated build script that ensures all dependencies are ready before building

$ErrorActionPreference = "Stop"

Write-Host "=== Pejelagarto Translator - Build Script ===" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check and download requirements
Write-Host "Step 1: Checking and downloading TTS requirements..." -ForegroundColor Yellow
& .\get-requirements.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to prepare requirements" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 2: Building Go executable..." -ForegroundColor Yellow

# Build the executable
go build -o pejelagarto-translator.exe main.go

if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Build failed" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Build successful!" -ForegroundColor Green
Write-Host ""

# Get file size
$FileInfo = Get-Item pejelagarto-translator.exe
$FileSizeMB = [math]::Round($FileInfo.Length / 1MB, 2)

Write-Host "=== Build Complete ===" -ForegroundColor Cyan
Write-Host "Executable: pejelagarto-translator.exe" -ForegroundColor White
Write-Host "File size: $FileSizeMB MB" -ForegroundColor White
Write-Host ""
Write-Host "All dependencies are embedded in the executable!" -ForegroundColor Green
Write-Host ""
Write-Host "To run:" -ForegroundColor Yellow
Write-Host "  .\pejelagarto-translator.exe" -ForegroundColor White
Write-Host ""
Write-Host "With ngrok:" -ForegroundColor Yellow
Write-Host "  .\pejelagarto-translator.exe -ngrok_token YOUR_TOKEN" -ForegroundColor White
Write-Host ""
