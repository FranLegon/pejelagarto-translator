#!/usr/bin/env pwsh
# Test script to isolate garble vs Windows Defender ngrok issues
# This script systematically tests different build configurations
# Simplified version without background jobs to avoid language mode issues

param(
    [Parameter(Mandatory=$false)]
    [switch]$SkipDefenderCheck,
    
    [Parameter(Mandatory=$false)]
    [switch]$AddDefenderExclusions,
    
    [Parameter(Mandatory=$false)]
    [int]$TestDurationSeconds = 8
)

$ErrorActionPreference = "Stop"

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Garble + ngrok Compatibility Test" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Check if running as administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $isAdmin -and $AddDefenderExclusions) {
    Write-Host "ERROR: Adding Defender exclusions requires Administrator privileges" -ForegroundColor Red
    Write-Host "Please run this script as Administrator with -AddDefenderExclusions flag" -ForegroundColor Yellow
    exit 1
}

# Test results
$results = @()

# Function to test a binary (simplified without background jobs)
function Test-NgrokBinary {
    param(
        [string]$BinaryPath,
        [string]$TestName,
        [int]$TimeoutSeconds = 8
    )
    
    Write-Host "`n[Testing: $TestName]" -ForegroundColor Yellow
    Write-Host "Binary: $BinaryPath" -ForegroundColor White
    
    if (-not (Test-Path $BinaryPath)) {
        Write-Host "❌ Binary not found!" -ForegroundColor Red
        return @{
            Name = $TestName
            Path = $BinaryPath
            Success = $false
            Error = "Binary not found"
        }
    }
    
    # Check if Windows Defender would block it
    if (-not $SkipDefenderCheck) {
        Write-Host "Checking Windows Defender status..." -ForegroundColor Cyan
        try {
            $defenderStatus = Get-MpComputerStatus -ErrorAction SilentlyContinue
            if ($defenderStatus.RealTimeProtectionEnabled) {
                Write-Host "⚠️  Windows Defender Real-Time Protection is ENABLED" -ForegroundColor Yellow
            } else {
                Write-Host "✓ Windows Defender Real-Time Protection is DISABLED" -ForegroundColor Green
            }
        } catch {
            Write-Host "⚠️  Could not check Defender status (not Windows or no permissions)" -ForegroundColor Yellow
        }
    }
    
    # Start the binary and capture output to a temp file
    Write-Host "Starting binary (will run for ${TimeoutSeconds}s)..." -ForegroundColor Cyan
    $tempOutput = [System.IO.Path]::GetTempFileName()
    
    try {
        # Start process with output redirection
        $psi = New-Object System.Diagnostics.ProcessStartInfo
        $psi.FileName = $BinaryPath
        $psi.RedirectStandardOutput = $true
        $psi.RedirectStandardError = $true
        $psi.UseShellExecute = $false
        $psi.CreateNoWindow = $true
        
        $process = New-Object System.Diagnostics.Process
        $process.StartInfo = $psi
        
        # Capture output
        $outputBuilder = New-Object System.Text.StringBuilder
        $errorBuilder = New-Object System.Text.StringBuilder
        
        $outputHandler = {
            if ($EventArgs.Data) {
                [void]$Event.MessageData.AppendLine($EventArgs.Data)
                Write-Host "  → $($EventArgs.Data)" -ForegroundColor Gray
            }
        }
        
        $outputEvent = Register-ObjectEvent -InputObject $process -EventName OutputDataReceived -Action $outputHandler -MessageData $outputBuilder
        $errorEvent = Register-ObjectEvent -InputObject $process -EventName ErrorDataReceived -Action $outputHandler -MessageData $errorBuilder
        
        $started = $process.Start()
        if (-not $started) {
            throw "Failed to start process"
        }
        
        $process.BeginOutputReadLine()
        $process.BeginErrorReadLine()
        
        # Wait for timeout
        $exited = $process.WaitForExit($TimeoutSeconds * 1000)
        
        # Get all output
        Start-Sleep -Milliseconds 500  # Allow event handlers to finish
        Unregister-Event -SourceIdentifier $outputEvent.Name -ErrorAction SilentlyContinue
        Unregister-Event -SourceIdentifier $errorEvent.Name -ErrorAction SilentlyContinue
        
        $outputStr = $outputBuilder.ToString() + $errorBuilder.ToString()
        
        if ($exited) {
            # Process exited before timeout
            $exitCode = $process.ExitCode
            Write-Host ""
            Write-Host "Process exited with code: $exitCode" -ForegroundColor $(if ($exitCode -eq 0) { "Green" } else { "Red" })
            
            # Check for ngrok error in output
            if ($outputStr -match "Failed to start ngrok listener|remote gone away|ERR_NGROK") {
                Write-Host "❌ FAILED - ngrok connection error detected!" -ForegroundColor Red
                return @{
                    Name = $TestName
                    Path = $BinaryPath
                    Success = $false
                    Error = "ngrok connection failed"
                    Output = $outputStr
                }
            } elseif ($outputStr -match "ngrok tunnel established|Public URL") {
                Write-Host "✓ SUCCESS - ngrok tunnel established!" -ForegroundColor Green
                return @{
                    Name = $TestName
                    Path = $BinaryPath
                    Success = $true
                    Error = $null
                    Output = $outputStr
                }
            } else {
                Write-Host "⚠️  Process exited but no clear ngrok status" -ForegroundColor Yellow
                return @{
                    Name = $TestName
                    Path = $BinaryPath
                    Success = $false
                    Error = "Unexpected exit"
                    Output = $outputStr
                }
            }
        } else {
            # Still running after timeout - likely working
            Write-Host ""
            Write-Host "Process still running after ${TimeoutSeconds}s" -ForegroundColor Cyan
            
            # Check output for success/failure
            if ($outputStr -match "Failed to start ngrok listener|remote gone away|ERR_NGROK") {
                Write-Host "❌ FAILED - ngrok connection error detected!" -ForegroundColor Red
                $process.Kill()
                return @{
                    Name = $TestName
                    Path = $BinaryPath
                    Success = $false
                    Error = "ngrok connection failed"
                    Output = $outputStr
                }
            } elseif ($outputStr -match "ngrok tunnel established|Public URL|Server is running") {
                Write-Host "✓ SUCCESS - Server appears to be running normally!" -ForegroundColor Green
                $process.Kill()
                return @{
                    Name = $TestName
                    Path = $BinaryPath
                    Success = $true
                    Error = $null
                    Output = $outputStr
                }
            } else {
                Write-Host "✓ LIKELY SUCCESS - No errors detected, server running" -ForegroundColor Green
                $process.Kill()
                return @{
                    Name = $TestName
                    Path = $BinaryPath
                    Success = $true
                    Error = $null
                    Output = $outputStr
                }
            }
        }
    } catch {
        Write-Host "❌ Exception: $_" -ForegroundColor Red
        return @{
            Name = $TestName
            Path = $BinaryPath
            Success = $false
            Error = $_.Exception.Message
        }
    } finally {
        # Cleanup
        if ($process -and -not $process.HasExited) {
            $process.Kill()
        }
        if ($process) {
            $process.Dispose()
        }
        if (Test-Path $tempOutput) {
            Remove-Item $tempOutput -Force -ErrorAction SilentlyContinue
        }
    }
}

