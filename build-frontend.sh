#!/bin/bash

# Build script for frontend mode (WASM)
# This script compiles the Go translation logic to WebAssembly

set -e

echo "Building Pejelagarto Translator in frontend mode..."
echo "This will compile Go code to WebAssembly for client-side translation."
echo ""

# Build the WASM module
echo "Compiling Go to WASM..."
GOOS=js GOARCH=wasm go build -tags frontend -o bin/translator.wasm

# Copy wasm_exec.js from Go installation
echo "Copying wasm_exec.js from Go installation..."
GOROOT=$(go env GOROOT)
# Try both possible locations
if [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/misc/wasm/wasm_exec.js" bin/
elif [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
    cp "$GOROOT/lib/wasm/wasm_exec.js" bin/
else
    echo "Warning: Could not find wasm_exec.js in Go installation"
    echo "You may need to download it manually from: https://github.com/golang/go/blob/master/misc/wasm/wasm_exec.js"
fi

echo ""
echo "✓ Frontend build complete!"
echo "  - WASM module: bin/translator.wasm"
echo "  - WASM loader: bin/wasm_exec.js"
echo ""
echo "To run the frontend server:"
echo "  go run server_frontend.go"
echo ""
echo "The server will serve:"
echo "  - HTML UI with WASM loader"
echo "  - Translation: Client-side (WASM)"
echo "  - TTS audio: Server-side (normal endpoints)"
