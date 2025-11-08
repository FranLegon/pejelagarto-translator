#!/bin/bash

# Test all build tag combinations
# Verifies that all combinations of "obfuscated" and "frontend" build tags work correctly
# NOTE: This runs quick verification (seed corpus only), not full fuzz testing

echo "========================================"
echo "Testing All Build Tag Combinations"
echo "========================================"
echo "(Quick verification - seed corpus only)"
echo ""

failed=0
passed=0

# Test 1: Normal build (no tags)
echo "Test 1/4: Normal build (no tags)"
if go build -o /tmp/test-normal > /dev/null 2>&1 && go test > /dev/null 2>&1; then
    echo "  ✓ Build and tests successful"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 2: Obfuscated build
echo "Test 2/4: Obfuscated build"
if go build -tags obfuscated -o /tmp/test-obfuscated > /dev/null 2>&1 && go test -tags obfuscated > /dev/null 2>&1; then
    echo "  ✓ Build and tests successful"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 3: Frontend build (WASM)
echo "Test 3/4: Frontend build (WASM)"
if GOOS=js GOARCH=wasm go build -tags frontend -o /tmp/test-frontend.wasm > /dev/null 2>&1 && \
   GOOS=js GOARCH=wasm go test -tags frontend -c -o /tmp/test-frontend-tests.wasm > /dev/null 2>&1; then
    echo "  ✓ Build and tests compile successfully"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 4: Obfuscated + Frontend build (WASM)
echo "Test 4/4: Obfuscated + Frontend build (WASM)"
if GOOS=js GOARCH=wasm go build -tags "obfuscated,frontend" -o /tmp/test-obf-frontend.wasm > /dev/null 2>&1 && \
   GOOS=js GOARCH=wasm go test -tags "obfuscated,frontend" -c -o /tmp/test-obf-frontend-tests.wasm > /dev/null 2>&1; then
    echo "  ✓ Build and tests compile successfully"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Summary
echo ""
echo "========================================"
echo "Summary: $passed/4 passed"
echo "========================================"

if [ $failed -eq 0 ]; then
    echo "✓ All build tag combinations are compatible!"
    exit 0
else
    echo "✗ Some build tag combinations failed"
    exit 1
fi
