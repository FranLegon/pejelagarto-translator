#!/bin/bash
# build-obfuscated.sh
# Build script for creating obfuscated binary on Linux/macOS

set -e

# Parse command line arguments
OS="${1:-linux}"

# Determine output filename based on OS
if [ "$OS" = "windows" ]; then
    OUTPUT_FILE="bin/piper-server.exe"
else
    OUTPUT_FILE="bin/piper-server"
fi

echo "Building obfuscated version for $OS..."
echo "Output: $OUTPUT_FILE"

# Run garble with obfuscation flags
garble -literals -tiny build -tags obfuscated -o "$OUTPUT_FILE" main.go

if [ $? -eq 0 ]; then
    echo "Build successful: $OUTPUT_FILE"
    
    # Make executable on Unix systems
    if [ "$OS" != "windows" ]; then
        chmod +x "$OUTPUT_FILE"
        echo "Made executable: $OUTPUT_FILE"
    fi
else
    echo "Build failed with exit code $?"
    exit 1
fi
