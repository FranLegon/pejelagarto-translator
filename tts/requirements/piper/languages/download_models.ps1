# Download All TTS Language Models
# This script downloads ONNX models and configuration files for all supported languages

# Set error action preference
$ErrorActionPreference = "Stop"

# Get the script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$languagesDir = $scriptDir

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Piper TTS Multi-Language Model Downloader" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Function to download a model
function Download-Model {
    param(
        [string]$Language,
        [string]$ModelName,
        [string]$OnnxUrl,
        [string]$JsonUrl,
        [string]$OutputDir
    )
    
    Write-Host "Downloading $Language model ($ModelName)..." -ForegroundColor Green
    
    # Create output directory if it doesn't exist
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    }
    
    $onnxPath = Join-Path $OutputDir "model.onnx"
    $jsonPath = Join-Path $OutputDir "model.onnx.json"
    
    # Check if files already exist
    if ((Test-Path $onnxPath) -and (Test-Path $jsonPath)) {
        Write-Host "  ✓ Model already exists, skipping..." -ForegroundColor Yellow
        return
    }
    
    try {
        # Download ONNX model
        Write-Host "  Downloading ONNX file..." -ForegroundColor Gray
        Invoke-WebRequest -Uri $OnnxUrl -OutFile $onnxPath -UseBasicParsing
        
        # Download JSON config
        Write-Host "  Downloading JSON config..." -ForegroundColor Gray
        Invoke-WebRequest -Uri $JsonUrl -OutFile $jsonPath -UseBasicParsing
        
        # Verify file sizes
        $onnxSize = (Get-Item $onnxPath).Length / 1MB
        $jsonSize = (Get-Item $jsonPath).Length / 1KB
        
        Write-Host "  ✓ Download complete! (ONNX: $([math]::Round($onnxSize, 2)) MB, JSON: $([math]::Round($jsonSize, 2)) KB)" -ForegroundColor Green
    }
    catch {
        Write-Host "  ✗ Error downloading $Language model: $_" -ForegroundColor Red
        # Clean up partial downloads
        if (Test-Path $onnxPath) { Remove-Item $onnxPath -Force }
        if (Test-Path $jsonPath) { Remove-Item $jsonPath -Force }
        throw
    }
    
    Write-Host ""
}

# Download Portuguese model
try {
    Download-Model `
        -Language "Portuguese (Brazilian)" `
        -ModelName "pt_BR-faber-medium" `
        -OnnxUrl "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/pt/pt_BR/faber/medium/pt_BR-faber-medium.onnx" `
        -JsonUrl "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/pt/pt_BR/faber/medium/pt_BR-faber-medium.onnx.json" `
        -OutputDir (Join-Path $languagesDir "portuguese")
} catch {
    Write-Host "Failed to download Portuguese model" -ForegroundColor Red
}

# Download Spanish model
try {
    Download-Model `
        -Language "Spanish (Spain)" `
        -ModelName "es_ES-davefx-medium" `
        -OnnxUrl "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx" `
        -JsonUrl "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx.json" `
        -OutputDir (Join-Path $languagesDir "spanish")
} catch {
    Write-Host "Failed to download Spanish model" -ForegroundColor Red
}

# Download English model
try {
    Download-Model `
        -Language "English (US)" `
        -ModelName "en_US-lessac-medium" `
        -OnnxUrl "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/en/en_US/lessac/medium/en_US-lessac-medium.onnx" `
        -JsonUrl "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/en/en_US/lessac/medium/en_US-lessac-medium.onnx.json" `
        -OutputDir (Join-Path $languagesDir "english")
} catch {
    Write-Host "Failed to download English model" -ForegroundColor Red
}

# Download Russian model
try {
    Download-Model `
        -Language "Russian" `
        -ModelName "ru_RU-irina-medium" `
        -OnnxUrl "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/ru/ru_RU/irina/medium/ru_RU-irina-medium.onnx" `
        -JsonUrl "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/ru/ru_RU/irina/medium/ru_RU-irina-medium.onnx.json" `
        -OutputDir (Join-Path $languagesDir "russian")
} catch {
    Write-Host "Failed to download Russian model" -ForegroundColor Red
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Download Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check what was downloaded
$languages = @("portuguese", "spanish", "english", "russian")
$successCount = 0

foreach ($lang in $languages) {
    $langDir = Join-Path $languagesDir $lang
    $onnxPath = Join-Path $langDir "model.onnx"
    $jsonPath = Join-Path $langDir "model.onnx.json"
    
    if ((Test-Path $onnxPath) -and (Test-Path $jsonPath)) {
        $size = (Get-Item $onnxPath).Length / 1MB
        Write-Host "✓ $lang : $([math]::Round($size, 2)) MB" -ForegroundColor Green
        $successCount++
    } else {
        Write-Host "✗ $lang : Missing files" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "$successCount of $($languages.Count) language models downloaded successfully!" -ForegroundColor Cyan

if ($successCount -eq $languages.Count) {
    Write-Host ""
    Write-Host "All models are ready! You can now run:" -ForegroundColor Green
    Write-Host "  .\pejelagarto-translator.exe -pronunciation_language=portuguese" -ForegroundColor Yellow
    Write-Host "  .\pejelagarto-translator.exe -pronunciation_language=spanish" -ForegroundColor Yellow
    Write-Host "  .\pejelagarto-translator.exe -pronunciation_language=english" -ForegroundColor Yellow
    Write-Host "  .\pejelagarto-translator.exe -pronunciation_language=russian" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Press any key to exit..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
