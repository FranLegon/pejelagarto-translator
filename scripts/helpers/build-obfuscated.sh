#!/bin/bash
# build-obfuscated.sh
# Build script for creating obfuscated binary with embedded binaries on Linux/macOS

set -e

# Parse command line arguments
OS="${1:-linux}"

# Determine output filename based on OS
if [ "$OS" = "windows" ]; then
    OUTPUT_FILE="bin/piper-server.exe"
else
    OUTPUT_FILE="bin/piper-server"
fi

echo "🔨 Building Obfuscated Version with Embedded Binaries"
echo "Building obfuscated version for $OS..."
echo "Output: $OUTPUT_FILE"

# Ensure bin directory exists
mkdir -p bin

# Build required embedded binaries first
echo ""
echo "📦 Building embedded binaries..."

# Build Windows binary
echo "  Building Windows binary..."
GOOS=windows GOARCH=amd64 go build -o bin/pejelagarto-translator.exe .
if [ $? -ne 0 ]; then
    echo "❌ Windows binary build failed!"
    exit 1
fi

# Build Linux binary
echo "  Building Linux binary..."
GOOS=linux GOARCH=amd64 go build -o bin/pejelagarto-translator .
if [ $? -ne 0 ]; then
    echo "❌ Linux binary build failed!"
    exit 1
fi

# Build Android APK if possible
echo "  Building Android APK..."
if [ -f "./scripts/helpers/build-android-apk.sh" ]; then
    ./scripts/helpers/build-android-apk.sh >/dev/null 2>&1 || echo "⚠️  Android APK build failed, continuing..."
fi

# Build Android WebView APK if possible  
echo "  Building Android WebView APK..."
if [ -f "./scripts/helpers/build-android-webview.sh" ]; then
    ./scripts/helpers/build-android-webview.sh >/dev/null 2>&1 || echo "⚠️  Android WebView APK build failed, continuing..."
fi

echo ""
echo "📦 Building obfuscated server with embedded binaries..."

# Run garble with obfuscation flags
garble -literals -tiny build -tags "obfuscated,downloadable" -o "$OUTPUT_FILE" main.go

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Obfuscated build complete!"
    echo "📁 Output: $OUTPUT_FILE"
    echo ""
    echo "Features included:"
    echo "  ✓ Code obfuscation (garble)"
    echo "  ✓ Embedded Windows/Linux binaries"
    echo "  ✓ Embedded Android APKs (if built)"
    echo "  ✓ Download buttons in web UI"
    
    # Make executable on Unix systems
    if [ "$OS" != "windows" ]; then
        chmod +x "$OUTPUT_FILE"
        echo "  ✓ Made executable: $OUTPUT_FILE"
    fi
else
    echo "❌ Build failed with exit code $?"
    exit 1
fi
