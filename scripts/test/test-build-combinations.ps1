# Test all build tag combinations
# Verifies that all combinations of build tags work correctly:
# - downloadable: embeds Windows/Linux binaries
# - ngrok_default: hardcoded ngrok credentials (includes downloadable)
# - obfuscated: code obfuscation
# - frontend: WASM client-side translation
# NOTE: This runs quick verification (seed corpus only), not full fuzz testing

Write-Host "========================================"
Write-Host "Testing All Build Tag Combinations"
Write-Host "========================================"
Write-Host "(Quick verification - seed corpus only)"
Write-Host ""

$failed = 0
$passed = 0
$total = 12

function Test-Build {
    param($Tags, $Output, $TestNum, $Description)
    
    Write-Host "Test $TestNum/$total`: $Description"
    
    if ($Tags) {
        go build -tags $Tags -o $Output 2>&1 | Out-Null
        $buildOk = $LASTEXITCODE -eq 0
        go test -tags $Tags 2>&1 | Out-Null
        $testOk = $LASTEXITCODE -eq 0
    } else {
        go build -o $Output 2>&1 | Out-Null
        $buildOk = $LASTEXITCODE -eq 0
        go test 2>&1 | Out-Null
        $testOk = $LASTEXITCODE -eq 0
    }
    
    if ($buildOk -and $testOk) {
        Write-Host "  ✓ Build and tests successful" -ForegroundColor Green
        return $true
    } else {
        Write-Host "  ✗ Failed" -ForegroundColor Red
        return $false
    }
}

function Test-WasmBuild {
    param($Tags, $Output, $TestNum, $Description)
    
    Write-Host "Test $TestNum/$total`: $Description"
    
    $env:GOOS = "js"
    $env:GOARCH = "wasm"
    
    # For WASM builds, build with the frontend tag (includes main.go for translations + wasm_main.go for entry point)
    go build -tags $Tags -o $Output 2>&1 | Out-Null
    $buildOk = $LASTEXITCODE -eq 0
    go test -tags $Tags -c -o ($Output -replace '\.wasm$', '-tests.wasm') 2>&1 | Out-Null
    $testOk = $LASTEXITCODE -eq 0
    
    Remove-Item Env:\GOOS
    Remove-Item Env:\GOARCH
    
    if ($buildOk -and $testOk) {
        Write-Host "  ✓ Build and tests compile successfully" -ForegroundColor Green
        return $true
    } else {
        Write-Host "  ✗ Failed" -ForegroundColor Red
        return $false
    }
}

# Test 1: Backend build
if (Test-Build -Tags $null -Output "bin/test-backend.exe" -TestNum 1 -Description "Backend build (no tags)") { $passed++ } else { $failed++ }

# Test 2: Downloadable build
if (Test-Build -Tags "downloadable" -Output "bin/test-downloadable.exe" -TestNum 2 -Description "Downloadable build") { $passed++ } else { $failed++ }

# Test 3: Ngrok default build
if (Test-Build -Tags "ngrok_default" -Output "bin/test-ngrok.exe" -TestNum 3 -Description "Ngrok default build") { $passed++ } else { $failed++ }

# Test 4: Obfuscated build
if (Test-Build -Tags "obfuscated" -Output "bin/test-obfuscated.exe" -TestNum 4 -Description "Obfuscated build") { $passed++ } else { $failed++ }

# Test 5: Frontend build (WASM)
if (Test-WasmBuild -Tags "frontend" -Output "bin/test-frontend.wasm" -TestNum 5 -Description "Frontend build (WASM)") { $passed++ } else { $failed++ }

# Test 6: Downloadable + Obfuscated
if (Test-Build -Tags "downloadable,obfuscated" -Output "bin/test-down-obf.exe" -TestNum 6 -Description "Downloadable + Obfuscated build") { $passed++ } else { $failed++ }

# Test 7: Downloadable + Frontend (WASM)
if (Test-WasmBuild -Tags "downloadable,frontend" -Output "bin/test-down-front.wasm" -TestNum 7 -Description "Downloadable + Frontend build (WASM)") { $passed++ } else { $failed++ }

# Test 8: Ngrok default + Obfuscated
if (Test-Build -Tags "ngrok_default,obfuscated" -Output "bin/test-ngrok-obf.exe" -TestNum 8 -Description "Ngrok default + Obfuscated build") { $passed++ } else { $failed++ }

# Test 9: Ngrok default + Frontend (WASM)
if (Test-WasmBuild -Tags "ngrok_default,frontend" -Output "bin/test-ngrok-front.wasm" -TestNum 9 -Description "Ngrok default + Frontend build (WASM)") { $passed++ } else { $failed++ }

# Test 10: Obfuscated + Frontend (WASM)
if (Test-WasmBuild -Tags "obfuscated,frontend" -Output "bin/test-obf-front.wasm" -TestNum 10 -Description "Obfuscated + Frontend build (WASM)") { $passed++ } else { $failed++ }

# Test 11: Downloadable + Obfuscated + Frontend (WASM)
if (Test-WasmBuild -Tags "downloadable,obfuscated,frontend" -Output "bin/test-all.wasm" -TestNum 11 -Description "Downloadable + Obfuscated + Frontend build (WASM)") { $passed++ } else { $failed++ }

# Test 12: Ngrok default + Obfuscated + Frontend (WASM)
if (Test-WasmBuild -Tags "ngrok_default,obfuscated,frontend" -Output "bin/test-ngrok-all.wasm" -TestNum 12 -Description "Ngrok default + Obfuscated + Frontend build (WASM)") { $passed++ } else { $failed++ }

# Summary
Write-Host ""
Write-Host "========================================"
Write-Host "Summary: $passed/$total passed"
Write-Host "========================================"

if ($failed -eq 0) {
    Write-Host "✓ All build tag combinations are compatible!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "✗ Some build tag combinations failed" -ForegroundColor Red
    exit 1
}
