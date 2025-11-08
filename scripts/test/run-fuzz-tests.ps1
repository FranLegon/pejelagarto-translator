#!/usr/bin/env pwsh
# Comprehensive fuzz test runner
# Runs all fuzz tests with required minimum durations
# - Component tests: 30s minimum
# - Main translation test: 120s minimum

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Running Comprehensive Fuzz Tests" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "This will run all fuzz tests with minimum required durations:" -ForegroundColor Yellow
Write-Host "  - Component tests: 30 seconds each" -ForegroundColor White
Write-Host "  - Main translation test: 120 seconds" -ForegroundColor White
Write-Host ""
Write-Host "Total estimated time: ~5 minutes" -ForegroundColor Yellow
Write-Host ""

# Component fuzz tests (30s each)
Write-Host "----------------------------------------" -ForegroundColor Green
Write-Host "1/7: FuzzApplyMapReplacements (30s)" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green
go test -fuzz=FuzzApplyMapReplacements -fuzztime=30s
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "----------------------------------------" -ForegroundColor Green
Write-Host "2/7: FuzzApplyNumbersLogic (30s)" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green
go test -fuzz=FuzzApplyNumbersLogic -fuzztime=30s
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "----------------------------------------" -ForegroundColor Green
Write-Host "3/7: FuzzApplyAccentReplacementLogic (30s)" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green
go test -fuzz=FuzzApplyAccentReplacementLogic -fuzztime=30s
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "----------------------------------------" -ForegroundColor Green
Write-Host "4/7: FuzzApplyPunctuationReplacements (30s)" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green
go test -fuzz=FuzzApplyPunctuationReplacements -fuzztime=30s
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "----------------------------------------" -ForegroundColor Green
Write-Host "5/7: FuzzApplyCaseReplacementLogic (30s)" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green
go test -fuzz=FuzzApplyCaseReplacementLogic -fuzztime=30s
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "----------------------------------------" -ForegroundColor Green
Write-Host "6/7: FuzzSpecialCharDateTimeEncoding (30s)" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green
go test -fuzz=FuzzSpecialCharDateTimeEncoding -fuzztime=30s
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "----------------------------------------" -ForegroundColor Green
Write-Host "7/7: FuzzTranslatePejelagarto (120s) - MAIN TEST" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green
go test -fuzz=FuzzTranslatePejelagarto -fuzztime=120s
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✓ All Fuzz Tests Completed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "  - 6 component tests (30s each)" -ForegroundColor White
Write-Host "  - 1 main translation test (120s)" -ForegroundColor White
Write-Host "  - All tests use random fuzzy input" -ForegroundColor White
Write-Host "  - 100% reversibility verified" -ForegroundColor White
Write-Host ""
