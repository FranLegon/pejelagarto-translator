#!/bin/bash

# Test all build tag combinations
# Verifies that all combinations of build tags work correctly:
# - downloadable: embeds Windows/Linux binaries
# - ngrok_default: hardcoded ngrok credentials (includes downloadable)
# - obfuscated: code obfuscation
# - frontend: WASM client-side translation
# NOTE: This runs quick verification (seed corpus only), not full fuzz testing

echo "========================================"
echo "Testing All Build Tag Combinations"
echo "========================================"
echo "(Quick verification - seed corpus only)"
echo ""

failed=0
passed=0
total=12

# Test 1: Backend build (no tags)
echo "Test 1/$total: Backend build (no tags)"
if go build -o /tmp/test-backend > /dev/null 2>&1 && go test > /dev/null 2>&1; then
    echo "  ✓ Build and tests successful"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 2: Downloadable build
echo "Test 2/$total: Downloadable build"
if go build -tags downloadable -o /tmp/test-downloadable > /dev/null 2>&1 && go test -tags downloadable > /dev/null 2>&1; then
    echo "  ✓ Build and tests successful"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 3: Ngrok default build (includes downloadable)
echo "Test 3/$total: Ngrok default build"
if go build -tags ngrok_default -o /tmp/test-ngrok > /dev/null 2>&1 && go test -tags ngrok_default > /dev/null 2>&1; then
    echo "  ✓ Build and tests successful"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 4: Obfuscated build
echo "Test 4/$total: Obfuscated build"
if go build -tags obfuscated -o /tmp/test-obfuscated > /dev/null 2>&1 && go test -tags obfuscated > /dev/null 2>&1; then
    echo "  ✓ Build and tests successful"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 5: Frontend build (WASM)
echo "Test 5/$total: Frontend build (WASM)"
if GOOS=js GOARCH=wasm go build -tags frontend -o /tmp/test-frontend.wasm > /dev/null 2>&1 && \
   GOOS=js GOARCH=wasm go test -tags frontend -c -o /tmp/test-frontend-tests.wasm > /dev/null 2>&1; then
    echo "  ✓ Build and tests compile successfully"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 6: Downloadable + Obfuscated
echo "Test 6/$total: Downloadable + Obfuscated build"
if go build -tags "downloadable,obfuscated" -o /tmp/test-down-obf > /dev/null 2>&1 && go test -tags "downloadable,obfuscated" > /dev/null 2>&1; then
    echo "  ✓ Build and tests successful"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 7: Downloadable + Frontend (WASM)
echo "Test 7/$total: Downloadable + Frontend build (WASM)"
if GOOS=js GOARCH=wasm go build -tags "downloadable,frontend" -o /tmp/test-down-front.wasm > /dev/null 2>&1 && \
   GOOS=js GOARCH=wasm go test -tags "downloadable,frontend" -c -o /tmp/test-down-front-tests.wasm > /dev/null 2>&1; then
    echo "  ✓ Build and tests compile successfully"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 8: Ngrok default + Obfuscated
echo "Test 8/$total: Ngrok default + Obfuscated build"
if go build -tags "ngrok_default,obfuscated" -o /tmp/test-ngrok-obf > /dev/null 2>&1 && go test -tags "ngrok_default,obfuscated" > /dev/null 2>&1; then
    echo "  ✓ Build and tests successful"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 9: Ngrok default + Frontend (WASM)
echo "Test 9/$total: Ngrok default + Frontend build (WASM)"
if GOOS=js GOARCH=wasm go build -tags "ngrok_default,frontend" -o /tmp/test-ngrok-front.wasm > /dev/null 2>&1 && \
   GOOS=js GOARCH=wasm go test -tags "ngrok_default,frontend" -c -o /tmp/test-ngrok-front-tests.wasm > /dev/null 2>&1; then
    echo "  ✓ Build and tests compile successfully"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 10: Obfuscated + Frontend (WASM)
echo "Test 10/$total: Obfuscated + Frontend build (WASM)"
if GOOS=js GOARCH=wasm go build -tags "obfuscated,frontend" -o /tmp/test-obf-front.wasm > /dev/null 2>&1 && \
   GOOS=js GOARCH=wasm go test -tags "obfuscated,frontend" -c -o /tmp/test-obf-front-tests.wasm > /dev/null 2>&1; then
    echo "  ✓ Build and tests compile successfully"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 11: Downloadable + Obfuscated + Frontend (WASM)
echo "Test 11/$total: Downloadable + Obfuscated + Frontend build (WASM)"
if GOOS=js GOARCH=wasm go build -tags "downloadable,obfuscated,frontend" -o /tmp/test-all.wasm > /dev/null 2>&1 && \
   GOOS=js GOARCH=wasm go test -tags "downloadable,obfuscated,frontend" -c -o /tmp/test-all-tests.wasm > /dev/null 2>&1; then
    echo "  ✓ Build and tests compile successfully"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Test 12: Ngrok default + Obfuscated + Frontend (WASM)
echo "Test 12/$total: Ngrok default + Obfuscated + Frontend build (WASM)"
if GOOS=js GOARCH=wasm go build -tags "ngrok_default,obfuscated,frontend" -o /tmp/test-ngrok-all.wasm > /dev/null 2>&1 && \
   GOOS=js GOARCH=wasm go test -tags "ngrok_default,obfuscated,frontend" -c -o /tmp/test-ngrok-all-tests.wasm > /dev/null 2>&1; then
    echo "  ✓ Build and tests compile successfully"
    ((passed++))
else
    echo "  ✗ Failed"
    ((failed++))
fi

# Summary
echo ""
echo "========================================"
echo "Summary: $passed/$total passed"
echo "========================================"

if [ $failed -eq 0 ]; then
    echo "✓ All build tag combinations are compatible!"
    exit 0
else
    echo "✗ Some build tag combinations failed"
    exit 1
fi
