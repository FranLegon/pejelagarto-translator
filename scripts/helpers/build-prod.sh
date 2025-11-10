#!/bin/bash
# Production build script with obfuscated + frontend + ngrok_default tags
# Uses garble for code obfuscation
# Output: Obfuscated server with WASM frontend and hardcoded ngrok credentials

set -e

# Parse arguments
OS="${1:-linux}"
ARCH="${2:-amd64}"

echo "======================================"
echo "  Production Build Script (Garble)"
echo "======================================"
echo ""
echo "Build Configuration:"
echo "  - Tags: obfuscated, frontend, ngrok_default, downloadable"
echo "  - OS: $OS"
echo "  - Architecture: $ARCH"
echo "  - Obfuscation: garble"
echo ""

# Build embedded binaries first (required for downloadable tag)
echo "[1/7] Building embedded binaries for downloads..."
echo "  Building Windows binary for embedding..."
GOOS=windows GOARCH=amd64 go build -o bin/pejelagarto-translator.exe .
if [ $? -ne 0 ]; then
    echo "Failed to build Windows embedded binary"
    exit 1
fi

echo "  Building Linux binary for embedding..."
GOOS=linux GOARCH=amd64 go build -o bin/pejelagarto-translator .
if [ $? -ne 0 ]; then
    echo "Failed to build Linux embedded binary"
    exit 1
fi

echo "  ✓ Embedded binaries ready for downloads"
echo ""

# Check if garble is installed
echo "[2/7] Checking for garble..."
if ! command -v garble &> /dev/null; then
    echo "ERROR: garble not found. Installing..."
    go install mvdan.cc/garble@latest
    if [ $? -ne 0 ]; then
        echo "Failed to install garble"
        exit 1
    fi
    echo "✓ garble installed successfully"
else
    echo "✓ garble found: $(which garble)"
fi
echo ""

# Build WASM first
echo "[3/7] Building WASM module..."
export GOOS=js
export GOARCH=wasm

WASM_OUTPUT="bin/main.wasm"
echo "  Output: $WASM_OUTPUT"

garble -tiny -literals -seed=random build -tags "frontend" -o "$WASM_OUTPUT" .
if [ $? -ne 0 ]; then
    echo "Failed to build WASM"
    exit 1
fi

WASM_SIZE=$(du -h "$WASM_OUTPUT" | cut -f1)
echo "✓ WASM built successfully ($WASM_SIZE)"
echo ""

# Copy wasm_exec.js
echo "[4/7] Copying wasm_exec.js..."
GOROOT=$(go env GOROOT)
WASM_EXEC_SRC="$GOROOT/misc/wasm/wasm_exec.js"
WASM_EXEC_DEST="bin/wasm_exec.js"

if [ -f "$WASM_EXEC_SRC" ]; then
    cp "$WASM_EXEC_SRC" "$WASM_EXEC_DEST"
    echo "✓ wasm_exec.js copied"
else
    echo "WARNING: wasm_exec.js not found at $WASM_EXEC_SRC"
fi
echo ""

# Determine output filename
OUTPUT_NAME="piper-server"
if [ "$OS" = "windows" ]; then
    OUTPUT_NAME="${OUTPUT_NAME}.exe"
fi
OUTPUT_PATH="bin/$OUTPUT_NAME"

# Build server (frontend server for WASM mode)
echo "[5/7] Building obfuscated frontend server..."
export GOOS="$OS"
export GOARCH="$ARCH"
export CGO_ENABLED=0

echo "  Tags: frontendserver,obfuscated,ngrok_default,downloadable"
echo "  Build: Using tags to auto-select files"
echo "  Output: $OUTPUT_PATH"

garble -tiny -literals -seed=random build \
    -tags "frontendserver,obfuscated,ngrok_default,downloadable" \
    -ldflags="-s -w -extldflags '-static'" \
    -trimpath \
    -o "$OUTPUT_PATH" \
    .

if [ $? -ne 0 ]; then
    echo "Failed to build server"
    exit 1
fi

SERVER_SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
echo "✓ Server built successfully ($SERVER_SIZE)"
echo ""

# Generate checksums
echo "[6/7] Generating checksums..."
CHECKSUM_FILE="bin/checksums-prod.txt"
SERVER_HASH=$(sha256sum "$OUTPUT_PATH" | cut -d' ' -f1)
WASM_HASH=$(sha256sum "$WASM_OUTPUT" | cut -d' ' -f1)

cat > "$CHECKSUM_FILE" <<EOF
Production Build Checksums
Generated: $(date '+%Y-%m-%d %H:%M:%S')
Build: $OS/$ARCH with obfuscated+frontend+ngrok_default+downloadable

Server ($OUTPUT_NAME):
  SHA256: $SERVER_HASH
  Size: $SERVER_SIZE

WASM (main.wasm):
  SHA256: $WASM_HASH
  Size: $WASM_SIZE
EOF

echo "✓ Checksums saved to $CHECKSUM_FILE"
echo ""

# Summary
echo "[7/7] Build Summary"
echo "======================================"
echo "Build Type: Production (Obfuscated)"
echo "Features:"
echo "  ✓ Code obfuscation (garble)"
echo "  ✓ WASM frontend"
echo "  ✓ Hardcoded ngrok credentials"
echo "  ✓ Embedded binaries (downloadable)"
echo ""
echo "Output Files:"
echo "  Server:  $OUTPUT_PATH ($SERVER_SIZE)"
echo "  WASM:    $WASM_OUTPUT ($WASM_SIZE)"
echo "  Runtime: bin/wasm_exec.js"
echo ""
echo "✓ Production build complete!"
echo "======================================"
