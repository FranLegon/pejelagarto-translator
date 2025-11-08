#!/bin/bash
# Build script for downloadable version with embedded binaries
# This script compiles binaries for Windows and Linux/Mac, then builds
# the main application with the downloadable tag to embed them

echo "🔨 Building Pejelagarto Translator - Downloadable Version"
echo ""

# Ensure bin directory exists
mkdir -p bin

# Build Windows binary
echo "📦 Building Windows binary..."
GOOS=windows GOARCH=amd64 go build -o bin/pejelagarto-translator.exe .
if [ $? -ne 0 ]; then
    echo "❌ Windows build failed!"
    exit 1
fi
echo "✅ Windows binary created: bin/pejelagarto-translator.exe"

# Build Linux/Mac binary
echo "📦 Building Linux/Mac binary..."
GOOS=linux GOARCH=amd64 go build -o bin/pejelagarto-translator .
if [ $? -ne 0 ]; then
    echo "❌ Linux/Mac build failed!"
    exit 1
fi
echo "✅ Linux/Mac binary created: bin/pejelagarto-translator"

# Build downloadable version with embedded binaries
echo ""
echo "📦 Building downloadable version with embedded binaries..."
go build -tags downloadable -o bin/pejelagarto-translator-downloadable .
if [ $? -ne 0 ]; then
    echo "❌ Downloadable build failed!"
    exit 1
fi

echo ""
echo "✅ Downloadable build complete!"
echo "📁 Output: bin/pejelagarto-translator-downloadable"
echo ""
echo "To run the downloadable version:"
echo "  ./bin/pejelagarto-translator-downloadable"
echo ""
echo "The download buttons will appear at the bottom of the web UI."