# Add Windows Defender exclusions if requested
if ($AddDefenderExclusions) {
    Write-Host "`n[Step 0: Adding Windows Defender Exclusions]" -ForegroundColor Green
    Write-Host ""
    
    $exclusionPaths = @(
        "$env:LOCALAPPDATA\Temp",
        "$PSScriptRoot\..\..\bin"
    )
    
    foreach ($path in $exclusionPaths) {
        Write-Host "Adding exclusion: $path" -ForegroundColor White
        try {
            Add-MpPreference -ExclusionPath $path -ErrorAction Stop
            Write-Host "✓ Added exclusion" -ForegroundColor Green
        } catch {
            Write-Host "⚠️  Could not add exclusion: $_" -ForegroundColor Yellow
        }
    }
    
    Write-Host ""
    Write-Host "Waiting 5 seconds for exclusions to take effect..." -ForegroundColor Cyan
    Start-Sleep -Seconds 5
}

# Test 1: Unobfuscated build (baseline - should work)
Write-Host "`n[Test 1: Unobfuscated Build (Baseline)]" -ForegroundColor Green
Write-Host "Building with build-prod-unobfuscated.ps1..." -ForegroundColor Cyan
& "$PSScriptRoot\..\helpers\build-prod-unobfuscated.ps1"
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
$results += Test-NgrokBinary -BinaryPath "bin\piper-server.exe" -TestName "Unobfuscated (Baseline)"

