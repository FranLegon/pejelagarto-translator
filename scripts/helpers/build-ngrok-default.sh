#!/bin/bash
# Build script for ngrok_default version with embedded binaries and hardcoded ngrok config
# This script compiles binaries for Windows and Linux/Mac, then builds
# the main application with both downloadable and ngrok_default tags

echo "🔨 Building Pejelagarto Translator - Ngrok Default Version"
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

# Build ngrok_default version with embedded binaries and hardcoded ngrok
echo ""
echo "📦 Building ngrok_default version..."
go build -tags ngrok_default -o bin/pejelagarto-translator-ngrok .
if [ $? -ne 0 ]; then
    echo "❌ Ngrok default build failed!"
    exit 1
fi

echo ""
echo "✅ Ngrok default build complete!"
echo "📁 Output: bin/pejelagarto-translator-ngrok"
echo ""
echo "To run the ngrok default version:"
echo "  ./bin/pejelagarto-translator-ngrok"
echo ""
echo "This version includes:"
echo "  • Hardcoded ngrok token and domain"
echo "  • Download buttons for embedded binaries"
echo "  • No need to pass ngrok flags"
