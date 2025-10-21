# get-requirements.ps1
# Downloads all Piper TTS requirements if they're not already present
# Run this script before building to ensure all dependencies are embedded

$ErrorActionPreference = "Stop"

Write-Host "=== Pejelagarto Translator - Dependency Checker ===" -ForegroundColor Cyan
Write-Host ""

# Determine the requirements directory
$RequirementsDir = Join-Path $PSScriptRoot "tts\requirements"
$PiperDir = Join-Path $RequirementsDir "piper"
$LanguagesDir = Join-Path $PiperDir "languages"

# Create directories if they don't exist
if (-not (Test-Path $RequirementsDir)) {
    Write-Host "Creating requirements directory..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $RequirementsDir -Force | Out-Null
}

if (-not (Test-Path $PiperDir)) {
    Write-Host "Creating piper directory..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $PiperDir -Force | Out-Null
}

if (-not (Test-Path $LanguagesDir)) {
    Write-Host "Creating languages directory..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $LanguagesDir -Force | Out-Null
}

# Function to download a file
function Download-File {
    param (
        [string]$Url,
        [string]$OutputPath
    )
    
    Write-Host "  Downloading from: $Url" -ForegroundColor Gray
    Write-Host "  Saving to: $OutputPath" -ForegroundColor Gray
    
    try {
        Invoke-WebRequest -Uri $Url -OutFile $OutputPath -UseBasicParsing
        Write-Host "  ✓ Downloaded successfully" -ForegroundColor Green
        return $true
    } catch {
        Write-Host "  ✗ Failed to download: $_" -ForegroundColor Red
        return $false
    }
}

# Check for Piper binary and DLLs
Write-Host "Checking Piper binary..." -ForegroundColor Cyan
$PiperExe = Join-Path $RequirementsDir "piper.exe"

