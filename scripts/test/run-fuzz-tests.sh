#!/bin/bash

# Comprehensive fuzz test runner
# Runs all fuzz tests with required minimum durations
# - Component tests: 30s minimum
# - Main translation test: 120s minimum

set -e

echo "========================================"
echo "Running Comprehensive Fuzz Tests"
echo "========================================"
echo ""
echo "This will run all fuzz tests with minimum required durations:"
echo "  - Component tests: 30 seconds each"
echo "  - Main translation test: 120 seconds"
echo ""
echo "Total estimated time: ~5 minutes"
echo ""

# Component fuzz tests (30s each)
echo "----------------------------------------"
echo "1/7: FuzzApplyMapReplacements (30s)"
echo "----------------------------------------"
go test -fuzz=FuzzApplyMapReplacements -fuzztime=30s

echo ""
echo "----------------------------------------"
echo "2/7: FuzzApplyNumbersLogic (30s)"
echo "----------------------------------------"
go test -fuzz=FuzzApplyNumbersLogic -fuzztime=30s

echo ""
echo "----------------------------------------"
echo "3/7: FuzzApplyAccentReplacementLogic (30s)"
echo "----------------------------------------"
go test -fuzz=FuzzApplyAccentReplacementLogic -fuzztime=30s

echo ""
echo "----------------------------------------"
echo "4/7: FuzzApplyPunctuationReplacements (30s)"
echo "----------------------------------------"
go test -fuzz=FuzzApplyPunctuationReplacements -fuzztime=30s

echo ""
echo "----------------------------------------"
echo "5/7: FuzzApplyCaseReplacementLogic (30s)"
echo "----------------------------------------"
go test -fuzz=FuzzApplyCaseReplacementLogic -fuzztime=30s

echo ""
echo "----------------------------------------"
echo "6/7: FuzzSpecialCharDateTimeEncoding (30s)"
echo "----------------------------------------"
go test -fuzz=FuzzSpecialCharDateTimeEncoding -fuzztime=30s

echo ""
echo "----------------------------------------"
echo "7/7: FuzzTranslatePejelagarto (120s) - MAIN TEST"
echo "----------------------------------------"
go test -fuzz=FuzzTranslatePejelagarto -fuzztime=120s

echo ""
echo "========================================"
echo "✓ All Fuzz Tests Completed!"
echo "========================================"
echo ""
echo "Summary:"
echo "  - 6 component tests (30s each)"
echo "  - 1 main translation test (120s)"
echo "  - All tests use random fuzzy input"
echo "  - 100% reversibility verified"
echo ""