# Test 2: Garble-obfuscated build
Write-Host "`n`n[Test 2: Garble-Obfuscated Build]" -ForegroundColor Green
Write-Host "Building with build-prod.ps1..." -ForegroundColor Cyan
& "$PSScriptRoot\..\helpers\build-prod.ps1"
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
# Rename to avoid overwriting
Move-Item "bin\piper-server.exe" "bin\piper-server-garbled.exe" -Force
$results += Test-NgrokBinary -BinaryPath "bin\piper-server-garbled.exe" -TestName "Garble-Obfuscated"

# Test 3: Garble without -literals flag (if literals break strings)
Write-Host "`n`n[Test 3: Garble WITHOUT -literals Flag]" -ForegroundColor Green
Write-Host "Testing if -literals flag breaks ngrok..." -ForegroundColor Cyan

# Build WASM first
Write-Host "Building WASM..." -ForegroundColor White
& {
    $env:GOOS = "js"
    $env:GOARCH = "wasm"
    garble -tiny -seed=random build -tags "frontend" -o "bin/main.wasm" .
}

if ($LASTEXITCODE -ne 0) {
    Write-Host "WASM build failed!" -ForegroundColor Red
} else {
    # Build server without -literals
    Write-Host "Building server without -literals..." -ForegroundColor White
    & {
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        $env:CGO_ENABLED = "0"
        
        garble -tiny -seed=random build `
            -tags "frontendserver,obfuscated,ngrok_default,downloadable" `
            -ldflags="-s -w -extldflags '-static'" `
            -trimpath `
            -o "bin\piper-server-no-literals.exe" `
            .
    }
    
    if ($LASTEXITCODE -eq 0) {
        $results += Test-NgrokBinary -BinaryPath "bin\piper-server-no-literals.exe" -TestName "Garble (no -literals)"
    } else {
        Write-Host "Build failed!" -ForegroundColor Red
    }
}

# Print summary
Write-Host "`n`n======================================" -ForegroundColor Cyan
Write-Host "  Test Results Summary" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan

foreach ($result in $results) {
    $status = if ($result.Success) { "✓ PASS" } else { "❌ FAIL" }
    $color = if ($result.Success) { "Green" } else { "Red" }
    
    Write-Host ""
    Write-Host "$status - $($result.Name)" -ForegroundColor $color
    if (-not $result.Success) {
        Write-Host "  Error: $($result.Error)" -ForegroundColor Yellow
    }
}

# Conclusion
Write-Host "`n`n======================================" -ForegroundColor Cyan
Write-Host "  Analysis" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

$unobfuscatedSuccess = ($results | Where-Object { $_.Name -eq "Unobfuscated (Baseline)" }).Success
$garbledSuccess = ($results | Where-Object { $_.Name -eq "Garble-Obfuscated" }).Success
$noLiteralsSuccess = ($results | Where-Object { $_.Name -eq "Garble (no -literals)" }).Success

if ($unobfuscatedSuccess -and -not $garbledSuccess) {
    Write-Host "CONCLUSION: Garble obfuscation breaks ngrok" -ForegroundColor Red
    Write-Host ""
    Write-Host "Evidence:" -ForegroundColor Yellow
    Write-Host "  - Unobfuscated build: WORKS ✓" -ForegroundColor Green
    Write-Host "  - Garble build: FAILS ❌" -ForegroundColor Red
    Write-Host ""
    
    if ($noLiteralsSuccess) {
        Write-Host "Root Cause: The -literals flag breaks ngrok string constants" -ForegroundColor Yellow
        Write-Host "Solution: Remove -literals flag from garble command" -ForegroundColor Cyan
    } else {
        Write-Host "Root Cause: Garble's core obfuscation (not just -literals) breaks ngrok" -ForegroundColor Yellow
        Write-Host "Solution: Use unobfuscated build for ngrok deployments" -ForegroundColor Cyan
    }
} elseif (-not $unobfuscatedSuccess) {
    Write-Host "CONCLUSION: Problem is NOT garble-related" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Evidence:" -ForegroundColor Yellow
    Write-Host "  - Even unobfuscated build fails" -ForegroundColor Red
    Write-Host ""
    Write-Host "Possible causes:" -ForegroundColor Cyan
    Write-Host "  - Network connectivity issues" -ForegroundColor White
    Write-Host "  - ngrok service unavailable" -ForegroundColor White
    Write-Host "  - Invalid ngrok credentials" -ForegroundColor White
    Write-Host "  - Domain configuration issues" -ForegroundColor White
} else {
    Write-Host "UNEXPECTED: All builds working!" -ForegroundColor Green
    Write-Host "The issue may be intermittent or environment-specific." -ForegroundColor Yellow
}

Write-Host ""