if (-not (Test-Path $PiperExe)) {
    Write-Host "Piper binary not found. Downloading..." -ForegroundColor Yellow
    
    $ZipPath = Join-Path $RequirementsDir "piper_windows_amd64.zip"
    $Url = "https://github.com/rhasspy/piper/releases/latest/download/piper_windows_amd64.zip"
    
    if (Download-File -Url $Url -OutputPath $ZipPath) {
        Write-Host "Extracting Piper..." -ForegroundColor Yellow
        
        try {
            # Extract to a temporary directory
            $TempExtractDir = Join-Path $RequirementsDir "temp_extract"
            if (Test-Path $TempExtractDir) {
                Remove-Item -Path $TempExtractDir -Recurse -Force
            }
            
            Expand-Archive -Path $ZipPath -DestinationPath $TempExtractDir -Force
            
            # Copy all files from the extracted directory
            $ExtractedFiles = Get-ChildItem -Path $TempExtractDir -Recurse -File
            foreach ($File in $ExtractedFiles) {
                $DestPath = Join-Path $RequirementsDir $File.Name
                Copy-Item -Path $File.FullName -Destination $DestPath -Force
                Write-Host "  ✓ Copied $($File.Name)" -ForegroundColor Green
            }
            
            # Copy espeak-ng-data directory if it exists
            $EspeakSource = Join-Path $TempExtractDir "piper\espeak-ng-data"
            $EspeakDest = Join-Path $RequirementsDir "espeak-ng-data"
            if (Test-Path $EspeakSource) {
                if (Test-Path $EspeakDest) {
                    Remove-Item -Path $EspeakDest -Recurse -Force
                }
                Copy-Item -Path $EspeakSource -Destination $EspeakDest -Recurse -Force
                Write-Host "  ✓ Copied espeak-ng-data directory" -ForegroundColor Green
            }
            
            # Clean up
            Remove-Item -Path $TempExtractDir -Recurse -Force
            Remove-Item -Path $ZipPath -Force
            
            Write-Host "✓ Piper binary and dependencies installed" -ForegroundColor Green
        } catch {
            Write-Host "✗ Failed to extract: $_" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "✗ Failed to download Piper binary" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "✓ Piper binary found" -ForegroundColor Green
}

# Check for espeak-ng-data
Write-Host "`nChecking espeak-ng-data..." -ForegroundColor Cyan
$EspeakData = Join-Path $RequirementsDir "espeak-ng-data"

if (-not (Test-Path $EspeakData)) {
    Write-Host "✗ espeak-ng-data not found. This should have been included with Piper." -ForegroundColor Red
    Write-Host "  Please manually download from: https://github.com/rhasspy/piper/releases/latest" -ForegroundColor Yellow
    exit 1
} else {
    Write-Host "✓ espeak-ng-data found" -ForegroundColor Green
}

# Check for language models
Write-Host "`nChecking language models..." -ForegroundColor Cyan

$Languages = @{
    "portuguese" = @{
        "voice" = "pt_BR-faber-medium"
        "url_base" = "https://huggingface.co/rhasspy/piper-voices/resolve/main/pt/pt_BR/faber/medium"
    }
    "spanish" = @{
        "voice" = "es_ES-davefx-medium"
        "url_base" = "https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/davefx/medium"
    }
    "english" = @{
        "voice" = "en_US-lessac-medium"
        "url_base" = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium"
    }
    "russian" = @{
        "voice" = "ru_RU-irina-medium"
        "url_base" = "https://huggingface.co/rhasspy/piper-voices/resolve/main/ru/ru_RU/irina/medium"
    }
}

foreach ($LangName in $Languages.Keys) {
    $LangInfo = $Languages[$LangName]
    $LangDir = Join-Path $LanguagesDir $LangName
    
    Write-Host "`n  Checking $LangName ($($LangInfo.voice))..." -ForegroundColor Yellow
    
    # Create language directory if it doesn't exist
    if (-not (Test-Path $LangDir)) {
        New-Item -ItemType Directory -Path $LangDir -Force | Out-Null
    }
    
    # Check for model.onnx
    $ModelFile = Join-Path $LangDir "model.onnx"
    $ModelJsonFile = Join-Path $LangDir "model.onnx.json"
    
    $NeedsDownload = $false
    
    if (-not (Test-Path $ModelFile)) {
        Write-Host "    model.onnx not found" -ForegroundColor Yellow
        $NeedsDownload = $true
    }
    
    if (-not (Test-Path $ModelJsonFile)) {
        Write-Host "    model.onnx.json not found" -ForegroundColor Yellow
        $NeedsDownload = $true
    }
    
    if ($NeedsDownload) {
        Write-Host "    Downloading $LangName model..." -ForegroundColor Yellow
        
        # Download model.onnx
        $ModelUrl = "$($LangInfo.url_base)/$($LangInfo.voice).onnx"
        if (Download-File -Url $ModelUrl -OutputPath $ModelFile) {
            Write-Host "    ✓ Downloaded model.onnx (~63 MB)" -ForegroundColor Green
        } else {
            Write-Host "    ✗ Failed to download model.onnx" -ForegroundColor Red
            continue
        }
        
        # Download model.onnx.json
        $ModelJsonUrl = "$($LangInfo.url_base)/$($LangInfo.voice).onnx.json"
        if (Download-File -Url $ModelJsonUrl -OutputPath $ModelJsonFile) {
            Write-Host "    ✓ Downloaded model.onnx.json" -ForegroundColor Green
        } else {
            Write-Host "    ✗ Failed to download model.onnx.json" -ForegroundColor Red
            continue
        }
    } else {
        Write-Host "    ✓ Model files already present" -ForegroundColor Green
    }
}

# Final summary
Write-Host "`n=== Dependency Check Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "All dependencies are ready!" -ForegroundColor Green
Write-Host ""
Write-Host "You can now build the executable with:" -ForegroundColor Yellow
Write-Host "  go build -o pejelagarto-translator.exe main.go" -ForegroundColor White
Write-Host ""
Write-Host "The compiled binary will include all embedded dependencies." -ForegroundColor Gray
Write-Host ""
