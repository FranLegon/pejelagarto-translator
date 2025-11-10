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

# Build Android APK if possible
echo "📦 Building Android APK..."
if [ -f "./scripts/helpers/build-android-apk.sh" ]; then
    if ./scripts/helpers/build-android-apk.sh >/dev/null 2>&1; then
        if [ -f "bin/pejelagarto-translator.apk" ]; then
            echo "✅ Android APK created: bin/pejelagarto-translator.apk"
        fi
    else
        echo "⚠️  Android APK build failed, continuing without it..."
    fi
else
    echo "⚠️  Android APK build script not found, skipping..."
fi

# Build Android WebView APK if possible
echo "📦 Building Android WebView APK..."
if [ -f "./scripts/helpers/build-android-webview.sh" ]; then
    if ./scripts/helpers/build-android-webview.sh >/dev/null 2>&1; then
        if [ -f "bin/pejelagarto-translator-webview.apk" ]; then
            echo "✅ Android WebView APK created: bin/pejelagarto-translator-webview.apk"
        fi
    else
        echo "⚠️  Android WebView APK build failed, continuing without it..."
    fi
else
    echo "⚠️  Android WebView APK build script not found, skipping..."
fi

# Build ngrok_default version with embedded binaries and hardcoded ngrok
echo ""
echo "📦 Building ngrok_default version with embedded binaries..."
go build -tags "ngrok_default,downloadable" -o bin/pejelagarto-translator-ngrok .
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
echo "  • Embedded Windows/Linux binaries and Android APKs"
echo "  • No need to pass ngrok flags"
