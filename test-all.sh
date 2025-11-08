#!/bin/bash

# Comprehensive test script that runs tests for both normal and WASM builds

set -e

echo "========================================"
echo "Running All Tests (Normal + WASM)"
echo "========================================"
echo ""

# Test 1: Normal build tests
echo "----------------------------------------"
echo "1. Running Normal Build Tests"
echo "----------------------------------------"
echo ""

go test -v -count=1 ./...

if [ $? -ne 0 ]; then
    echo ""
    echo "✗ Normal build tests failed!"
    exit 1
fi

echo ""
echo "✓ Normal build tests passed!"
echo ""

# Test 2: WASM build tests
echo "----------------------------------------"
echo "2. Running WASM Build Tests"
echo "----------------------------------------"
echo ""

./test-wasm.sh

if [ $? -ne 0 ]; then
    echo ""
    echo "✗ WASM build tests failed!"
    exit 1
fi

echo ""
echo "✓ WASM build tests passed!"
echo ""

# Test 3: Verify both builds compile
echo "----------------------------------------"
echo "3. Verifying Build Compilation"
echo "----------------------------------------"
echo ""

echo "Building normal binary..."
go build -o bin/pejelagarto-translator-test

if [ $? -ne 0 ]; then
    echo "✗ Normal build compilation failed!"
    exit 1
fi
echo "✓ Normal build compiled successfully"

echo ""
echo "Building WASM module..."
GOOS=js GOARCH=wasm go build -tags frontend -o bin/translator-test.wasm

if [ $? -ne 0 ]; then
    echo "✗ WASM build compilation failed!"
    exit 1
fi
echo "✓ WASM build compiled successfully"

# Cleanup test binaries
rm -f bin/pejelagarto-translator-test bin/translator-test.wasm

echo ""
echo "========================================"
echo "✓ All Tests Passed!"
echo "========================================"
echo ""
echo "Summary:"
echo "  ✓ Normal build: Tests passed, compilation successful"
echo "  ✓ WASM build: Tests passed, compilation successful"
echo ""
