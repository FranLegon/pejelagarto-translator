#!/bin/bash
# Production build script (unobfuscated) with frontend + ngrok_default tags
# Uses standard Go compiler optimization instead of garble
# This version works reliably with ngrok (garble breaks ngrok SDK)
# Output: Optimized server with WASM frontend and hardcoded ngrok credentials

set -e

# Parse arguments
OS="${1:-linux}"
ARCH="${2:-amd64}"

# Validate OS
case "$OS" in
    linux|darwin|windows)
        ;;
    *)
        echo "Invalid OS: $OS"
        echo "Usage: $0 [linux|darwin|windows] [amd64|arm64]"
        exit 1
        ;;
esac

# Validate architecture
case "$ARCH" in
    amd64|arm64)
        ;;
    *)
        echo "Invalid architecture: $ARCH"
        echo "Usage: $0 [linux|darwin|windows] [amd64|arm64]"
        exit 1
        ;;
esac

echo "======================================"
echo "  Production Build (Unobfuscated)"
echo "======================================"
echo ""
echo "Build Configuration:"
echo "  - Tags: obfuscated, frontend, ngrok_default, downloadable"
echo "  - OS: $OS"
echo "  - Architecture: $ARCH"
echo "  - Optimization: Standard Go (-ldflags='-s -w')"
echo "  - ngrok: Compatible ✓"
echo ""

# Build WASM first
echo "[1/5] Building WASM module..."
WASM_OUTPUT="bin/main.wasm"
echo "  Output: $WASM_OUTPUT"

GOOS=js GOARCH=wasm go build -tags "frontend" -o "$WASM_OUTPUT" .

if [ $? -ne 0 ]; then
    echo "Failed to build WASM"
    exit 1
fi

WASM_SIZE=$(du -h "$WASM_OUTPUT" | cut -f1)
echo "✓ WASM built successfully ($WASM_SIZE)"
echo ""

# Copy wasm_exec.js
echo "[2/5] Copying wasm_exec.js..."
GOROOT=$(go env GOROOT)
WASM_EXEC_SRC="$GOROOT/lib/wasm/wasm_exec.js"
WASM_EXEC_DEST="bin/wasm_exec.js"

if [ -f "$WASM_EXEC_SRC" ]; then
    cp "$WASM_EXEC_SRC" "$WASM_EXEC_DEST"
    echo "✓ wasm_exec.js copied"
else
    echo "WARNING: wasm_exec.js not found at $WASM_EXEC_SRC"
fi
echo ""

# Determine output filename
OUTPUT_NAME="pejelagarto-server"
if [ "$OS" = "windows" ]; then
    OUTPUT_NAME="${OUTPUT_NAME}.exe"
fi
OUTPUT_PATH="bin/$OUTPUT_NAME"

# Build server (frontend server for WASM mode)
echo "[3/5] Building optimized frontend server..."
echo "  Tags: frontendserver,obfuscated,ngrok_default,downloadable"
echo "  Flags: -ldflags='-s -w' (strip symbols)"
echo "  Output: $OUTPUT_PATH"

CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build \
    -ldflags="-s -w -extldflags '-static'" \
    -tags "frontendserver,obfuscated,ngrok_default,downloadable" \
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
echo "[4/5] Generating checksums..."
CHECKSUM_FILE="bin/checksums-prod.txt"
SERVER_HASH=$(sha256sum "$OUTPUT_PATH" | awk '{print $1}')
WASM_HASH=$(sha256sum "$WASM_OUTPUT" | awk '{print $1}')

cat > "$CHECKSUM_FILE" <<EOF
Production Build Checksums (Unobfuscated)
Generated: $(date '+%Y-%m-%d %H:%M:%S')
Build: $OS/$ARCH with obfuscated+frontend+ngrok_default+downloadable
Optimization: Standard Go (-ldflags='-s -w')

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
echo "[5/5] Build Summary"
echo "======================================"
echo "Build Type: Production (Unobfuscated)"
echo "Features:"
echo "  ✓ Standard Go optimization (-s -w)"
echo "  ✓ WASM frontend"
echo "  ✓ Hardcoded ngrok credentials"
echo "  ✓ Embedded binaries (downloadable)"
echo "  ✓ ngrok SDK compatible"
echo "  ✓ Windows Defender friendly"
echo ""
echo "Output Files:"
echo "  Server:  $OUTPUT_PATH ($SERVER_SIZE)"
echo "  WASM:    $WASM_OUTPUT ($WASM_SIZE)"
echo "  Runtime: bin/wasm_exec.js"
echo ""
echo "✓ Production build complete!"
echo "======================================"
