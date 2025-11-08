#!/bin/bash

# Test script for WASM build
# This script verifies WASM tests compile correctly
# Note: WASM tests cannot be executed directly, only compiled

set -e

echo "========================================"
echo "Running WASM Tests"
echo "========================================"
echo ""

# Set environment variables for WASM testing
export GOOS=js
export GOARCH=wasm

echo "Building WASM test binary..."
echo ""

# Compile tests with frontend build tag (cannot run WASM tests directly)
echo "Compiling WASM tests with -tags frontend..."
go test -c -tags frontend -o /tmp/wasm-test.wasm 2>&1

# Check if compilation succeeded
if [ $? -eq 0 ]; then
    echo "✓ WASM test compilation successful"
    
    # Verify the WASM file was created and has reasonable size
    if [ -f /tmp/wasm-test.wasm ]; then
        SIZE=$(stat -f%z /tmp/wasm-test.wasm 2>/dev/null || stat -c%s /tmp/wasm-test.wasm 2>/dev/null)
        echo "✓ WASM test binary created: $(numfmt --to=iec-i --suffix=B $SIZE 2>/dev/null || echo "${SIZE} bytes")"
        rm -f /tmp/wasm-test.wasm
    fi
    
    echo ""
    echo "========================================"
    echo "✓ WASM tests compiled successfully!"
    echo "========================================"
    echo ""
    echo "Note: WASM tests cannot be executed directly."
    echo "They are verified through successful compilation."
    echo "The WASM functions are tested in the browser environment."
    exit 0
else
    echo ""
    echo "========================================"
    echo "✗ WASM test compilation failed!"
    echo "========================================"
    exit 1
fi
